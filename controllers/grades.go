package controllers

import (
	"fmt"
	"helper/v3/models"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type Grade struct {
	Subject string `json:"subject"`
	Date    string `json:"date"`
	Grade   string `json:"grade"`
	Comment string `json:"comment,omitempty"`
}

// FetchGrades retrieves the grades from the Pepal grades page.
func FetchGrades(cookie string) ([]models.Grade, error) {
	godotenv.Load()
	apiURL := os.Getenv("PEPAL_BASE_URL") + "?my=notes"

	// Create an HTTP client
	client := &http.Client{}

	// Create the GET request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.Error().Err(err).Msg("Error creating request")
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set the request headers
	req.Header.Set("Cookie", "sdv="+cookie)

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check the HTTP status code
	if resp.StatusCode != http.StatusOK {
		log.Error().Str("status", resp.Status).Msg("Failed to load page")
		return nil, fmt.Errorf("failed to load page: %s", resp.Status)
	}

	// Use goquery to parse the HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading HTML")
		return nil, fmt.Errorf("error reading HTML: %v", err)
	}

	// Extract the grades
	var grades []models.Grade
	var currentCourse string
	var lastGrade *models.Grade

	doc.Find("table.table-bordered tbody tr").Each(func(i int, s *goquery.Selection) {
		var grade models.Grade

		// Detect courses
		if s.HasClass("warning") || s.HasClass("info") {
			currentCourse = strings.TrimSpace(s.Find("td").First().Text())
			return
		}

		s.Children().Each(func(j int, td *goquery.Selection) {
			text := strings.TrimSpace(td.Text())
			switch j {
			case 0:
				grade.Subject = strings.TrimSpace(strings.Replace(text, "PUBLIE", "", -1))
			case 2:
				grade.Date = text
			case 3:
				grade.Grade = text
			}
		})

		// Add the grade if the required fields are present
		if grade.Subject != "" && grade.Date != "" && grade.Grade != "" {
			grade.Course = currentCourse
			grades = append(grades, grade)
			lastGrade = &grades[len(grades)-1]
		} else if lastGrade != nil && grade.Subject == "" && grade.Date == "" && grade.Grade == "" {
			// Add the comment to the last added grade
			lastGrade.Comment = strings.TrimSpace(s.Text())
		}
	})

	log.Info().Int("gradeCount", len(grades)).Msg("Fetched grades successfully")
	return grades, nil
}
