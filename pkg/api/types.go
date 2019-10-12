package api

//Claims -  custom claims
type Claims struct {
	Email    string `json:"email"`
	Verified bool   `json:"email_verified"`
}

type GitHubUser struct {
	Login string `json:"login"`
}
