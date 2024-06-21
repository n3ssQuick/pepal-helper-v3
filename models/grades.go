package models

type Grade struct {
	Subject string `json:"subject"`
	Date    string `json:"date"`
	Grade   string `json:"grade"`
	Comment string `json:"comment,omitempty"`
	Course  string `json:"course"`
}

type GradesOutput struct {
	Body struct {
		Grades []Grade `json:"grades"`
	} `json:"body"`
}
