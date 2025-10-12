package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
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

	// crude small statement to make sure the db works how we think it works that we might as well just run every boot up
	initQuery := `
	create table if not exists auth (
		id int auto_increment not null primary key,
		username minitext unique not null,
		password char(60) not null
	);
	create table if not exists nextUserId (id int);
	Insert Into nextUserId (id) Select 0 Where NOT Exists (select * from nextUserId);

	create table if not exists friends (id int, json text);
	`
	//Insert into auth (id, username, password) values ('1', 'null', '000000000000000000000000000000000000000000000000000000000000');

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

// couldnt get auto_increment to work so i just made it myself
func getNextId(db *sql.DB) int {
	var id int

	row := db.QueryRow("select * from nextUserId")

	err := row.Scan(&id)
	if err != nil {
		return -1
	}

	_, err = db.Exec("Update nextUserId Set id = ?", id+1)
	if err != nil {
		panic(err)
	}

	return id
}

func makeNewUser(db *sql.DB, username string, password []byte) (int, error) {
	nextId := getNextId(db)
	_, err := db.Exec("Insert Into auth (id, username, password) Values (?, ?, ?)", nextId, username, password)
	if err != nil {
		panic(err)
	}
	return nextId, nil
}

func getId(db *sql.DB, username string) (int, error) {
	var user user

	row := db.QueryRow("Select id From auth Where username = ?", username)

	err := row.Scan(&user.id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("Could not find user")
		}
		return 0, err
	}

	return user.id, nil
}

func getUserName(db *sql.DB, id int) (string, error) {
	var user user

	row := db.QueryRow("Select username From auth where id = ?", id)
	err := row.Scan(&user.username)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("User does not exist")
		}
		return "", err
	}

	return user.username, nil

}

func getHash(db *sql.DB, id int) ([]byte, error) {

	var user user

	row := db.QueryRow("Select password From auth where id = ?", id)
	err := row.Scan(&user.password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("User does not exist")
		}
		return nil, err
	}

	return []byte(user.password), nil
}

func contains(slice []int, value int) (int, error) {
	for i, v := range slice {
		if v == value {
			return i, nil
		}
	}

	return -1, ErrNotInSlice

}

// ty stack overflow
func remove(slice []int, s int) []int {
	index, err := contains(slice, s)

	if err != nil {
		return slice
	}

	return append(slice[:index], slice[index+1:]...)
}

func getFriends(db *sql.DB, id int) ([]int, error) {
	var jsonFriends []byte
	var friends []int

	row := db.QueryRow("Select json From friends where id = ?", id)
	err := row.Scan(&jsonFriends)

	if err != nil {
		if err == sql.ErrNoRows {

			// manual json encoding which will surely not come back to bite me
			db.Exec("Insert into friends (id, json) Values (?, ?)", id, "{[]}")

			// this is not scary at all
			return getFriends(db, id)
		}

		return nil, err
	}

	json.Unmarshal(jsonFriends, &friends)

	return friends, nil
}

func addFriend(db *sql.DB, id int, friend int) error {

	//_, err := db.Exec("Insert Into friends ()", id)

	friends, err := getFriends(db, id)

	if err != nil {
		return err
	}

	index, err := contains(friends, friend)

	if index != -1 && err != ErrNotInSlice {
		return ErrAlreadyInSlice
	}

	friends = append(friends, friend)

	jsonFriends, err := json.Marshal(friends)

	_, err = db.Exec("Update friends Set json = ? Where id = ?", jsonFriends, id)

	return err
}

func removeFriend(db *sql.DB, id int, friend int) error {
	friends, err := getFriends(db, id)

	if err != nil {
		return err
	}

	// not a built-in believe it or not
	friends = remove(friends, friend)

	jsonFriends, err := json.Marshal(friends)

	_, err = db.Exec("Update friends Set json = ? Where id = ?", jsonFriends, id)

	return err

}

type Profile struct {
	Id       int
	Name     string
	Status   int
	Duration int
}

// big tech core
func getUserProfile(db *sql.DB, id int) (Profile, error) {
	var profile Profile

	row := db.QueryRow("Select * From profile Where id = ?", id)
	err := row.Scan(&profile)
	if err != nil {
		return profile, err
	}

	return profile, nil
}

// all these get_ functions kinda do the same thing ... hmmmmm
func ain() {
	db := initDB()

	// hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	// if hash == nil {
	// }
	//
	// makeNewUser(db, "mqngo2", hash)
	//
	// fmt.Println(nameExists(db, "hi"))
	// name, _ := getId(db, "mqngo")
	// fmt.Println(getHash(db, name))

	err := addFriend(db, 3, 0)

	fmt.Println(err)

	friends, err := getFriends(db, 0)
	fmt.Println(err)

	fmt.Println(friends)
}
