package models

import "time"

type BlackListedToken struct {
	JTI       string    `json:"jti"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}
