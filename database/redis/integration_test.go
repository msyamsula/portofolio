package redis

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewIntegration(t *testing.T) {
	db := New(Config{
		Host:     "127.0.0.1",
		Port:     "6379",
		Password: "admin",
		Ttl:      60 * time.Second,
	})
	assert.NotNil(t, db.Client)
	fmt.Println(db.Client)

}
