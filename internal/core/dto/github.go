package dto

type Owner struct {
	Login string `json:"login"`
}

type LatestCreatedRepo struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	FullName     string `json:"full_name"`
	Owner        *Owner `json:"owner"`
	HTMLURL      string `json:"html_url"`
	LanguagesURL string `json:"languages_url"`
	URL          string `json:"url"`
	Description  string `json:"description"`
}
