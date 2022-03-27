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
) // }}}

//TODO: implement audio commentary

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

type ArtworkData struct {
	Id             int            `db:"Id" json:"Id"`
	OriginalTitle  string         `db:"OriginalTitle" json:"OriginalTitle"`
	Title          sql.NullString `db:"Title" json:"Title"`
	YearOfCreation int            `db:"YearOfCreation" json:"YearOfCreation"`
	Description    string         `db:"Description" json:"Description"`
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
			BorrowedTo
		FROM Artworks`)

	if err != nil {
		httpError(&w, 500, err.Error())
		return
	}

	for res.Next() {
		var tmp ArtworkData

		err := res.Scan(&tmp.Id, &tmp.OriginalTitle, &tmp.Title, &tmp.YearOfCreation, &tmp.Description, &tmp.Owner, &tmp.BorrowedTo)

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
			BorrowedTo
		FROM Artworks
		WHERE Id = ?`, imageId).Scan(&artwork.Id, &artwork.OriginalTitle, &artwork.Title, &artwork.YearOfCreation, &artwork.Description, &artwork.Owner, &artwork.BorrowedTo)

	if err == sql.ErrNoRows {
		httpError(&w, 400, "no entries")
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

// route endpoints {{{
func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/", homePage) //.Schemes("https")

	artData := myRouter.PathPrefix("/art").Subrouter()
	artData.HandleFunc("/getartwork", getAllArtwork)
	artData.HandleFunc("/getartwork/{id}", getSingleArtwork).Methods("GET", "OPTIONS")

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
}

func main() {
	Successln("GO server started")

	handleRequests()
}
