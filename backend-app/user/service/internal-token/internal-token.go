package internaltoken

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/msyamsula/portofolio/backend-app/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type internalToken struct {
	appTokenSecret string
	appTokenTtl    time.Duration
}

func (s *internalToken) CreateToken(ctx context.Context, id, email, name string) (string, error) {
	var span trace.Span
	_, span = otel.Tracer("").Start(ctx, "internalToken.CreateToken")
	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
		}
		span.End()
	}()

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
		logger.Logger.Error(err.Error())
		return "", err
	}

	return tokenString, nil
}

func (g *internalToken) ValidateToken(ctx context.Context, tokenString string) (UserData, error) {
	var span trace.Span
	_, span = otel.Tracer("").Start(ctx, "internalToken.ValidateToken")
	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
		}
		span.End()
	}()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (secret interface{}, err error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secret, nil
	})

	if err != nil {
		logger.Logger.Error(err.Error())
		return UserData{}, errors.New("error parsing token")
	}

	// Check if token is valid
	if !token.Valid {
		err = errors.New("invalid token")
		logger.Logger.Error(err.Error())
		return UserData{}, err
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
		err = errors.New("cannot extract claims")
		logger.Logger.Error(err.Error())
		return UserData{}, err
	}

}
