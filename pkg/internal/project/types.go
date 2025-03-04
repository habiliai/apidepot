package project

type GetProjectsInput struct {
	Name    *string `json:"name" form:"name"`
	Page    int     `json:"page" form:"page"`
	PerPage int     `json:"per_page" form:"per_page"`
}
