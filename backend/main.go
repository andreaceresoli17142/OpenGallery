package main

import ( // {{{

	"crypto/rand"
	"crypto/rsa"
	"log"
	"net/http"
	"runtime"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gorilla/mux"
) // }}}

var fileDir string

func homePage(w http.ResponseWriter, r *http.Request) { // {{{

	httpSuccess(&w, 200, "hey, this is a homepage")
}

// }}}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// log.Println("Executing middleware", r.Method)
		origin := r.Header["Origin"]
		// fmt.Println("origin", origin)
		if len(origin) > 0 {
			w.Header().Set("Access-Control-Allow-Origin", strings.Join(origin, ","))
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "OPTIONS" {
			headers := strings.Join(r.Header["Access-Control-Request-Headers"], ",")
			// fmt.Println("HEADERS", headers)
			w.Header().Set("Access-Control-Allow-Headers", headers)
			return
		}
		next.ServeHTTP(w, r)
		// log.Println("Executing middleware again")
	})
}

// route endpoints {{{
func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/", homePage) //.Schemes("https")
	myRouter.HandleFunc("/getusrdata", getUserDataReq).Methods("GET", "OPTIONS")
	myRouter.HandleFunc("/signin", signIn).Methods("POST", "OPTIONS")
	myRouter.HandleFunc("/change", changeUserData).Methods("POST", "OPTIONS")
	myRouter.HandleFunc("/getconversations", getConversations).Methods("GET", "OPTIONS")

	oauthRouter := myRouter.PathPrefix("/oauth").Subrouter()
	oauthRouter.HandleFunc("/", paleoIdAuth).Methods("GET", "OPTIONS")
	oauthRouter.HandleFunc("/getlink", getOauthLink).Methods("GET", "OPTIONS")
	oauthRouter.HandleFunc("/signin", signInOauth).Methods("POST", "OPTIONS")
	oauthRouter.HandleFunc("/gettkcoup/{state}", oauthGetTokenCouple).Methods("GET", "OPTIONS")

	authRouter := myRouter.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/login", login).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/userft", refreshTokenReq).Methods("POST", "OPTIONS")

	pwrRouter := myRouter.PathPrefix("/pwr").Subrouter()
	pwrRouter.HandleFunc("/getotp/{email}", send_otp_retrivePassword).Methods("GET", "OPTIONS")
	pwrRouter.HandleFunc("/useotp", use_otp_retrivePassword).Methods("POST", "OPTIONS")

	freqRouter := myRouter.PathPrefix("/freq").Subrouter()
	freqRouter.HandleFunc("/makereq", makeFriendRequest).Methods("POST", "OPTIONS")

	log.Fatal(http.ListenAndServe(":8080", corsMiddleware(myRouter)))
} // }}}

func init() {

	ok := loadEnv()

	// if enviroment variables loading fails exit the program
	if !ok {
		return
	}

	_, fileDir, _, ok = runtime.Caller(1)
	if !ok {
		log.Fatal("error getting file directory")
	}

	var err error
	privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	publicKey = &privateKey.PublicKey
}

func main() {
	Successln("GO server started")

	handleRequests()
}
