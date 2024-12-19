package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// type DownloadInfo struct {
// 	Platform string `json:"platform"`
// 	Url      string `json:"url"`
// }

// type VersionEntry struct {
// 	Version   string `json:"version"`
// 	Revision  string `json:"revision"`
// 	Downloads struct {
// 		Chrome       []DownloadInfo `json:"chrome"`
// 		Chromedriver []DownloadInfo `json:"chromedriver"`
// 	} `json:"downloads"`
// }

// type KnownGoodVersions struct {
// 	Timestamp string         `json:"timestamp"`
// 	Versions  []VersionEntry `json:"versions"`
// }

func sub() {
	// URL that returns pure JSON (even if incorrectly labeled as HTML)
	url := "https://googlechromelabs.github.io/chrome-for-testing/latest-patch-versions-per-build-with-downloads.json"

	resp, err := http.Get(url)

	// URL Fetch
	if err != nil {
		log.Fatal("Error fetching the URL:", err)
	}
	defer resp.Body.Close()

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
