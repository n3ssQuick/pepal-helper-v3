package models

type CalendarOutput struct {
	Body struct {
		Schedule []Event `json:"schedule"`
	}
}

type Event struct {
	Day       string `json:"day"`
	FullDay   bool   `json:"full_day"`
	Morning   bool   `json:"morning"`
	Afternoon bool   `json:"afternoon"`
	Remote    bool   `json:"remote"`
	Location  string `json:"location,omitempty"`
	Professor string `json:"professor"`
	Subject   string `json:"subject"`
}
