package persistent

type User struct {
	Username string `json:"username"`
	Id       int64  `json:"id"`
	Online   bool   `json:"online"`
	Unread   int64  `json:"unread"`
}

type PostgresConfig struct {
	Username string
	Password string
	DbName   string
	Host     string
	Port     string
}
