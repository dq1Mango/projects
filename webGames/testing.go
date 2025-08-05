package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func man() {
	password := []byte("password")
	hash, _ := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	fmt.Println(hash)
	fmt.Println(len(hash))
	fmt.Println(bcrypt.CompareHashAndPassword(hash, []byte("askljdfhja")))
}
