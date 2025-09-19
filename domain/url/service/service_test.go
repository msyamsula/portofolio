package url

import (
	"testing"

	"github.com/msyamsula/portofolio/domain/url/repository"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {

	dep := Dependencies{
		Repo:          &repository.Repository{},
		Host:          "",
		Length:        0,
		CharacterPool: "",
	}

	db := New(dep)
	assert.NotNil(t, db)

}
