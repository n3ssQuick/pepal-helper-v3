package controllers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"helper/v3/models"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// FetchAndParseCalendar télécharge, lit et analyse le fichier .ics, et retourne les événements de la semaine en cours
func FetchAndParseCalendar(calUUID string) ([]models.Event, error) {
	err := FetchCalendar(calUUID)
	if err != nil {
		return nil, err
	}

	filePath := "assets/" + calUUID + ".ics"
	content, err := ReadCalendar(filePath)
	if err != nil {
		return nil, err
	}

	events, err := ParseCalendar(content)
	if err != nil {
		return nil, err
	}

	weeklyEvents := FilterWeeklyEvents(events)
	return weeklyEvents, nil
}

// FetchCalendar télécharge le fichier situé à l'URL formée avec le calUUID et le sauvegarde dans le dossier assets
func FetchCalendar(calUUID string) error {
	baseURL := "https://www.pepal.eu/ical_student/"
	url := baseURL + calUUID

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("erreur lors de la requête GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("échec de la requête: %s", resp.Status)
	}

	out, err := os.Create("assets/" + calUUID + ".ics")
	if err != nil {
		return fmt.Errorf("erreur lors de la création du fichier: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("erreur lors de l'écriture du fichier: %v", err)
	}

	log.Printf("Fichier %s.ics téléchargé avec succès dans le dossier assets\n", calUUID)
	return nil
}

// ReadCalendar lit le contenu du fichier .ics
func ReadCalendar(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("erreur lors de l'ouverture du fichier: %v", err)
	}
	defer file.Close()

	var content string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content += scanner.Text() + "\n"
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("erreur lors de la lecture du fichier: %v", err)
	}

	return content, nil
}

// ParseCalendar analyse le contenu du fichier .ics et retourne une liste d'événements
func ParseCalendar(content string) ([]models.Event, error) {
	var events []models.Event
	lines := strings.Split(content, "\n")
	var currentEvent models.Event
	var startDate time.Time
	var endDate time.Time

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "BEGIN:VEVENT") {
			currentEvent = models.Event{}
		} else if strings.HasPrefix(line, "SUMMARY:") {
			currentEvent.Subject = strings.TrimPrefix(line, "SUMMARY:")
		} else if strings.HasPrefix(line, "DTSTART:") {
			startDate, _ = time.Parse("20060102T150405", strings.TrimPrefix(line, "DTSTART:"))
		} else if strings.HasPrefix(line, "DTEND:") {
			endDate, _ = time.Parse("20060102T150405", strings.TrimPrefix(line, "DTEND:"))
		} else if strings.HasPrefix(line, "LOCATION:") {
			currentEvent.Location = strings.TrimPrefix(line, "LOCATION:")
			currentEvent.Remote = currentEvent.Location == ""
		} else if strings.HasPrefix(line, "PROF:") {
			currentEvent.Professor = strings.TrimPrefix(line, "PROF:")
		} else if strings.HasPrefix(line, "END:VEVENT") {
			if currentEvent.Subject == "" {
				currentEvent.Subject = "entreprise"
				currentEvent.FullDay = true
				currentEvent.Morning = false
				currentEvent.Afternoon = false
				currentEvent.Remote = false
				currentEvent.Professor = ""
			} else {
				currentEvent.Day = startDate.Format("2006-01-02")
				if startDate.Hour() < 12 {
					currentEvent.Morning = true
				} else {
					currentEvent.Afternoon = true
				}
				if startDate.Hour() == 9 && endDate.Hour() == 16 {
					currentEvent.FullDay = true
					currentEvent.Morning = false
					currentEvent.Afternoon = false
				}
			}
			events = append(events, currentEvent)
		}
	}

	return events, nil
}

// FilterWeeklyEvents filtre les événements pour ne garder que ceux de la semaine en cours
func FilterWeeklyEvents(events []models.Event) []models.Event {
	var weeklyEvents []models.Event
	now := time.Now()
	year, week := now.ISOWeek()
	for _, event := range events {
		eventTime, _ := time.Parse("2006-01-02", event.Day)
		eventYear, eventWeek := eventTime.ISOWeek()
		if eventYear == year && eventWeek == week {
			weeklyEvents = append(weeklyEvents, event)
		}
	}
	return weeklyEvents
}

// CalendarToJSON convertit une liste d'événements en JSON
func CalendarToJSON(events []models.Event) (string, error) {
	jsonData, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return "", fmt.Errorf("erreur lors de la conversion en JSON: %v", err)
	}
	return string(jsonData), nil
}
