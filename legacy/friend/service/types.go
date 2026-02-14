package service

import (
	repo "github.com/msyamsula/portofolio/backend-app/friend/persistent"
)

type ServiceConfig struct {
	Persistent repo.Repository
}
