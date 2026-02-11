package dto

// URLRecord represents the URL mapping stored in database
type URLRecord struct {
	ShortCode string `db:"short_code" json:"short_code"`
	LongURL   string `db:"long_url" json:"long_url"`
	CreatedAt int64  `db:"created_at" json:"created_at"`
}
