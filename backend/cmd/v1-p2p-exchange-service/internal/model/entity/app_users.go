package entity

import "time"

type AppUser struct {
	ID            int64     `db:"id"`
	Username      string    `db:"username"`
	PasswordHash  string    `db:"password_hash"`
	Email         *string   `db:"email"`
	ExpoPushToken *string   `db:"expo_push_token"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}
