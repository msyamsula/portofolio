package internaltoken

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type internalToken struct {
	appTokenSecret string
	appTokenTtl    time.Duration
}

func (s *internalToken) CreateToken(id, email, name string) (string, error) {
	// Secret used for signing
	secret := []byte(s.appTokenSecret)

	// Create a new token object
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    id,
		"name":  name,
		"email": email,
		"exp":   time.Now().Add(s.appTokenTtl).Unix(),
	})

	// Sign the token with the secret
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (g *internalToken) ValidateToken(ctx context.Context, tokenString string) (UserData, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (secret interface{}, err error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secret, nil
	})

	if err != nil {
		return UserData{}, errors.New("error parsing token")
	}

	// Check if token is valid
	if !token.Valid {
		return UserData{}, errors.New("invalid token")
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// fmt.Println("Token is valid")
		// fmt.Println("User ID:", claims["user_id"])
		// fmt.Println("Email:", claims["email"])
		// fmt.Println("Expires at:", claims["exp"])
		return UserData{
			ID:    claims["user_id"].(string),
			Email: claims["email"].(string),
			Name:  claims["name"].(string),
		}, nil
	} else {
		return UserData{}, errors.New("Cannot extract claims")
	}

}
