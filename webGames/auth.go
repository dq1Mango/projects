package main

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var jwtKey = []byte("secret-key")
var tokenLife = 7 * 24 * time.Hour

type claims struct {
	Id      int       `json:"id"`
	Expires time.Time `json:"expires"` //dont get why this works
}

// this is a good idea
func (claim *claims) Valid() error { return nil }

// take that forced polymorphism

//	jwt.MapClaims{
//			"id":      id,
//			"expires": time.Now().Add(tokenLife),
//		}

func genToken(id int) (string, error) {
	claims := &claims{
		Id: id,

		Expires: time.Now().Add(tokenLife),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtKey)

	fmt.Println("heres our new token: ", tokenString)

	return tokenString, err
}

func decodeToken(tokenString string) (*claims, error) {
	// this gets a little ugly
	token, err := jwt.ParseWithClaims(
		tokenString, &claims{},
		func(token *jwt.Token) (any, error) { return jwtKey, nil },
	)
	if err != nil {
		panic(err)
	}

	if claims, ok := token.Claims.(*claims); ok {
		return claims, nil
	} else {
		fmt.Println(ok)
		return nil, nil
	}

}

func pain() {
	newToken, _ := genToken(67)
	claims, _ := decodeToken(newToken)
	fmt.Println(claims.Expires.Compare(time.Now()))
}
