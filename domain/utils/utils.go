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

func CreatePrepareQuery(s string) string {
	var b strings.Builder
	n := len(s)
	i := 0
	for i < n {
		j := i
		for j < n && s[j] != ':' {
			b.WriteByte(s[j])
			j++
		}
		i = j

		if i == n {
			break
		}

		i++
		j = i
		for j < n && s[j] != ',' && s[j] != ')' && s[j] != ' ' {
			j++
		}
		b.WriteByte('?')
		i = j
	}
	return b.String()
}
