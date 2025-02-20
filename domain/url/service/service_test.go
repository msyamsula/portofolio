package url

import (
	"testing"

	"github.com/msyamsula/portofolio/domain/hasher"
	"github.com/msyamsula/portofolio/domain/url/repository"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {

	dep := Dependencies{
		Repo:   &repository.Repository{},
		Hasher: &hasher.Service{},
	}

	db := New(dep)
	assert.NotNil(t, db)

}
