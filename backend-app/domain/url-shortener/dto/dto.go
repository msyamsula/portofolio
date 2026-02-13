package dto

import "time"

// URLRecord represents the URL mapping stored in database
type URLRecord struct {
	ShortCode string    `db:"short" json:"short_code"`
	LongURL   string    `db:"long" json:"long_url"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
