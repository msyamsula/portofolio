package url

import (
	"github.com/msyamsula/portofolio/domain/url/repository"
)

type Dependencies struct {
	Repo *repository.Repository
	Host string

	// hash properties
	Length        int64
	CharacterPool string
}

type Service struct {
	repo *repository.Repository
	host string

	// hash properties
	length        int64
	characterPool string
}

func New(dep Dependencies) *Service {
	return &Service{
		repo: dep.Repo,
		// hasher: dep.Hasher,
		host:          dep.Host,
		length:        dep.Length,
		characterPool: dep.CharacterPool,
	}
}
