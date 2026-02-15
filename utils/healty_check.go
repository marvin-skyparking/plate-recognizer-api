// package utils

// import (
// 	"errors"
// 	"net/http"
// 	"sync/atomic"
// 	"time"
// )

// var rrCounter uint64

// var endpoints = []string{
// 	"http://plate-recognizer-1:8080",
// 	"http://plate-recognizer-2:8081",
// 	"http://plate-recognizer-3:8082",
// 	"http://plate-recognizer-4:8083",
// }

// func isHealthy(base string) bool {
// 	client := http.Client{
// 		Timeout: 2 * time.Second,
// 	}

// 	// Plate Recognizer does NOT have /health
// 	// Use HEAD or GET to plate-reader endpoint
// 	req, err := http.NewRequest(
// 		http.MethodHead,
// 		base+"/v1/plate-reader/",
// 		nil,
// 	)
// 	if err != nil {
// 		return false
// 	}

// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return false
// 	}
// 	defer resp.Body.Close()

// 	// 200 / 401 / 405 all mean "service is alive"
// 	return resp.StatusCode < 500
// }

// func GetHealthyPlateReaderURL() (string, error) {
// 	total := len(endpoints)

// 	for i := 0; i < total; i++ {
// 		idx := int(atomic.AddUint64(&rrCounter, 1) % uint64(total))
// 		base := endpoints[idx]

// 		if isHealthy(base) {
// 			return base + "/v1/plate-reader/", nil
// 		}
// 	}

// 	return "", errors.New("no healthy plate-recognizer available")
// }

package utils

import (
	"errors"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

var rrCounter uint64

// If running in Docker, usually internal ports are SAME.
// Change only if your containers truly expose different ports.
var endpoints = []string{
	"http://plate-recognizer-1:8080",
	"http://plate-recognizer-2:8081",
	"http://plate-recognizer-3:8082",
	"http://plate-recognizer-4:8083",
}

var httpClient = &http.Client{
	Timeout: 3 * time.Second,
}

// ==============================
// HEALTH CHECK
// ==============================

func isHealthy(base string) bool {
	url := base + "/v1/plate-reader/"

	// Try HEAD first (lightweight)
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return false
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()

	// Some servers don't support HEAD → retry with GET
	if resp.StatusCode == http.StatusMethodNotAllowed {
		req2, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return false
		}

		resp2, err := httpClient.Do(req2)
		if err != nil {
			return false
		}
		defer resp2.Body.Close()

		return validAPIResponse(resp2)
	}

	return validAPIResponse(resp)
}

func validAPIResponse(resp *http.Response) bool {
	// Reject server errors
	if resp.StatusCode >= 500 {
		return false
	}

	// Reject HTML pages (common reverse proxy / error page)
	ct := resp.Header.Get("Content-Type")
	if ct != "" && !strings.Contains(ct, "application/json") {
		return false
	}

	return true
}

// ==============================
// ROUND ROBIN
// ==============================

func GetHealthyPlateReaderURL() (string, error) {
	total := len(endpoints)
	if total == 0 {
		return "", errors.New("no endpoints configured")
	}

	for i := 0; i < total; i++ {
		idx := int(atomic.AddUint64(&rrCounter, 1) % uint64(total))
		base := endpoints[idx]

		if isHealthy(base) {
			return base + "/v1/plate-reader/", nil
		}
	}

	return "", errors.New("no healthy plate-recognizer available")
}
