package models

type LoginOutput struct {
	Body struct {
		Cookie string `json:"cookie"`
	}
}
