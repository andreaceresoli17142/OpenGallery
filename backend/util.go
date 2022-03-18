package main

import ( //{{{
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/smtp"
	"regexp"
	"time"
) // }}}

// structs {{{
type OauthResp struct {
	AccessToken string `json:"access_token"`
}

type UserData struct {
	Id           int    `db:"id"`
	Username     string `db:"username"`
	Email        string `db:"email"`
	Date_of_join string `db:"date_of_join"`
	Salt         int    `db:"salt"`
	PHash        string `db:"pHash"`
}

type actData struct {
	User_id       int    `db:"userid"`
	Access_token  string `db:"accessToken"`
	Refresh_token string `db:"refreshToken"`
	Act_expt      int    `db:"act_expt"`
	Rft_expt      int    `db:"rft_expt"`
}

type Conversation struct {
	Id          int    `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
}

// }}}

// variables {{{
var (
	validateEmail string = `^[a-zA-Z0-9.!#$%&'*+=?^_{|}~-@]*$`
	validatePass  string = `^[a-zA-Z0-9.!#$%&'*+=?^_{|}~-]{8,}$`
	validateUser  string = `^[a-zA-Z0-9.!#$%&'*+=?^_{|}~-]{5,}$`

	vPassErr  string = "password must be at least 8 characters long"
	vUserErr  string = "username must be at least 6 characters long"
	vEmailErr string = "email is not valid"

	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
)

//}}}

//! orribile, da cambiare
func validate(input string, regex string) (string, bool) {
	// remove " ' < > / \ to validate user input
	re := regexp.MustCompile(`[\\\/\<\>\"\']*`)

	ok := false

	if regex != "" {
		var err error
		ok, err = regexp.Match(regex, []byte(input))
		// Debugf("ok? %v, err %v", ok, err)
		if err != nil {
			ok = false
		}
	} else {
		ok = true
	}

	return re.ReplaceAllString(input, ""), ok
}

func AppendError(s string, e error) error {
	return fmt.Errorf("%s: %e", s, e)
}

func RandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func RandomInt(n int) int {

	rand.Seed(time.Now().UnixNano())

	return rand.Int() % n
}

//http returns {{{
func httpError(w *http.ResponseWriter, code int, msg interface{}) {
	Debugf("Error: %v", msg)
	(*w).Header().Set("Content-Type", "application/json")
	(*w).WriteHeader(code)
	fmt.Fprintf((*w), `{"code": %d, "msg":"%s"}`, code, msg)
}

func httpSuccess(w *http.ResponseWriter, code int, s string) {
	(*w).Header().Set("Content-Type", "application/json")
	(*w).WriteHeader(code)
	fmt.Fprintf((*w), `{"code": %d, "msg":"%s"}`, code, s)
}

func httpSuccessf(w *http.ResponseWriter, code int, s string, args ...interface{}) {
	(*w).Header().Set("Content-Type", "application/json")
	(*w).WriteHeader(code)
	rs := fmt.Sprintf(s, args...)
	fmt.Fprintf((*w), `{"code": %d, %s}`, code, rs)
}

//}}}

func httpGetBody(r *http.Request, v interface{}) error {
	b, err := ioutil.ReadAll(r.Body)

	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &v)

	if err != nil {
		return err
	}
	return nil
}

func sendEmail(reciver_email string, subject string, messagge string) error {

	auth := smtp.PlainAuth("", email_email, email_password, email_server)

	to := []string{reciver_email}
	msg_string := "To:" + reciver_email + "\r\nSubject: " + subject + "\r\n\r\n" + messagge + "\r\n"
	msg := []byte(msg_string)
	err := smtp.SendMail(email_server+":"+email_port, auth, reciver_email, to, msg)

	if err != nil {
		return AppendError("error sending mail: ", err)
	}

	return nil
}
