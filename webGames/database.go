package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type user struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type newUser struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

func initDB() *sql.DB {

	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		panic(err)
	}

	initQuery := `
	create table if not exists auth (
		username minitext unique not null primary key,
		password char(60) not null
	);
	`

	_, err = db.Exec(initQuery)
	if err != nil {
		panic(err)
	}

	return db
}

func userExists(db *sql.DB, username string) bool {
	row := db.QueryRow("Select * From auth where username = ?", username)
	fmt.Println(row)
	return true
}

func newUser(db *sql.DB, username string, password []byte) error {
	_, err := db.Exec("Insert Into auth (username, password) Values (?, ?)", username, password)
	if err != nil {
		panic(err)
	}
	return nil
}

func getHash(db *sql.DB, username string) {

	row := db.QueryRow("Select * From auth where username = ?", username)
}

func main() {
	db := initDB()

	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	newUser(db, "mqngo", hash)

	fmt.Println(userExists(db, "mqngo"))
}
