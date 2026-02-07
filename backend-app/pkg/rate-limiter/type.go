package ratelimiter

type Config struct {
	BucketSize     int
	RefillInterval int // in minute
}
