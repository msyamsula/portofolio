package cache

import "time"

type DynamoConfig struct {
	Ttl       time.Duration
	Region    string
	TableName string
}

type RedisConfig struct {
	Host string
	Port string
	Env  string
}
