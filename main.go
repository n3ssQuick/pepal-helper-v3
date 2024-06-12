package main

import (
	"context"
	"fmt"
	"helper/v3/controllers"
	"helper/v3/models"
	"net/http"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func addRoutes(api huma.API) {
	// Get Cookie
	huma.Register(api, huma.Operation{
		OperationID: "login",
		Method:      http.MethodPost,
		Path:        "/login",
		Summary:     "Login",
		Description: "Login and get User cookie",
	}, func(ctx context.Context, input *struct {
		Body struct {
			Username string `path:"username" maxLength:"30" example:"myusername" doc:"Username"`
			Password string `path:"password" example:"mypassword" doc:"Password"`
		}
	}) (*models.LoginOutput, error) {
		resp := &models.LoginOutput{}
		cookie, err := controllers.Login(input.Body.Username, input.Body.Password)
		resp.Body.Cookie = cookie
		return resp, err

	})

	// Get Course IDs
	huma.Register(api, huma.Operation{
		OperationID: "getCourseIDs",
		Method:      http.MethodPost,
		Path:        "/getCourseIDs",
		Summary:     "Get Course IDs",
		Description: "Get Course IDs for the day",
	}, func(ctx context.Context, input *struct {
		Body struct {
			Cookie string `json:"cookie" example:"yourcookie" doc:"Cookie"`
		}
	}) (*models.CourseIDsOutput, error) {
		resp := &models.CourseIDsOutput{}
		courses, err := controllers.GetCourseIDs(input.Body.Cookie)
		if err != nil {
			return nil, err
		}
		resp.Body.Courses = courses
		return resp, nil
	})

	// Get Attendance Status
	huma.Register(api, huma.Operation{
		OperationID: "getAttendanceStatus",
		Method:      http.MethodPost,
		Path:        "/getAttendanceStatus",
		Summary:     "Get Attendance Status",
		Description: "Get the attendance status for a course",
	}, func(ctx context.Context, input *struct {
		Body struct {
			Cookie   string `json:"cookie" example:"yourcookie" doc:"Cookie"`
			CourseID string `json:"courseID" example:"2275021" doc:"Course ID"`
		}
	}) (*models.AttendanceStatusOutput, error) {
		resp := &models.AttendanceStatusOutput{}
		status, err := controllers.GetAttendanceStatus(input.Body.Cookie, input.Body.CourseID)
		if err != nil {
			return nil, err
		}
		resp.Body.Status = status
		return resp, nil
	})

	// Set Presence
	huma.Register(api, huma.Operation{
		OperationID: "setPresence",
		Method:      http.MethodPost,
		Path:        "/setPresence",
		Summary:     "Set Presence",
		Description: "Mark presence for a course",
	}, func(ctx context.Context, input *struct {
		Body struct {
			Cookie   string `json:"cookie" example:"yourcookie" doc:"Cookie"`
			CourseID string `json:"courseID" example:"2275021" doc:"Course ID"`
		}
	}) (*models.GenericOutput, error) {
		err := controllers.SetPresence(input.Body.Cookie, input.Body.CourseID)
		if err != nil {
			return nil, err
		}
		return &models.GenericOutput{
			Body: struct {
				Message string `json:"message"`
			}{
				Message: "Presence marked successfully",
			},
		}, nil
	})

	// Get Calendar
	huma.Register(api, huma.Operation{
		OperationID: "fetchCalendar",
		Method:      http.MethodPost,
		Path:        "/fetchCalendar",
		Summary:     "Fetch Calendar",
		Description: "Fetch the calendar and return the schedule for the week",
	}, func(ctx context.Context, input *struct {
		Body struct {
			CalUUID string `json:"calUUID" example:"wilson_inessa" doc:"Calendar UUID"`
		}
	}) (*models.CalendarOutput, error) {
		resp := &models.CalendarOutput{}
		events, err := controllers.FetchAndParseCalendar(input.Body.CalUUID)
		if err != nil {
			return nil, err
		}
		resp.Body.Schedule = events
		return resp, nil
	})
}

func main() {
	router := chi.NewMux()
	config := huma.DefaultConfig("Certificate API", "1.0.0")
	api := humachi.New(router, config)
	addRoutes(api)

	// Start API
	err := http.ListenAndServe("0.0.0.0:8888", router)
	if err != nil {
		log.Error().Err(err)
	}
}

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "02/01/2006 15:04:05"})
	fmt.Println("Server started on port 8888")
}
