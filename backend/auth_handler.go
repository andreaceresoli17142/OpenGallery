package main

import ( // {{{

	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
) // }}}

// oauth {{{
func addState(state string) error {

	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		return AppendError("addState: ", err)
	}

	defer db.Close()

	_, err = db.Exec(`INSERT INTO LoginState VALUES (?, null)`, state)

	if err != nil {
		return AppendError("addState: ", err)
	}

	return nil
}

func findState(state string) (string, error) {

	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		return "false", AppendError("findState: ", err)
	}

	defer db.Close()

	var str string
	err = db.QueryRow(`SELECT idstring FROM LoginState WHERE idstring = ?`, state).Scan(&str)

	if err == sql.ErrNoRows {
		return "", nil
	}

	if err != nil {
		return "", AppendError("findState: ", err)
	}

	return str, nil
}

func remState(state string) error {

	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		return AppendError("remState: ", err)
	}

	defer db.Close()

	_, err = db.Exec("DELETE FROM LoginState WHERE idstring = ?", state)

	if err != nil {
		return AppendError("remState: ", err)
	}

	return nil
}

func oauthGetTokenCouple(w http.ResponseWriter, r *http.Request) {
	fmt.Println("endpoint hit: oauth get token couple")

	headerContentTtype := r.Header.Get("Content-Type")
	if headerContentTtype != "application/json" {
		httpError(&w, 400, "Content Type is not application/json")
		return
	}

	err := r.ParseForm()

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	vars := mux.Vars(r)
	state, ok := validate(vars["state"], "")

	if !ok {
		httpError(&w, 400, "error validating state")
		return
	}

	// get email
	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	defer db.Close()

	var str string
	err = db.QueryRow(`SELECT userEmail FROM LoginState WHERE idstring = ?`, state).Scan(&str)

	if err == sql.ErrNoRows {
		httpError(&w, 500, err)
		return
	}

	if err != nil {
		httpError(&w, 500, err)
		return
	}
	// finished getting email

	usrId, err := getUserId_Email(str)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	act, act_expt, rft, rft_expt, err := generateTokenCouple(usrId)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	err = updateLoginDate(usrId)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	httpSuccessf(&w, 200, `"access_token":"%s", "act_expt": %d, "refresh_token":"%s", "rft_expt":%d`, act, act_expt, rft, rft_expt)
}

func signInOauth(w http.ResponseWriter, r *http.Request) {
	fmt.Println("endpoint hit: oauth sign in")

	headerContentTtype := r.Header.Get("Content-Type")
	if headerContentTtype != "application/json" {
		httpError(&w, 400, "Content Type is not application/json")
		return
	}

	type ReqData struct {
		State    string `json:"state"`
		Password string `json:"password"`
		Username string `json:"username"`
	}

	var re ReqData
	err := httpGetBody(r, &re)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	username, ok := validate(re.Username, validateUser)

	if !ok {
		httpError(&w, 400, vUserErr)
		return
	}

	state, ok := validate(re.State, "")

	if !ok {
		httpError(&w, 400, "error validating state")
		return
	}

	// get email
	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	defer db.Close()

	var str string
	err = db.QueryRow(`SELECT userEmail FROM LoginState WHERE idstring = ?`, state).Scan(&str)

	if err == sql.ErrNoRows {
		httpError(&w, 500, err)
		return
	}

	if err != nil {
		httpError(&w, 500, err)
		return
	}
	// finished getting email

	email, ok := validate(str, "")

	if !ok {
		httpError(&w, 400, vEmailErr)
		return
	}

	password, ok := validate(re.Password, validatePass)

	if !ok {
		httpError(&w, 400, vPassErr)
		return
	}

	resp, err := addUser(username, email, password)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	if !resp {
		httpError(&w, 400, "username or email already in use")
		return
	}

	usrId, err := getUserId_Email(email)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	act, act_expt, rft, rft_expt, err := generateTokenCouple(usrId)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	err = updateLoginDate(usrId)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	httpSuccessf(&w, 200, `"access_token":"%s", "act_expt": %d,  "refresh_token":"%s", "rft_expt":%d`, act, act_expt, rft, rft_expt)
}

func paleoIdAuth(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	state, code := query.Get("state"), query.Get("code")
	fmt.Println("endpoint hit: paleoId Auth")

	ret, err := findState(state)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	if ret == "" {
		return
	}

	payload := strings.NewReader(fmt.Sprintf(`{"grant_type":"%s" , "code":"%s", "redirect_uri":"%s", "client_id":"%s", "client_secret":"%s" }`, "authorization_code", code, redirectUri, clientId, clientSecret))

	req, err := http.NewRequest("POST", "https://id.paleo.bg.it/oauth/token", payload)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	defer resp.Body.Close()

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	// now handle the response
	var respData OauthResp
	err = json.Unmarshal(body, &respData)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	url := "https://id.paleo.bg.it/api/v2/user"
	req, err = http.NewRequest("GET", url, nil)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+respData.AccessToken)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	defer res.Body.Close()

	body, err = ioutil.ReadAll(res.Body)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	var resp1Data UserData
	err = json.Unmarshal(body, &resp1Data)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	email := resp1Data.Email

	// pair state with email

	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	defer db.Close()

	_, err = db.Exec("UPDATE LoginState SET userEmail = ? WHERE idstring = ?", email, state)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	// redirect user

	userId, err := getUserId_Email(email)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	username, _, _, err := getUserData(userId)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	if username == "" {
		http.Redirect(w, r, "http://localhost/signUp/oauth.html", http.StatusMovedPermanently)
		return
	}

	http.Redirect(w, r, "http://localhost/getoauthtk.html", http.StatusMovedPermanently)
}

func getOauthLink(w http.ResponseWriter, r *http.Request) {
	fmt.Println("endpoint hit: getlink")

	var state string

	for {
		state = RandomString(15)
		ret, err := findState(state)

		if err != nil {
			httpError(&w, 500, err)
			return
		}

		if ret == "" {
			break
		}
	}

	err := addState(state)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	httpSuccessf(&w, 200, `"state":"%v", "link":"https://id.paleo.bg.it/oauth/authorize?client_id=%v&response_type=code&state=%v&redirect_uri=%v"`, state, clientId, state, redirectUri)
}

// }}}

// using refresh token {{{
func refreshTokenReq(w http.ResponseWriter, r *http.Request) {
	fmt.Println("endpoint hit: use refresh token")

	// get refresh token from header
	reft := r.Header.Get("refresh-token")

	// use refresh token and return new token couple
	act, a_expt, rft, r_expt, err := useRefreshToken(reft)

	if err != nil {
		httpError(&w, 500, err.Error())
		return
	}

	if act != "" {
		httpSuccessf(&w, 200, `"access_token":"%s", "act_expt": %d,  "refresh_token":"%s", "rft_expt":%d`, act, a_expt, rft, r_expt)
		return
	}

	httpError(&w, 400, "invalid refresh token")
}

func getUserIdFromRefreshToken(refresh_token string) (int, error) {

	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		return -1, AppendError("getUserIdFromRefreshToken: ", err)
	}

	defer db.Close()

	var ret actData
	err = db.QueryRow("SELECT userid, rft_expt FROM Token WHERE refreshToken = (?);", refresh_token).Scan(&ret.User_id, &ret.Rft_expt)

	// return -1 if empty or token is expired
	if err == sql.ErrNoRows && int(time.Now().Unix()) < ret.Rft_expt {
		return -1, nil
	}

	// else return error
	if err != nil {
		return -1, AppendError("getUserIdFromRefreshToken: ", err)
	}

	return ret.User_id, nil
}

//? function may be usless
func useRefreshToken(refresh_token string) (string, int, string, int, error) {
	// get userid from refresh token
	usrId, err := getUserIdFromRefreshToken(refresh_token)

	// return error
	if err != nil {
		return "", -1, "", -1, err
	}

	// return -1 if the token does not exists
	if usrId == -1 {
		return "", -1, "", -1, err
	}

	return generateTokenCouple(usrId)
}

// }}}

// generating tokens {{{
func generateTokenCouple(usrId int) (string, int, string, int, error) {
	// generate random string for access token (check if token already exists)
	act := ""

	for {

		act = RandomString(64)
		ret, err := accessTokenExists(act)

		if err != nil {
			return "", -1, "", -1, AppendError("generateTokenCouple: ", err)
		}

		if !ret {
			break
		}
	}

	// generate random string for access token (check if token already exists)
	rft := ""

	for {

		rft = RandomString(64)
		ret, err := refreshTokenExists(rft)

		if err != nil {
			return "", -1, "", -1, AppendError("generateTokenCouple: ", err)
		}

		if !ret {
			break
		}
	}

	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		return "", -1, "", -1, AppendError("generateTokenCouple: ", err)
	}

	defer db.Close()

	// set access token expire time to 30 min
	act_expt := int(time.Now().Unix()) + 1800
	// set refresh token expire time to 7d
	rft_expt := int(time.Now().Unix()) + 604800

	_, err = db.Exec(`
	INSERT INTO Token VALUES (?, ?, ?, ?, ?)
	ON DUPLICATE KEY
	UPDATE accessToken = ?, act_expt = ?, refreshToken = ?, rft_expt = ?
	;`, usrId, act, act_expt, rft, rft_expt, act, act_expt, rft, rft_expt)

	if err != nil {
		return "", -1, "", -1, AppendError("generateTokenCouple: ", err)
	}

	signature, err := generateRsaSignature([]byte(act), privateKey)

	if err != nil {
		return "", -1, "", -1, AppendError("generateTokenCouple: ", err)
	}

	act_signed := act + "." + signature

	signature, err = generateRsaSignature([]byte(rft), privateKey)

	if err != nil {
		return "", -1, "", -1, AppendError("generateTokenCouple: ", err)
	}

	rft_signed := rft + "." + signature

	return act_signed, act_expt, rft_signed, rft_expt, nil
}

func accessTokenExists(access_token string) (bool, error) {

	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		return false, AppendError("accessTokenExists: ", err)
	}

	defer db.Close()
	var ret string
	q := fmt.Sprintf("SELECT accessToken FROM Token WHERE accessToken = \"%s\";", access_token)
	err = db.QueryRow(q).Scan(&ret)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, AppendError("accessTokenExists: ", err)
	}

	return true, nil
}

func refreshTokenExists(refresh_token string) (bool, error) {

	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		return false, AppendError("refreshTokenExists: ", err)
	}

	defer db.Close()
	var ret string
	q := fmt.Sprintf("SELECT refreshToken FROM Token WHERE refreshToken = \"%s\";", refresh_token)
	err = db.QueryRow(q).Scan(&ret)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, AppendError("refreshTokenExists: ", err)
	}

	return true, nil
}

//}}}

// get userid from email {{{
func getUserId_Email(userEmail string) (int, error) {

	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		return -1, AppendError("getUserId_Email: ", err)
	}

	defer db.Close()
	var ret int
	q := fmt.Sprintf("SELECT id FROM Users WHERE email = \"%s\";", userEmail)
	err = db.QueryRow(q).Scan(&ret)

	if err == sql.ErrNoRows {
		return -1, nil
	}

	if err != nil {
		return -1, AppendError("getUserId_Email: ", err)
	}

	return ret, nil
}

// }}}

// get bearer tokens from header {{{
func BearerAuthHeader(authHeader string) string {

	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, "Bearer")

	if len(parts) != 2 {
		return ""
	}

	tokenSigPair := strings.TrimSpace(parts[1])

	if len(tokenSigPair) < 1 {
		return ""
	}

	token, err := verifyRsaSignature(publicKey, tokenSigPair)

	if err != nil {
		return ""
	}

	ret, _ := validate(token, "")

	return ret
}

//}}}

// get usrid from access tokens {{{
func getAccessToken_usrid(access_token string) (int, error) {

	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		return -1, AppendError("getAccessToken_usrid: ", err)
	}

	defer db.Close()
	var ret actData
	err = db.QueryRow("SELECT userid, act_expt FROM Token WHERE accessToken = (?);", access_token).Scan(&ret.User_id, &ret.Act_expt)

	if err == sql.ErrNoRows {
		return -1, nil
	}

	if err != nil {
		return -1, AppendError("getAccessToken_usrid: ", err)
	}

	//? i think checking for nil is unnecessary
	// if ret.Access_token != "nil" {

	// check if token is expired
	if int(time.Now().Unix()) < ret.Act_expt {
		// update last login date
		err = updateLoginDate(ret.User_id)
		if err != nil {
			return -1, AppendError("getAccessToken_usrid: ", err)
		}
		return ret.User_id, nil
	}
	// }
	return -1, nil
}

// }}}

// login {{{
func backendLogin(usr_id int, password string) (bool, error) {

	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		return false, AppendError("backendLogin: ", err)
	}

	defer db.Close()
	var loginData UserData
	err = db.QueryRow("SELECT salt, pHash FROM Users WHERE id = (?);", usr_id).Scan(&loginData.Salt, &loginData.PHash)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, AppendError("backendLogin: ", err)
	}

	// compute hash of password and salt
	data := []byte(fmt.Sprint(loginData.Salt) + password)
	hash := sha256.Sum256(data)
	sum := fmt.Sprintf("%x", hash[:])

	if sum == loginData.PHash {
		return true, nil
	}
	return false, nil
}

func login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("endpoint hit: login")

	type ReqData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var resp ReqData

	err := httpGetBody(r, &resp)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	err = r.ParseForm()

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	email, ok := validate(resp.Email, "")

	if !ok {
		httpError(&w, 400, vEmailErr)
		return
	}

	password, ok := validate(resp.Password, validatePass)

	if !ok {
		httpError(&w, 400, vPassErr)
		return
	}

	usrId, err := getUserId_Email(email)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	if usrId == -1 {
		httpError(&w, 400, "wrong email or password")
		return
	}

	ret, err := backendLogin(usrId, password)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	if ret {

		if err != nil {
			httpError(&w, 500, err)
		}

		act, act_expt, rft, rft_expt, err := generateTokenCouple(usrId)

		if err != nil {
			httpError(&w, 500, err)
			return
		}

		err = updateLoginDate(usrId)

		if err != nil {
			httpError(&w, 500, err)
			return
		}

		httpSuccessf(&w, 200, `"access_token":"%s", "act_expt": %d,  "refresh_token":"%s", "rft_expt":%d`, act, act_expt, rft, rft_expt)
		return
	}

	httpError(&w, 400, "wrong email or password")
}

// }}}

// change user data {{{
func changeUserData(w http.ResponseWriter, r *http.Request) {
	fmt.Println("endpoint hit: change user data")

	headerContentTtype := r.Header.Get("Content-Type")
	if headerContentTtype != "application/json" {
		httpError(&w, 400, "Content Type is not application/json")
		return
	}

	type ReqData struct {
		New_username string `json:"new_username"`
		New_email    string `json:"new_email"`
		New_pw       string `json:"new_pw"`
		Old_pw       string `json:"old_pw"`
	}

	var resp ReqData

	err := httpGetBody(r, &resp)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	err = r.ParseForm()

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	act := BearerAuthHeader(r.Header.Get("Authorization"))
	new_username, ok := validate(resp.New_username, validateUser)

	if !ok {
		httpError(&w, 400, vUserErr)
		return
	}

	new_email, ok := validate(resp.New_email, "")

	if !ok {
		httpError(&w, 400, vEmailErr)
		return
	}

	new_pw, ok := validate(resp.New_pw, validatePass)

	if !ok {
		httpError(&w, 400, vPassErr)
		return
	}

	old_pw, ok := validate(resp.Old_pw, validatePass)

	if !ok {
		httpError(&w, 400, vPassErr)
		return
	}

	usrId, err := getAccessToken_usrid(act)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	if usrId == -1 {
		httpError(&w, 400, "invalid access token")
		return
	}

	ret, err := backendLogin(usrId, old_pw)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	if ret {

		q := "UPDATE Users SET"

		sumChangesFlag := false

		if new_username != "" {
			usrid_1, err := getUserId_Usr(new_username)
			if err != nil {
				httpError(&w, 500, err)
				return
			}
			if usrid_1 != -1 {
				httpError(&w, 400, "username already exists")
				// fmt.Fprint(w, "{ \"resp_code\":300, error: \"username already exists\" }")
				return
			}
			q += " username = \"" + new_username + "\","
			sumChangesFlag = true
		}

		if new_email != "" {
			usrid_1, err := getUserId_Email(new_email)
			if err != nil {
				httpError(&w, 500, err)
				return
			}
			if usrid_1 != -1 {
				httpError(&w, 400, "account using this email already exists")
				// fmt.Fprint(w, "{ \"resp_code\":300, error: \"account using this email already exists\" }")
				return
			}
			q += " email = \"" + new_email + "\","
			sumChangesFlag = true
		}

		if new_pw != "" {
			loginState, err := backendLogin(usrId, new_pw)
			if err != nil {
				httpError(&w, 500, err)
				return
			}
			if !loginState {
				httpError(&w, 400, "password already in use")
				// fmt.Fprint(w, "{ \"resp_code\":300, error: \"password already in use\" }")
				return
			}

			salt := RandomInt(100000)

			data := []byte(fmt.Sprint(salt) + new_pw)

			hash := sha256.Sum256(data)
			sum := fmt.Sprintf("%x", hash[:])

			q += " salt = " + fmt.Sprintf("%d", salt) + ", pHash = \"" + sum + "\","
			sumChangesFlag = true
		}

		if !sumChangesFlag {
			httpError(&w, 400, "you need to specify a change")
			// fmt.Fprintf(w, "{ \"resp_code\":300, error:\"you need to specify a change\"  }")
			return
		}

		q = q[:len(q)-1]

		q += " WHERE id = " + fmt.Sprintf("%d", usrId) + ";"

		db, err := sql.Open("mysql", databaseString)

		if err != nil {
			httpError(&w, 500, err)
			return
		}

		defer db.Close()

		_, err = db.Exec(q)

		if err != nil {
			httpError(&w, 500, err)
			return
		}

		httpSuccess(&w, 200, "data changed successfully")
		// fmt.Fprint(w, "{ \"resp_code\":200, error: \"data changed successfully\" }")
		return
	}

	httpError(&w, 400, "invalid access token")
}

// }}}

// get user data {{{
func getUserDataReq(w http.ResponseWriter, r *http.Request) {
	fmt.Println("endpoint hit: get use data")

	act := BearerAuthHeader(r.Header.Get("Authorization"))

	usrId, err := getAccessToken_usrid(act)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	if usrId == -1 {
		httpError(&w, 400, "invalid access token")
		// fmt.Fprintf(w, "{ \"resp_code\":400, error:\"invalid access token\"  }")
		return
	}

	username, email, date_of_join, err := getUserData(usrId)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	if username == "" {
		httpError(&w, 400, "id is empty")
		// fmt.Fprintf(w, "{ \"resp_code\":400, error:\"invalid wrong id\"  }")
		return
	}

	httpSuccessf(&w, 200, `"username": "%v", "email": "%v", "date_of_join": "%v"`, username, email, date_of_join)
}

// }}}

// sign in {{{
func signIn(w http.ResponseWriter, r *http.Request) {
	fmt.Println("endpoint hit: sign in")

	headerContentTtype := r.Header.Get("Content-Type")
	if headerContentTtype != "application/json" {
		httpError(&w, 400, "Content Type is not application/json")
		return
	}

	type ReqData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Username string `json:"username"`
	}

	var re ReqData

	err := httpGetBody(r, &re)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	err = r.ParseForm()

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	username, ok := validate(re.Username, validateUser)

	if !ok {
		httpError(&w, 400, vUserErr)
		return
	}

	email, ok := validate(re.Email, "")

	if !ok {
		httpError(&w, 400, vEmailErr)
		return
	}

	password, ok := validate(re.Password, validatePass)

	if !ok {
		httpError(&w, 400, vPassErr)
		return
	}

	resp, err := addUser(username, email, password)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	if !resp {
		httpError(&w, 400, "username or email already in use")
		// fmt.Fprint(w, "{ \"resp_code\":400 \"error\":\"username or email already in use\" }")
		return
	}

	httpSuccess(&w, 200, "sign in succesfull")
}

// }}}

// retrive password {{{
type otpStruct struct {
	UserId int    `db:"userId"`
	Otp    string `db:"otp"`
	Expt   int    `db:"expt"`
}

func getOtp() (string, error) {

	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		return "", err
	}

	defer db.Close()
	otp := RandomString(32)
	var o otpStruct
	err = db.QueryRow("SELECT otp FROM PwOtp WHERE otp = (?);", otp).Scan(&o.Otp)

	if err == sql.ErrNoRows {
		return otp, nil
	}

	if err != nil {
		return "", err
	}

	return "", nil
}

// send otp email {{{
func send_otp_retrivePassword(w http.ResponseWriter, r *http.Request) {
	fmt.Println("endpoint hit: retrive password (get otp)")

	// headerContentTtype := r.Header.Get("Content-Type")
	// if headerContentTtype != "application/json" {
	// 	httpError(&w, 400, "Content Type is not application/json")
	// 	return
	// }

	// type ReqData struct {
	// 	Email string `json:"email"`
	// }

	// var resp ReqData

	// err := httpGetBody(r, &resp)

	// if err != nil {
	// 	httpError(&w, 500, err)
	// 	return
	// }

	err := r.ParseForm()

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	vars := mux.Vars(r)
	email, ok := validate(vars["email"], "")

	if !ok {
		httpError(&w, 400, vEmailErr)
		return
	}

	// email := validate(r.Form.Get("email"))
	usrId, err := getUserId_Email(email)
	if err != nil {
		httpError(&w, 500, err)
		return
	}

	if usrId == -1 {
		httpError(&w, 400, "no user connected to email")
		// fmt.Fprint(w, "{ \"resp_code\":400, error: \"no user connected to email\" }")
		return
	}

	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	defer db.Close()

	expt := int(time.Now().Unix()) + 60

	otp := ""
	for {
		otp, err = getOtp()
		if err != nil {
			httpError(&w, 500, err)
			return
		}
		if otp != "" {
			break
		}
	}

	_, err = db.Exec(`
	INSERT INTO PwOtp VALUES ((?), (?), (?))
	ON DUPLICATE KEY
	UPDATE userId = (?), otp = (?), expt = (?)
	;`, usrId, otp, expt, usrId, otp, expt)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	err = sendEmail(email, "instan-tex otp code", "\n the code is: "+otp)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	httpSuccess(&w, 200, "otp sended successfully")
}

// }}}

// use otp password retrival {{{
func use_otp_retrivePassword(w http.ResponseWriter, r *http.Request) {
	fmt.Println("endpoint hit: retrive password (use otp)")

	headerContentTtype := r.Header.Get("Content-Type")
	if headerContentTtype != "application/json" {
		httpError(&w, 400, "Content Type is not application/json")
		return
	}

	type ReqData struct {
		New_pw string `json:"new_password"`
	}

	var resp ReqData

	err := httpGetBody(r, &resp)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	err = r.ParseForm()

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	new_password, ok := validate(resp.New_pw, validatePass)

	if !ok {
		httpError(&w, 400, vPassErr)
		return
	}

	otp := BearerAuthHeader(r.Header.Get("Authorization"))

	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	defer db.Close()

	var otpData otpStruct

	err = db.QueryRow("SELECT userId, expt FROM PwOtp WHERE otp = (?);", otp).Scan(&otpData.UserId, &otpData.Expt)

	if err == sql.ErrNoRows {
		httpError(&w, 400, "token does not exists")
		// fmt.Fprint(w, "{ \"resp_code\":400, error:\"token does not exists \" }")
		return
	}

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	if int(time.Now().Unix()) > otpData.Expt {
		httpError(&w, 400, "token expired")
		// fmt.Fprint(w, "{ \"resp_code\":400, error:\"token expired\" }")
		return
	}

	// db, err = sql.Open("mysql", databaseString)

	// if err != nil {
	// 	httpError(&w, 500, err)
	// 	return
	// }

	defer db.Close()

	salt := RandomInt(100000)

	data := []byte(fmt.Sprint(salt) + new_password)

	hash := sha256.Sum256(data)
	sum := fmt.Sprintf("%x", hash[:])

	_, err = db.Exec("UPDATE Users SET salt = (?), pHash = (?) WHERE id = (?) ", salt, sum, otpData.UserId)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	httpSuccess(&w, 200, "data changed successfully")
}

// }}}

// }}}
