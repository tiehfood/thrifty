package main

type Flow struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Icon        string  `json:"icon"`
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Settings struct {
	MultiUserEnabled bool   `json:"multiUserEnabled"`
	CurrentUser      *User  `json:"currentUser"`
	NumberFormat     string `json:"numberFormat"`
}

// SettingsPatch uses pointers so absent fields are nil (not updated).
type SettingsPatch struct {
	MultiUserEnabled *bool   `json:"multiUserEnabled"`
	CurrentUserId    *string `json:"currentUserId"`
	NumberFormat     *string `json:"numberFormat"`
}

type HTTPResponse struct {
	Ok string `json:"ok"`
}

type HTTPError struct {
	Error string `json:"error"`
}

type Migration struct {
	Version    int
	Statements []string
}
