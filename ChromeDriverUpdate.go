package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
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

func AsyncChromeUpdate(ch chan error, wg *sync.WaitGroup) {

	defer wg.Done()

	err := UpdateChromeDriver()

	ch <- err // Send the result to the channel
}

func UpdateChromeDriver() error {
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
	if response.StatusCode != http.StatusOK {
		log.Fatalf("Unexpected HTTP status: %d %s", response.StatusCode, response.Status)
	}

	// Finds Chrome Major Version
	// in 131.0.6778.140, returns 131.0.6778
	r, err := regexp.Compile(`([\d\.]+)\.\d+$`)
	if err != nil {
		log.Fatal("Regex Compilation Error: Unable to compile Chrome Major Version Check")
	}

	chromeMajorVersion := r.FindStringSubmatch(chromeversion)

	var downloadlink string

	for _, ver := range ChromeData.Versions {

		// fmt.Println(a.Version)
		apiVersion := r.FindStringSubmatch(ver.Version)

		if chromeMajorVersion[1] == apiVersion[1] {
			for _, item := range ver.Downloads.Chromedriver {
				if item.Platform == "win64" {
					downloadlink = item.Url
					fmt.Println("Success, ", downloadlink)
				}
			}
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

	return nil
}
