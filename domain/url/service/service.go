package url

import (
	"github.com/msyamsula/portofolio/domain/hasher"
	"github.com/msyamsula/portofolio/domain/url/repository"
)

type Dependencies struct {
	Repo   *repository.Repository
	Hasher *hasher.Service
}

type Service struct {
	repo   *repository.Repository
	hasher *hasher.Service
}

func New(dep Dependencies) *Service {
	return &Service{
		repo:   dep.Repo,
		hasher: dep.Hasher,
	}
}
