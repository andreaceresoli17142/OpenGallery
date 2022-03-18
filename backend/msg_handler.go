package main

import ( // {{{
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
) // }}}

// get conversations {{{
func getConversations(w http.ResponseWriter, r *http.Request) {
	fmt.Println("endpoint hit: get conversations")
	access_token := BearerAuthHeader(r.Header.Get("Authorization"))

	usr_id, err := getAccessToken_usrid(access_token)

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	if usr_id == -1 {
		httpError(&w, 300, "invalid access token")
		return
	}

	db, err := sql.Open("mysql", databaseString)

	defer db.Close()

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	res, err := db.Query("SELECT c.id as id, cn.name as name, cn.description as description FROM Conversations c INNER JOIN ConversationName cn WHERE c.participantId = (?) GROUP BY c.id;", usr_id)

	defer res.Close()

	if err == sql.ErrNoRows {
		var r []string
		httpSuccessf(&w, 200, "data", r)
		return
	}

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	var convs []Conversation

	for res.Next() {
		var conv Conversation

		err := res.Scan(&conv.Id, &conv.Name, &conv.Description)

		if err != nil {
			httpError(&w, 500, err)
			return
		}

		convs = append(convs, conv)
	}

	a, _ := json.Marshal(convs)

	httpSuccessf(&w, 200, "data", string(a))
}

//! NEEDS TESTING
// takes userid, name and description, if description is empty we are creaing a conversation otherwise we are creating a group
// then create conversation, add user and data, return conversation id
func addConversations(userId int, name string, description string) (int, error) {

	// generating conversation id
	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		return -1, AppendError("error: addConversation", err)
	}

	defer db.Close()

	var convId int
	err = db.QueryRow(`SELECT COUNT(*) FROM (SELECT * FROM Conversations GROUP BY id) as a;`).Scan(convId)

	if err == sql.ErrNoRows {
		return -1, AppendError("error: addConversation", err)
	}

	if err != nil {
		return -1, AppendError("error: addConversation", err)
	}

	convId = convId + 1
	// finished conversation id

	// adding conversation
	db, err = sql.Open("mysql", databaseString)

	if err != nil {
		return -1, AppendError("error: addConversation", err)
	}

	defer db.Close()

	err = db.QueryRow(`INSERT INTO Conversations (id, participantId) VALUES (?, ?);`, convId, userId).Scan() // .Scan(convId)

	// if err == sql.ErrNoRows {
	// 	return -1, AppendError("error: addConversation", err)
	// }

	if err != nil {
		return -1, AppendError("error: addConversation", err)
	}
	// finised adding conversation

	// if the conversation is not a group exit the function
	if name == "" {
		return convId, nil
	}

	// adding conversation name and description
	db, err = sql.Open("mysql", databaseString)

	if err != nil {
		return -1, AppendError("error: addConversation", err)
	}

	defer db.Close()

	// if description is set then add it
	if description != "" {
		err = db.QueryRow(`INSERT INTO GroupName ( name, description ) VALUES (?, ?);`, name, description).Scan() // .Scan(convId)
	} else { // otherwise set descritpion as null
		err = db.QueryRow(`INSERT INTO GroupName ( name, description ) VALUES (?, ?);`, name, nil).Scan() // .Scan(convId)
	}

	// if err == sql.ErrNoRows {
	// 	return -1, AppendError("error: addConversation", err)
	// }

	if err != nil {
		return -1, AppendError("error: addConversation", err)
	}

	return convId, nil
}

func addUserToConv(convId int, userId int) error {
	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		return AppendError("error: addUserToConv", err)
	}

	defer db.Close()

	err = db.QueryRow(`INSERT INTO Conversations (id, participantId) VALUES (?, ?);`, convId, userId).Scan() // .Scan(convId)

	// if err == sql.ErrNoRows {
	// 	return AppendError("error: addUserToConv", err)
	// }

	if err != nil {
		return AppendError("error: addUserToConv", err)
	}
	return nil
}

// }}}

// firends requests {{{

func makeFriendRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("endpoint hit: get friend request")

	parts := strings.Split(r.Header.Get("Authorization"), "Bearer")

	if len(parts) != 2 {
		return
	}

	access_token := strings.TrimSpace(parts[1])

	// access_token := BearerAuthHeader(r.Header.Get("Authorization"))

	requesterId, err := getAccessToken_usrid(access_token)

	if err != nil {
		httpError(&w, 500, "access token not valid")
		return
	}

	type ReqData struct {
		userId string `json:"reciverid"`
	}

	var resp ReqData

	err = httpGetBody(r, &resp)

	if err != nil {
		httpError(&w, 500, "backend error")
		return
	}

	requesteeId, _ := strconv.Atoi(resp.userId)

	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		httpError(&w, 500, "backend error")
		return
	}

	defer db.Close()

	var ret int

	err = db.QueryRow(`SELECT id FROM FriendRequests WHERE senderId = ? AND reciverId = ?;`, requesterId, requesteeId).Scan(&ret) // .Scan(convId)

	if err != sql.ErrNoRows {
		httpError(&w, 500, "firend request already exists")
		return
	}

	if err != nil && err != sql.ErrNoRows {
		httpError(&w, 500, "backend error: "+err.Error())
		return
	}

	db, err = sql.Open("mysql", databaseString)

	if err != nil {
		httpError(&w, 500, "backend error: "+err.Error())
		return
	}

	defer db.Close()

	_, err = db.Exec(`INSERT INTO FriendRequests (senderId, reciverId) VALUES (?, ?);`, requesterId, requesteeId) // .Scan(convId)

	if err != nil && err != sql.ErrNoRows {
		httpError(&w, 500, "backend error: "+err.Error())
		return
	}

	httpSuccess(&w, 200, "request successfully sent")
}

func getFriendRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("endpoint hit: get friend request")

	parts := strings.Split(r.Header.Get("Authorization"), "Bearer")

	if len(parts) != 2 {
		return
	}

	access_token := strings.TrimSpace(parts[1])

	// access_token := BearerAuthHeader(r.Header.Get("Authorization"))

	userId, err := getAccessToken_usrid(access_token)

	if err != nil {
		httpError(&w, 500, "access token not valid")
		return
	}

	db, err := sql.Open("mysql", databaseString)

	defer db.Close()

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	res, err := db.Query(`
	SELECT f.id, f.senderId, u.username
	FROM FriendRequests f INNER JOIN Users u ON u.id = f.senderId
	WHERE f.reciverId = ? ;
	`, userId)

	defer res.Close()

	if err == sql.ErrNoRows {
		var r []string
		httpSuccessf(&w, 200, "data", r)
		return
	}

	if err != nil {
		httpError(&w, 500, err)
		return
	}

	var convs []Conversation

	for res.Next() {
		var conv Conversation

		err := res.Scan(&conv.Id, &conv.Name, &conv.Description)

		if err != nil {
			httpError(&w, 500, err)
			return
		}

		convs = append(convs, conv)
	}

	a, _ := json.Marshal(convs)

	httpSuccessf(&w, 200, "data", string(a))
}

//}}}
