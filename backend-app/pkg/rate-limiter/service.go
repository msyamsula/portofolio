package ratelimiter

import (
	"fmt"
	"time"
)

type service struct {
	_bucket int

	bucketSize     int
	refillInterval int // in minute
}

func (s *service) refill() {
	for {
		time.Sleep(time.Duration(s.refillInterval) * time.Minute)
		s._bucket = s.bucketSize
	}
}

func (s *service) IsAllowed() bool {
	return s._bucket > 0
}

func (s *service) Use() {
	s._bucket--
}

func (s *service) Run() {
	go s.refill()
	go func() {
		for {
			time.Sleep(1 * time.Second)
			fmt.Println(s._bucket, s.IsAllowed())
		}
	}()
}
