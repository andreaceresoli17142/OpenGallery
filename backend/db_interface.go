package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func getUserId_Usr(username string) (int, error) {

	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		return -1, err
	}

	defer db.Close()
	var ret UserData
	// q := fmt.Sprintf("SELECT id FROM Users WHERE username = %s;", username)
	err = db.QueryRow("SELECT id FROM Users WHERE username = (?);", username).Scan(&ret.Id)

	if err == sql.ErrNoRows {
		return -1, nil
	}

	if err != nil {
		return -1, err
	}

	return ret.Id, nil
}

func getUserData(usrId int) (string, string, string, error) {

	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		return "", "", "", err
	}

	defer db.Close()
	var ret UserData
	q := fmt.Sprintf("SELECT username, email, date_of_join FROM Users WHERE id = %d;", usrId)
	err = db.QueryRow(q).Scan(&ret.Username, &ret.Email, &ret.Date_of_join)

	if err == sql.ErrNoRows {
		return "", "", "", nil
	}

	if err != nil {
		return "", "", "", err
	}

	return ret.Username, ret.Email, ret.Date_of_join, nil
}

func updateLoginDate(usrId int) error {
	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		return err
	}

	defer db.Close()

	_, err = db.Exec("UPDATE Users SET last_login = CURRENT_TIMESTAMP() WHERE id = (?) ;", usrId)

	if err != nil {
		return err
	}
	return nil
}

func userExists(username string, email string) (bool, error) {
	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		return false, err
	}

	defer db.Close()
	var ret string
	q := fmt.Sprintf("SELECT id FROM Users WHERE email = \"%s\" OR username = \"%s\";", email, username)
	err = db.QueryRow(q).Scan(&ret)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func addUser(username string, email string, password string) (bool, error) {

	usrExt, err := userExists(username, email)

	if err != nil {
		return false, err
	}

	if usrExt {
		return false, nil
	}

	// INSERT INTO Users ( username, email, date_of_join, salt, pHash ) VALUES ( "pima", "pippo.mario@gimelli.com", CURRENT_DATE(), 123456, "62d18522b74d75b2a84776c91ba5498377441d4c4af0cea22ca7de9e09475d3a" );
	db, err := sql.Open("mysql", databaseString)

	if err != nil {
		return false, err
	}

	defer db.Close()

	salt := RandomInt(100000)

	data := []byte(fmt.Sprint(salt) + password)

	hash := sha256.Sum256(data)
	sum := fmt.Sprintf("%x", hash[:])

	q := fmt.Sprintf("INSERT INTO Users ( username, email, date_of_join, salt, pHash ) VALUES (\"%s\", \"%s\", CURRENT_DATE(), %d, \"%s\");", username, email, salt, sum)

	_, err = db.Exec(q)

	if err != nil {
		return false, err
	}
	return true, nil
}
