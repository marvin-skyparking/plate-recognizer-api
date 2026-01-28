package utils

import (
	"errors"
	"net/http"
	"sync/atomic"
	"time"
)

var rrCounter uint64

var endpoints = []string{
	"http://plate-recognizer-1:8080",
	"http://plate-recognizer-2:8081",
}

func isHealthy(base string) bool {
	client := http.Client{
		Timeout: 2 * time.Second,
	}

	// Plate Recognizer does NOT have /health
	// Use HEAD or GET to plate-reader endpoint
	req, err := http.NewRequest(
		http.MethodHead,
		base+"/v1/plate-reader/",
		nil,
	)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// 200 / 401 / 405 all mean "service is alive"
	return resp.StatusCode < 500
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
