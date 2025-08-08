package main

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type user struct {
	id       int
	username string
	password string
}

func initDB() *sql.DB {

	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		panic(err)
	}

	initQuery := `
	create table if not exists auth (
		id int autoincrement not null primary key,
		username minitext unique not null,
		password char(60) not null
	);
	`

	_, err = db.Exec(initQuery)
	if err != nil {
		panic(err)
	}

	return db
}

func nameExists(db *sql.DB, username string) bool {

	row := db.QueryRow("Select * From auth where username = ?", username)

	err := row.Scan()
	if err == sql.ErrNoRows {
		return false
	} else {
		return true
	}
}

func makeNewUser(db *sql.DB, username string, password []byte) error {
	_, err := db.Exec("Insert Into auth (username, password) Values (?, ?)", username, password)
	if err != nil {
		panic(err)
	}
	return nil
}

func getId(db *sql.DB, username string) (int, error) {
	var user user

	row := db.QueryRow("Select id From auth Where username = ?", username)

	err := row.Scan(&user.id)
	if err != nil {
		return 0, errors.New("Could not find user")
	}

	return user.id, nil
}

func getHash(db *sql.DB, id int) (string, error) {

	var user user

	row := db.QueryRow("Select password From auth where id = ?", id)
	err := row.Scan(&user.password)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("User does not exist")
		}
		return "", err
	}

	return user.password, nil
}

func main() {
	db := initDB()

	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if hash == nil {
	}

	makeNewUser(db, "mqngo", hash)

	fmt.Println(nameExists(db, "hi"))
	name, _ := getId(db, "mqngo")
	fmt.Println(getHash(db, name))
}
