package models

import "time"

type User struct {
	ID         int64
	City       string
	Created_at time.Time
}
