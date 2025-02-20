package repository

import (
	"github.com/msyamsula/portofolio/database/postgres"
	"github.com/msyamsula/portofolio/database/redis"
)

type Dependencies struct {
	Persistence *postgres.Postgres
	Cache       *redis.Redis
}

type Repository struct {
	persistence *postgres.Postgres
	cache       *redis.Redis
}

func New(dep Dependencies) *Repository {
	return &Repository{
		persistence: dep.Persistence,
		cache:       dep.Cache,
	}
}
