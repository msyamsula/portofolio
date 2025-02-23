package url

import (
	"github.com/msyamsula/portofolio/domain/url/hasher"
	"github.com/msyamsula/portofolio/domain/url/repository"
)

type Dependencies struct {
	Repo   *repository.Repository
	Hasher *hasher.Service
	Host   string
}

type Service struct {
	repo   *repository.Repository
	hasher *hasher.Service
	host   string
}

func New(dep Dependencies) *Service {
	return &Service{
		repo:   dep.Repo,
		hasher: dep.Hasher,
		host:   dep.Host,
	}
}
