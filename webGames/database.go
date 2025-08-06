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
		id int auto increment primary key,
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

func newUser(db *sql.DB, username string, password []byte) error {
	_, err := db.Exec("Insert Into auth (username, password) Values (?, ?)", username, password)
	if err != nil {
		panic(err)
	}
	return nil
}

func getHash(db *sql.DB, username string) (string, error) {

	var user user

	row := db.QueryRow("Select * From auth where username = ?", username)
	err := row.Scan(&user.id, &user.username, &user.password)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("User does not exist")
		}
		return "", err
	}

	return user.password, nil
}

func ain() {
	db := initDB()

	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if hash == nil {
	}

	fmt.Println(nameExists(db, "hi"))
	fmt.Println(getHash(db, "mqngo"))
}
