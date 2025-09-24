package shortener

import "time"

type URL struct {
	ID        string    `bson:"_id"`
	LongURL   string    `bson:"long_url"`
	CreatedAt time.Time `bson:"created_at"`
}