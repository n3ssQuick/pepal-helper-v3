package controllers

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func Login(username, password string) (string, error) {
	godotenv.Load()
	apiURL := os.Getenv("PEPAL_BASE_URL") + "include/php/ident.php"

	// Create a cookie jar to manage cookies
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Printf("Error creating cookie jar: %v", err)
		return "", err
	}

	// Create an HTTP client with the cookie jar
	client := &http.Client{
		Jar: jar,
	}

	// Create the POST data
	data := url.Values{}
	data.Set("login", username)
	data.Set("pass", password)

	// Create the POST request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Printf("Error creating POST request: %v", err)
		return "", err
	}

	// Set the headers for the POST request
	req.Header.Set("Host", "www.pepal.eu")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the POST request
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending POST request: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return "", err
	}

	bodyString := string(body)

	// Check for the "Accès refusé !" message
	if strings.Contains(bodyString, "Accès refusé !") {
		log.Println("Incorrect username or password")
		return "", errors.New("identifiant et/ou mot de passe incorrect(s)")
	}

	// Check for the "Connexion réussie" message
	if resp.StatusCode != http.StatusOK {
		log.Printf("Login failed with status: %v", resp.Status)
		return "", errors.New("login failed")
	}

	// Check for the "Connexion réussie" message
	for _, cookie := range jar.Cookies(req.URL) {
		if cookie.Name == "sdv" {
			log.Println("Login successful")
			return cookie.Value, nil
		}
	}

	log.Println("Cookie not found")
	return "", errors.New("cookie not found")
}
