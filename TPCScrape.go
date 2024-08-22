package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

const (
	nemsURL = "https://www.nems.emcsg.com/nems-prices"
)

type Parameters struct {
	Parameters map[string]string `json:"Parameters"`
	Filepath   string            `json:"filepath"`
}

func main() {
	// Set up Chrome options
	caps := selenium.Capabilities{"browserName": "chrome"}
	chromeCaps := chrome.Capabilities{
		Args: []string{
			"--headless",
		},
	}
	caps.AddChrome(chromeCaps)

	// Start a Selenium WebDriver server instance
	service, err := selenium.NewChromeDriverService("./chromedriver.exe", 9515)
	if err != nil {
		log.Fatal("Error starting the ChromeDriver server:", err)
	}
	defer service.Stop()

	// Connect to the WebDriver instance
	driver, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", 9515))
	if err != nil {
		log.Fatal("Error connecting to the WebDriver:", err)
	}
	defer driver.Quit()

	// Navigate to the URL
	if err := driver.Get(nemsURL); err != nil {
		log.Fatal("Error navigating to the URL:", err)
	}

	// Get cookies
	seleniumCookies, err := driver.GetCookies()
	if err != nil {
		log.Fatal("Error getting cookies:", err)
	}
	cookies := make([]*http.Cookie, len(seleniumCookies))
	for i, cookie := range seleniumCookies {
		cookies[i] = &http.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		}
	}

	// Find the form element
	formElem, err := driver.FindElement(selenium.ByXPATH, "//form[@action='/api/sitecore/DataSync/DataDownload']")
	if err != nil {
		log.Fatal("Error finding form element:", err)
	}

	// Load parameters from JSON file
	file, err := os.Open("Parameters.json")
	if err != nil {
		log.Fatal("Error opening Parameters.json:", err)
	}
	defer file.Close()

	var params Parameters
	if err := json.NewDecoder(file).Decode(&params); err != nil {
		log.Fatal("Error decoding JSON:", err)
	}

	// Get form action URL
	formActionURL, err := formElem.GetAttribute("action")
	if err != nil {
		log.Fatal("Error getting form action URL:", err)
	}

	// Construct full URL
	fullURL, err := url.Parse(formActionURL)
	if err != nil {
		log.Fatal("Error parsing form action URL:", err)
	}
	query := fullURL.Query()
	for key, value := range params.Parameters {
		query.Set(key, value)
	}
	fullURL.RawQuery = query.Encode()

	// Get user agent
	userAgent, err := driver.ExecuteScript("return navigator.userAgent;", nil)
	if err != nil {
		log.Fatal("Error getting user agent:", err)
	}

	// Create HTTP client and request
	client := &http.Client{}
	req, err := http.NewRequest("GET", fullURL.String(), nil)
	if err != nil {
		log.Fatal("Error creating request:", err)
	}

	// Set headers and cookies
	req.Header.Set("User-Agent", userAgent.(string))
	req.Header.Set("Referer", fullURL.String())
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error sending request:", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Unexpected status code: %d", resp.StatusCode)
	}

	// Write response to file
	file, err = os.Create(params.Filepath)
	if err != nil {
		log.Fatal("Error creating file:", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Fatal("Error writing to file:", err)
	}

	fmt.Println("Data successfully downloaded and saved.")
}
