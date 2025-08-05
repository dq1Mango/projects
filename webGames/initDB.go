package main

import "database/sql"
import "github.com/go-sql-driver/mysql"

func init() {
	db, err := sql.Open("mysql", "web")
}
