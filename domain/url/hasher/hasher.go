package hasher

import (
	"crypto/rand"
	"math/big"
	"strings"
)

type Service struct {
	length int64
	word   string
}

type Config struct {
	Length int64
	Word   string
}

func New(cfg Config) *Service {
	return &Service{
		length: cfg.Length,
		word:   cfg.Word,
	}
}

func (s *Service) Hash() string {
	var result strings.Builder
	limit := big.NewInt(int64(len(s.word)))
	for i := 0; i < int(s.length); i++ {
		randomIdx, _ := rand.Int(rand.Reader, limit)
		idx := randomIdx.Int64()
		idx %= s.length
		result.WriteByte(s.word[idx])
	}

	return result.String()
}
