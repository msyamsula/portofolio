package redis

import (
	"fmt"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	db, mock := redismock.NewClientMock()
	r := &Redis{
		Client: db,
	}
	assert.NotNil(t, mock)
	fmt.Println(r)
}
