package controllers

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"helper/v3/models"

	"golang.org/x/net/html"

	"github.com/joho/godotenv"
)

// ExtractCourseIDs parses the HTML content and extracts course IDs, names, and periods.
func ExtractCourseIDs(htmlContent string) ([]models.Course, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	var courses []models.Course
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			var course models.Course
			var isCourseRow bool
			tdIndex := 0
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && c.Data == "td" {
					tdIndex++
					for td := c.FirstChild; td != nil; td = td.NextSibling {
						if td.Type == html.TextNode {
							if tdIndex == 1 && strings.Contains(td.Data, ":") {
								isCourseRow = true
								period := determinePeriod(td.Data)
								course.Period = period
							} else if tdIndex == 2 {
								course.Name = strings.TrimSpace(td.Data)
							}
						} else if td.Type == html.ElementNode && td.Data == "a" {
							for _, attr := range td.Attr {
								if attr.Key == "href" && strings.Contains(attr.Val, "/presences/s/") {
									parts := strings.Split(attr.Val, "/")
									if len(parts) > 3 {
										course.ID = parts[3]
									}
								}
							}
						}
					}
				}
			}
			if isCourseRow && course.ID != "" {
				courses = append(courses, course)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	if len(courses) == 0 {
		return nil, errors.New("no course IDs found")
	}

	return courses, nil
}

// determinePeriod determines whether the course is in the morning or afternoon.
func determinePeriod(timeRange string) string {
	times := strings.Split(timeRange, "-")
	if len(times) != 2 {
		return ""
	}

	startTime, err := time.Parse("15:04", strings.TrimSpace(times[0]))
	if err != nil {
		return ""
	}

	if startTime.Hour() < 12 {
		return "Matin"
	} else {
		return "Après-midi"
	}
}

func GetCourseIDs(cookie string) ([]models.Course, error) {
	godotenv.Load()
	apiURL := os.Getenv("PEPAL_BASE_URL") + "presences"

	// Create an HTTP client
	client := &http.Client{}

	// Create the GET request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	// Set the headers for the GET request
	req.Header.Set("Host", "www.pepal.eu")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", "sdv="+cookie)

	// Send the GET request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	bodyString := string(bodyBytes)

	// Check if the user is not logged in
	if strings.Contains(bodyString, "form class=\"login-form\"") {
		return nil, errors.New("user not logged in")
	}

	// Extract course IDs
	return ExtractCourseIDs(bodyString)
}

func GetAttendanceStatus(cookie, courseID string) (string, error) {
	godotenv.Load()

	// Verify if the course ID is part of the day's courses
	courses, err := GetCourseIDs(cookie)
	if err != nil {
		return "", err
	}

	validCourse := false
	for _, course := range courses {
		if course.ID == courseID {
			validCourse = true
			break
		}
	}

	if !validCourse {
		return "", errors.New("invalid course ID for the current day")
	}

	// Load the attendance page for the course
	apiURL := os.Getenv("PEPAL_BASE_URL") + "presences/s/" + courseID

	// Create an HTTP client
	client := &http.Client{}

	// Create the GET request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}

	// Set the headers for the GET request
	req.Header.Set("Host", "www.pepal.eu")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", "sdv="+cookie)

	// Send the GET request
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	bodyString := string(bodyBytes)

	// Check if the user is not logged in
	doc, err := html.Parse(strings.NewReader(bodyString))
	if err != nil {
		return "", err
	}

	// Extract the attendance status
	var status string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			for _, attr := range n.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "panel-body") {
					textContent := getTextContent(n)
					if strings.Contains(textContent, "L'appel n'est pas encore ouvert") {
						status = "Closed"
					} else if strings.Contains(textContent, "L'appel est clôturé") {
						if strings.Contains(textContent, "Vous avez été noté présent") {
							status = "Present"
						} else {
							status = "Closed"
						}
					} else if strings.Contains(textContent, "Valider la présence en retard") {
						status = "Late"
					} else if strings.Contains(textContent, "Valider la présence") {
						status = "Open"
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	if status == "" {
		return "", errors.New("unable to determine attendance status")
	}

	return status, nil
}

// getTextContent retrieves the concatenated text content of a node.
func getTextContent(n *html.Node) string {
	var textContent string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			textContent += n.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return textContent
}

func SetPresence(cookie, courseID string) error {
	godotenv.Load()

	// Call GetAttendanceStatus to check if the attendance is open
	status, err := GetAttendanceStatus(cookie, courseID)
	if err != nil {
		return err
	}
	fmt.Print(status)

	if status != "Open" {
		log.Printf("Cannot set presence: %s", status)
		return errors.New("cannot set presence: " + status)
	}

	// Set the presence
	postURL := os.Getenv("PEPAL_BASE_URL") + "student/upload.php"
	data := url.Values{}
	data.Set("act", "set_present")
	data.Set("seance_pk", courseID)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("POST", postURL, strings.NewReader(data.Encode()))
	if err != nil {
		log.Printf("Error creating POST request for setting presence: %v", err)
		return err
	}

	// Define the headers for the POST request
	req.Header.Set("Host", "www.pepal.eu")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", "sdv="+cookie)
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending POST request for setting presence: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to set presence: %v", resp.Status)
		return errors.New("failed to set presence")
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body for setting presence: %v", err)
		return err
	}

	// Check if the response is compressed
	var bodyString string
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		reader, err := gzip.NewReader(bytes.NewReader(bodyBytes))
		if err != nil {
			log.Printf("Error creating gzip reader: %v", err)
			return err
		}
		defer reader.Close()
		unzippedBodyBytes, err := io.ReadAll(reader)
		if err != nil {
			log.Printf("Error reading unzipped response body: %v", err)
			return err
		}
		bodyString = string(unzippedBodyBytes)
	} else {
		bodyString = string(bodyBytes)
	}

	if !strings.Contains(bodyString, "location.reload();") {
		log.Println("Presence not marked successfully")
		return errors.New("presence not marked successfully")
	}

	return nil
}
