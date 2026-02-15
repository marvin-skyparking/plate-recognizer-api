package utils

import (
	"errors"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

var rrCounter uint64

// Most Docker setups use same internal port
var endpoints = []string{
	"http://plate-recognizer-1:8080",
	"http://plate-recognizer-2:8080",
	"http://plate-recognizer-3:8080",
	"http://plate-recognizer-4:8080",
}

var client = &http.Client{
	Timeout: 2 * time.Second,
}

func isHealthy(base string) bool {
	url := base + "/v1/plate-reader/"

	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// must not be server error
	if resp.StatusCode >= 500 {
		return false
	}

	// reject HTML responses (proxy/error page)
	ct := resp.Header.Get("Content-Type")
	if ct != "" && !strings.Contains(ct, "application/json") {
		return false
	}

	return true
}

func GetHealthyPlateReaderURL() (string, error) {
	total := len(endpoints)

	for i := 0; i < total; i++ {
		idx := int(atomic.AddUint64(&rrCounter, 1) % uint64(total))
		base := endpoints[idx]

		if isHealthy(base) {
			return base + "/v1/plate-reader/", nil
		}
	}

	return "", errors.New("no healthy plate-recognizer available")
}
