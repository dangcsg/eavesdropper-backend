package requests

type NewUser struct {
	Email           string
	PrefersDarkMode bool
}

type UpdateUser struct {
	Handle          *string `json:"handle,omitempty"`
	FirstName       *string `json:"firstName,omitempty"`
	LastName        *string `json:"lastName,omitempty"`
	PrefersDarkMode *bool   `json:"prefersDarkMode,omitempty"`
}
