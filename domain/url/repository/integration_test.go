//go:build integration
// +build integration

package postgres

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewIntegration(t *testing.T) {
	db := New(Config{
		Username: "admin",
		Password: "admin",
		DbName:   "postgres",
		Host:     "127.0.0.1",
		Port:     "5432",
	})
	assert.NotNil(t, db)

	ctx := context.Background()
	a, b := db.GetLongUrl(ctx, "abc")
	fmt.Println(a)
	fmt.Println(b)
	assert.NoError(t, db.SetShortUrl(ctx, "mantap", "sekali"))

}
