package utils

import (
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

var rrCounter uint64

var endpoints = []string{
	"http://platerecognizer:8000",
}

func isHealthy(base string) bool {
	client := http.Client{
		Timeout: 2 * time.Second,
	}

	url := base + "/v1/recognize"

	fmt.Printf("[LPR HEALTH] checking: %s\n", url)

	req, err := http.NewRequest(
		http.MethodHead,
		url,
		nil,
	)
	if err != nil {
		fmt.Printf("[LPR HEALTH] request creation error: %v\n", err)
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[LPR HEALTH] HTTP error: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	fmt.Printf("[LPR HEALTH] status: %d\n", resp.StatusCode)

	return resp.StatusCode < 500
}

func GetHealthyPlateReaderURL() (string, error) {
	total := len(endpoints)

	for i := 0; i < total; i++ {
		idx := int(atomic.AddUint64(&rrCounter, 1) % uint64(total))
		base := endpoints[idx]

		if isHealthy(base) {
			return base + "/v1/recognize", nil
		}
	}

	return "", errors.New("no healthy plate-recognizer available")
}
