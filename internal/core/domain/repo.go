package domain

type ListRepoInput struct {
	Language            string `json:"language" validate:"omitempty"`
	License             string `json:"license" validate:"omitempty"`
	NameContains        string `json:"name_contains" validate:"omitempty"`
	DescriptionContains string `json:"description_contains" validate:"omitempty"`
	MinSize             int64  `json:"min_size" validate:"omitempty,min=1"`
	MaxSize             int64  `json:"max_size" validate:"omitempty,min=1"`
}

type ListRepoOutput struct {
	FullName    string         `json:"full_name"`
	Owner       string         `json:"owner"`
	Repository  string         `json:"repository"`
	License     string         `json:"license"`
	Description string         `json:"description"`
	Languages   map[string]int `json:"languages"`
}

func (l *ListRepoOutput) RepoSize() int64 {
	totalSize := int64(0)
	for _, currentLanguageSize := range l.Languages {
		totalSize += int64(currentLanguageSize)
	}
	return totalSize
}
