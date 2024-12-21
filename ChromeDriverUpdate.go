package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type DownloadInfo struct {
	Platform string `json:"platform"`
	Url      string `json:"url"`
}

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

const (
	chromedriverURL = "https://googlechromelabs.github.io/chrome-for-testing/known-good-versions-with-downloads.json"
)

func UpdateChromeDriver() {
	// URL that returns pure JSON (even if incorrectly labeled as HTML)

	chromeversion, err := getChromeVersion()
	if err != nil {
		panic(err)
		// fmt.Println("No Chrome Version Found", err)
		// return
	} else {
		fmt.Println("Chrome Version : ", chromeversion)
	}

	response, err := http.Get(chromedriverURL)
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal("Error reading response body:", err)
	}

	var ChromeData KnownGoodVersions
	err = json.Unmarshal(body, &ChromeData)
	if err != nil {
		log.Fatal("Error unmarshalling JSON:", err)
	}

	// Status Check (200)
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Unexpected HTTP status: %d %s", resp.StatusCode, resp.Status)
	}

	// Step 2: Decode the JSON into a struct or a generic interface
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Fatal("Error decoding JSON:", err)
	}

	// At this point, 'data' contains your parsed JSON. You can now access fields:
	fmt.Println("Parsed JSON data:", data)
}
