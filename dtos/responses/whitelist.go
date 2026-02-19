package responses

import "time"

type WhitelistResponse struct {
	Users []WhitelistedUserResponse `json:"users"`
}

type WhitelistedUserResponse struct {
	ID        string    `json:"id"`
	Handle    string    `json:"handle"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}