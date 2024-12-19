package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strings"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

const (
	nemsURL         = "https://www.nems.emcsg.com/nems-prices"
	chromedriverURL = "https://googlechromelabs.github.io/chrome-for-testing/known-good-versions-with-downloads.json"
)

type Parameters struct {
	Parameters map[string]string `json:"Parameters"`
	Filepath   string            `json:"filepath"`
}

// func checkChromeVersion() string {

// 	version, err := chrome.GetAttribute()
// 	if err != nil {
// 		log.Fatal("Error getting the current Chrome version:", err)
// 	}
// 	fmt.Printf("System's Chrome version: %s\n", version)

// 	return version

// }

func getChromeVersion() (string, error) {
	// Command to get Chrome version on Windows
	cmd := exec.Command("powershell", "-Command", "(Get-Item 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe').VersionInfo.ProductVersion")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error getting Chrome version: %v", err)
	}

	// Trim whitespace and convert to string
	version := strings.TrimSpace(string(output))

	return version, nil
}

type DownloadInfo struct {
	Platform string `json:"platform"`
	Url      string `json:"url"`
	// Sha256   string `json:"sha256"`
}

// type DownloadType struct {
// 	Chrome       []DownloadInfo `json:"chrome"`
// 	Chromedriver []DownloadInfo `json:"chrome"`
// }

type VersionEntry struct {
	Version   string `json:"version"`
	Revision  string `json:"revision"`
	Downloads struct {
		Chrome       []DownloadInfo `json:"chrome"`
		Chromedriver []DownloadInfo `json:"chromedriver"`
	} `json:"downloads"`
}

type KnownGoodVersions struct {
	Timestamp string         `json:"timestamp"`
	Versions  []VersionEntry `json:"versions"`
}

func main() {

	chromeversion, err := getChromeVersion()

	// if err != nil {
	// 	panic(err)
	// 	// fmt.Println("No Chrome Version Found", err)
	// 	// return
	// 	os.Exit(1)
	// } else {
	// 	fmt.Println("Chrome Version : ", chromeversion)
	// }

	response, err := http.Get(chromedriverURL)
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
	defer response.Body.Close()

	// var data map[string]map[string]interface{}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal("Error reading response body:", err)
	}

	fmt.Println(reflect.TypeOf(body))

	var data1 KnownGoodVersions

	if err := json.Unmarshal(body, &data1); err != nil {
		log.Fatal("Error Unmarshalling JSON:", err)
	}

	// Finds Chrome Major Version
	// in 131.0.6778.140, returns 131.0.6778
	r, err := regexp.Compile(`([\d\.]+)\.\d+$`)
	if err != nil {
		log.Fatal("Regex Compilation Error: Unable to compile Chrome Major Version Check")
	}

	chromeMajorVersion := r.FindStringSubmatch(chromeversion)

	var downloadlink string

	for _, ver := range data1.Versions {

		// fmt.Println(a.Version)
		apiVersion := r.FindStringSubmatch(ver.Version)

		if chromeMajorVersion[1] == apiVersion[1] {
			for _, item := range ver.Downloads.Chromedriver {
				if item.Platform == "win64" {
					downloadlink = item.Url
					fmt.Println("Success, ", downloadlink)

					break
				}
			}
		}

		if downloadlink != "" {
			break
		}
	}

	if downloadlink == "" {
		log.Fatal("Chromedriver Link not found for this version of Chrome", chromeversion)
	}

	// Downloading zip file
	resp, err := http.Get(downloadlink)
	if err != nil {
		log.Fatal("Download Failure:", err)
	}

	// Write the response body to a temporary zip file
	tempZipFile, err := os.CreateTemp("", "chromedriver-*.zip")
	if err != nil {
		log.Fatal("Failed to create temporary file:", err)
	}

	if _, err := io.Copy(tempZipFile, resp.Body); err != nil {
		log.Fatal("Failed to write to temporary file:", err)
	}

	archive, err := zip.OpenReader(tempZipFile.Name())
	if err != nil {
		log.Fatal("Fail to open zipfile:", err)
	}

	// Extract the chromedriver.exe from the zip file
	for _, f := range archive.File {
		if f.Name == "chromedriver-win64/chromedriver.exe" {
			rc, err := f.Open()
			if err != nil {
				log.Fatal("Failed to open chromedriver.exe in zipfile:", err)
			}
			defer rc.Close()

			localFile, err := os.Create("./chromedriver.exe")
			if err != nil {
				log.Fatal("Failed to create local file:", err)
			}
			defer localFile.Close()

			if _, err := io.Copy(localFile, rc); err != nil {
				log.Fatal("Failed to write to local file:", err)
			}

			fmt.Println("Successfully replaced local chromedriver.exe")
		}
	}
	archive.Close()
	os.Remove(tempZipFile.Name())
	resp.Body.Close()
	// fmt.Println("Chrome Download at: ", found.Downloads.Chromedriver)

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
	resp, err = client.Do(req)
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
