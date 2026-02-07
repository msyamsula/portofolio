package ratelimiter

type RateLimiter interface {
	Run()
	Use()
	IsAllowed() bool
}

func NewRateLimiter(config Config) RateLimiter {
	return &service{
		bucketSize:     config.BucketSize,
		refillInterval: config.RefillInterval,
		_bucket:        config.BucketSize,
	}
}
