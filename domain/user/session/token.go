package session

import (
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// token is jwt
func CreatToken(id int64, username string) string {

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{},
		Id:               id,
		Username:         username,
	}
	sessionToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, _ := sessionToken.SignedString("key")
	return signedToken
}
