package main

import ( // {{{

	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"runtime"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
) // }}}

//TODO: implement websockets to notify user of new comments
//TODO: prevent xss attacks in comments
//TODO: set characetr limits on comment parameters

var (
	store   = sessions.NewCookieStore([]byte("like-cookies"))
	fileDir string
)

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
		w.Header().Set("Access-Control-Allow-Credentials", "true")
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

type ArtworkData struct {
	Id             int            `db:"Id" json:"Id"`
	OriginalTitle  string         `db:"OriginalTitle" json:"OriginalTitle"`
	Title          sql.NullString `db:"Title" json:"Title"`
	YearOfCreation int            `db:"YearOfCreation" json:"YearOfCreation"`
	Description    string         `db:"Description" json:"Description"`
	Likes          int            `db:"Likes" json:"Likes"`
	Owner          string         `db:"Owner" json:"Owner"`
	BorrowedTo     sql.NullString `db:"BorrowedTo" json:"BorrowedTo"`
	Artists        []string       `json:"Artists"`
	Pictures       []string       `json:"Pictures"`
}

type ArtistData struct {
	Id          int            `db:"Id" json:"Id"`
	Name        string         `db:"Name" json:"Name"`
	SecondName  sql.NullString `db:"SecondName" json:"SecondName"`
	Surname     string         `db:"Surname" json:"Surname"`
	DateOfBirth string         `db:"DateOfBirth" json:"DateOfBirth"`
	Nationality string         `json:"Nationality"`
	Description sql.NullString `db:"Description" json:"Description"`
}

func getAllArtwork(w http.ResponseWriter, r *http.Request) {
	Debugln("endpoint hit: getAllArtwork")

	/* open db */
	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		httpError(&w, 500, AppendError("getAllArtwork [opening db]: ", err).Error())
		return
	}

	defer db.Close()

	/* get artworks */
	var artworks []*ArtworkData

	res, err := db.Query(`
		SELECT
			Id,
			OriginalTitle,
			Title,
			YearOfCreation,
			Description,
			Owner,
			BorrowedTo,
			Likes
		FROM Artworks`)

	if err != nil {
		httpError(&w, 500, err.Error())
		return
	}

	for res.Next() {
		var tmp ArtworkData

		err := res.Scan(&tmp.Id, &tmp.OriginalTitle, &tmp.Title, &tmp.YearOfCreation, &tmp.Description, &tmp.Owner, &tmp.BorrowedTo, &tmp.Likes)

		if err != nil {
			httpError(&w, 500, err)
			return
		}

		artworks = append(artworks, &tmp)
	}

	/* append artists and pictures */
	for _, artwork := range artworks {

		imageId := artwork.Id

		/* get artists name */
		var artist ArtistData

		res, err := db.Query(`
			SELECT Name, SecondName, Surname
			FROM Artists
			WHERE Id IN (
				SELECT ArtistId
				FROM CreatedBy
				WHERE ArtworkId = (
					SELECT Id
					FROM Artworks
					WHERE Id = ?
				)
			)`, imageId)

		if err != nil {
			httpError(&w, 500, err.Error())
			return
		}

		for res.Next() {

			err := res.Scan(&artist.Name, &artist.SecondName, &artist.Surname)

			if err != nil {
				httpError(&w, 500, err)
				return
			}

			namestr := artist.Name
			if artist.SecondName.Valid {
				namestr += " " + artist.SecondName.String
			}
			namestr += " " + artist.Surname

			artwork.Artists = append(artwork.Artists, namestr)
		}

		/* get artwork picture paths */
		var picPath string

		res, err = db.Query(`
			SELECT PicturePath
			FROM ArtworkPicture
			WHERE ArtworkId = (
				SELECT Id
				FROM Artworks
				WHERE Id = ?
			)`, imageId)

		if err != nil {
			httpError(&w, 500, err.Error())
			return
		}

		for res.Next() {

			err := res.Scan(&picPath)

			if err != nil {
				httpError(&w, 500, err)
				return
			}

			artwork.Pictures = append(artwork.Pictures, picPath)
		}

	}

	/* return json */
	ret, err := json.Marshal(artworks)

	if err != nil {
		httpError(&w, 500, err.Error())
		return
	}

	httpSuccessRaw(&w, 200, string(ret))
}

func getAllArtworkRanked(w http.ResponseWriter, r *http.Request) {
	Debugln("endpoint hit: getAllArtwork")

	/* open db */
	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		httpError(&w, 500, AppendError("getAllArtwork [opening db]: ", err).Error())
		return
	}

	defer db.Close()

	/* get artworks */
	var artworks []*ArtworkData

	res, err := db.Query(`
		SELECT
			Id,
			OriginalTitle,
			Likes
		FROM Artworks
		ORDER BY Likes DESC`)

	if err != nil {
		httpError(&w, 500, err.Error())
		return
	}

	for res.Next() {
		var tmp ArtworkData

		err := res.Scan(&tmp.Id, &tmp.OriginalTitle, &tmp.Likes)

		if err != nil {
			httpError(&w, 500, err)
			return
		}

		artworks = append(artworks, &tmp)
	}

	/* append artists */
	for _, artwork := range artworks {

		imageId := artwork.Id

		/* get artists name */
		var artist ArtistData

		res, err := db.Query(`
			SELECT Name, SecondName, Surname
			FROM Artists
			WHERE Id IN (
				SELECT ArtistId
				FROM CreatedBy
				WHERE ArtworkId = (
					SELECT Id
					FROM Artworks
					WHERE Id = ?
				)
			)`, imageId)

		if err != nil {
			httpError(&w, 500, err.Error())
			return
		}

		for res.Next() {

			err := res.Scan(&artist.Name, &artist.SecondName, &artist.Surname)

			if err != nil {
				httpError(&w, 500, err)
				return
			}

			namestr := artist.Name
			if artist.SecondName.Valid {
				namestr += " " + artist.SecondName.String
			}
			namestr += " " + artist.Surname

			artwork.Artists = append(artwork.Artists, namestr)
		}

	}

	/* return json */
	ret, err := json.Marshal(artworks)

	if err != nil {
		httpError(&w, 500, err.Error())
		return
	}

	httpSuccessRaw(&w, 200, string(ret))
}

func getSingleArtwork(w http.ResponseWriter, r *http.Request) {
	Debugln("endpoint hit: getSingleArtwork")

	/* get delected id */
	vars := mux.Vars(r)
	imageId := vars["id"]

	var artwork ArtworkData

	/* open db */
	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		httpError(&w, 500, AppendError("getSingleArtwork [opening db]: ", err).Error())
		return
	}

	defer db.Close()

	/* get basic artwork data */
	err = db.QueryRow(`
		SELECT
			Id,
			OriginalTitle,
			Title,
			YearOfCreation,
			Description,
			Owner,
			BorrowedTo,
			Likes
		FROM Artworks
		WHERE Id = ?`, imageId).Scan(&artwork.Id, &artwork.OriginalTitle, &artwork.Title, &artwork.YearOfCreation, &artwork.Description, &artwork.Owner, &artwork.BorrowedTo, &artwork.Likes)

	if err == sql.ErrNoRows {
		httpError(&w, 404, "no entries")
		return
	}

	if err != nil {
		httpError(&w, 500, err.Error())
		return
	}

	/* get artists name */
	var artist ArtistData

	res, err := db.Query(`
		SELECT Name, SecondName, Surname
		FROM Artists
		WHERE Id IN (
			SELECT ArtistId
			FROM CreatedBy
			WHERE ArtworkId = (
				SELECT Id
				FROM Artworks
				WHERE Id = ?
			)
		)`, imageId)

	if err != nil {
		httpError(&w, 500, err.Error())
		return
	}

	for res.Next() {

		err := res.Scan(&artist.Name, &artist.SecondName, &artist.Surname)

		if err != nil {
			httpError(&w, 500, err)
			return
		}

		namestr := artist.Name
		if artist.SecondName.Valid {
			namestr += " " + artist.SecondName.String
		}
		namestr += " " + artist.Surname

		artwork.Artists = append(artwork.Artists, namestr)
	}

	/* get artwork picture paths */
	var picPath string

	res, err = db.Query(`
		SELECT PicturePath
		FROM ArtworkPicture
		WHERE ArtworkId = (
			SELECT Id
			FROM Artworks
			WHERE Id = ?
		)`, imageId)

	if err != nil {
		httpError(&w, 500, err.Error())
		return
	}

	for res.Next() {

		err := res.Scan(&picPath)

		if err != nil {
			httpError(&w, 500, err)
			return
		}

		artwork.Pictures = append(artwork.Pictures, picPath)
	}

	/* return json */
	ret, err := json.Marshal(artwork)

	if err != nil {
		httpError(&w, 500, err.Error())
		return
	}

	httpSuccessRaw(&w, 200, string(ret))
}

func getLikeStatus(w http.ResponseWriter, r *http.Request) {
	Debugln("endpoint hit: getLikeStatus")

	/* get delected id */
	vars := mux.Vars(r)
	imageId := vars["id"]
	session, _ := store.Get(r, "like-cookies")

	if session.Values[imageId] == nil {
		httpSuccessf(&w, 200, `"Value":%v`, false)
		return
	}

	httpSuccessf(&w, 200, `"Value":%v`, session.Values[imageId])
}

func getLikeNumber(w http.ResponseWriter, r *http.Request) {
	Debugln("endpoint hit: getLikeNumber")

	/* get delected id */
	vars := mux.Vars(r)
	imageId := vars["id"]

	/* open db */
	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		httpError(&w, 500, AppendError("getLikeNumber [opening db]: ", err).Error())
		return
	}

	defer db.Close()

	var liken int

	err = db.QueryRow(`SELECT Likes FROM Artworks WHERE Id = ?`, imageId).Scan(&liken)

	if err != nil {
		httpError(&w, 500, AppendError("getLikeNumber [querying db]: ", err).Error())
		return
	}

	httpSuccessf(&w, 200, `"Value":%v`, liken)
}

func toggleLike(w http.ResponseWriter, r *http.Request) {
	Debugln("endpoint hit: toggleLike")

	/* get delected id */
	vars := mux.Vars(r)
	imageId := vars["id"]
	session, _ := store.Get(r, "like-cookies")

	var likeStatus bool

	/* open db */
	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		httpError(&w, 500, AppendError("toggleLike [opening db]: ", err).Error())
		return
	}

	defer db.Close()

	if session.Values[imageId] == nil || session.Values[imageId] == false {

		_, err = db.Exec(`UPDATE Artworks SET Likes=Likes+1 WHERE Id = ?`, imageId)

		if err != nil {
			httpError(&w, 500, AppendError("toggleLike [adding like db]: ", err).Error())
			return
		}
		likeStatus = true
		session.Values[imageId] = true
	} else {

		_, err = db.Exec(`UPDATE Artworks SET Likes=Likes-1 WHERE Id = ?`, imageId)

		if err != nil {
			httpError(&w, 500, AppendError("toggleLike [removing like db]: ", err).Error())
			return
		}

		likeStatus = false
		session.Values[imageId] = false
	}

	err = session.Save(r, w)
	if err != nil {
		httpError(&w, 500, AppendError("toggleLike [saving session]: ", err).Error())
	}
	httpSuccessf(&w, 200, `"Value":%v`, likeStatus)
}

type Comment struct {
	Username string `db:"Username" json:"usr"`
	Text     string `db:"Comment" json:"text"`
}

func postcomment(w http.ResponseWriter, r *http.Request) {

	Debugln("endpoint hit: postcomment")

	/* get delected id */
	vars := mux.Vars(r)
	imageId := vars["id"]

	var toAdd Comment
	err := httpGetBody(r, &toAdd)
	// responseData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		httpError(&w, 300, "missing body parameter")
		return
	}

	if len(toAdd.Username) > 30 {
		httpError(&w, 300, "username too long")
		return
	}

	if toAdd.Username == "" || toAdd.Text == "" {
		httpError(&w, 300, "username or text are empty")
		return
	}

	/* open db */
	db, err := sql.Open("mysql", databaseString)

	_, err = db.Exec(`INSERT INTO Comments (ArtworkId, Username, Comment) VALUES (?, ?, ?)`, imageId, toAdd.Username, toAdd.Text)

	if err != nil {
		httpError(&w, 500, AppendError("postcomment [adding comment]: ", err).Error())
		return
	}

	httpSuccess(&w, 200, "successfull")
}

func getcomment(w http.ResponseWriter, r *http.Request) {

	Debugln("endpoint hit: getcomment")

	/* get delected id */
	vars := mux.Vars(r)
	imageId := vars["id"]

	var comments []*Comment

	/* open db */
	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		httpError(&w, 500, AppendError("postcomment [opening db]: ", err).Error())
		return
	}

	defer db.Close()

	res, err := db.Query(`SELECT Username, Comment FROM Comments WHERE ArtworkId = ?`, imageId)

	if err != nil {
		httpError(&w, 500, AppendError("postcomment [adding comment]: ", err).Error())
		return
	}

	for res.Next() {
		var tmp Comment

		err := res.Scan(&tmp.Username, &tmp.Text)

		if err != nil {
			httpError(&w, 500, err)
			return
		}

		comments = append(comments, &tmp)
	}

	ret, err := json.Marshal(comments)

	if err != nil {
		httpError(&w, 500, AppendError("error parsing response", err).Error())
		return
	}

	httpSuccessRaw(&w, 200, string(ret))
}

// route endpoints {{{
func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/", homePage) //.Schemes("https")
	myRouter.HandleFunc("/togglelike/{id}", toggleLike).Methods("POST", "OPTIONS")
	myRouter.HandleFunc("/getlikestatus/{id}", getLikeStatus).Methods("GET", "OPTIONS")
	myRouter.HandleFunc("/getlikenumber/{id}", getLikeNumber).Methods("GET", "OPTIONS")

	comments := myRouter.PathPrefix("/comment/").Subrouter()
	comments.HandleFunc("/post/{id}", postcomment).Methods("POST", "OPTIONS")
	comments.HandleFunc("/getall/{id}", getcomment).Methods("GET", "OPTIONS")

	artData := myRouter.PathPrefix("/art").Subrouter()
	artData.HandleFunc("/getartworkranked", getAllArtworkRanked).Methods("GET", "OPTIONS")
	artData.HandleFunc("/getartwork", getAllArtwork).Methods("GET", "OPTIONS")
	artData.HandleFunc("/getartwork/{id}", getSingleArtwork).Methods("GET", "OPTIONS")

	log.Fatal(http.ListenAndServe(":8080", corsMiddleware(myRouter)))
} // }}}

func init() {

	store.Options = &sessions.Options{
		Domain:   "localhost",
		Path:     "/",
		HttpOnly: true,
	}

	ok := loadEnv()

	// if enviroment variables loading fails exit the program
	if !ok {
		return
	}

	_, fileDir, _, ok = runtime.Caller(1)
	if !ok {
		log.Fatal("error getting file directory")
	}
}

func main() {
	Successln("GO server started")

	handleRequests()
}
