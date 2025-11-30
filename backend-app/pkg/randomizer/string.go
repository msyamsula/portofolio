package randomizer

import (
	"crypto/rand"
	"math/big"
	"strings"
)

func (s *StringRandomizer) String() (string, error) {
	var short strings.Builder
	var err error
	for i := 0; i < s.size; i++ {
		var bidx *big.Int
		bidx, err = rand.Int(rand.Reader, big.NewInt(int64(len(s.characterPool))))
		if err != nil {
			return "", err
		}
		idx := int(bidx.Int64())
		short.WriteByte(s.characterPool[int(idx)])
	}

	return short.String(), nil
}
