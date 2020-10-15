package api

//Claims -  custom claims
type Claims struct {
	Email    string `json:"email"`
	Verified bool   `json:"email_verified"`
	Name     string `json:"name"`
}

type GitHubUser struct {
	Login   string `json:"login"`
	Message string `json:"message"`
}
