package internaltoken

import "time"

type UserData struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type InternalTokenConfig struct {
	AppTokenSecret string
	AppTokenTtl    time.Duration
}
