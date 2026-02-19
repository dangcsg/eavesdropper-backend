package requests

type AddToWhitelistRequest struct {
	Users []WhitelistUser `json:"users"`
}

type RemoveFromWhitelistRequest struct {
	Handles []string `json:"handles,omitempty"`
	Emails  []string `json:"emails,omitempty"`
}

type WhitelistUser struct {
	Handle string `json:"handle"`
	Email  string `json:"email"`
}