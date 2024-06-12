package models

type Course struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Period string `json:"period"`
}

type CourseIDsOutput struct {
	Body struct {
		Courses []Course `json:"courses"`
	}
}

type AttendanceStatusOutput struct {
	Body struct {
		Status string `json:"status"`
	} `json:"body"`
}

type GenericOutput struct {
	Body struct {
		Message string `json:"message"`
	} `json:"body"`
}
