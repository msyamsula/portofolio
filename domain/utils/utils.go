package utils

import (
	"math/rand"
	"strings"
	"time"
)

var (
	characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	randomizer = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func RandomName(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		idx := randomizer.Intn(len(characters))
		b.WriteByte(characters[idx])
	}

	return b.String()
}
