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

type SettingsPatch struct {
	MultiUserEnabled *bool   `json:"multiUserEnabled"`
	CurrentUserId    *string `json:"currentUserId"`
	NumberFormat     *string `json:"numberFormat"`
}

type Icon struct {
	ID     string `json:"id"`
	Data   string `json:"data"`
	IsUsed bool   `json:"isUsed"`
}

type IconRequest struct {
	Data string `json:"data"`
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
