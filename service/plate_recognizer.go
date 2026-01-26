package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"plate-recognizer-api/utils"
	"time"
)

var rrCounter uint64

type Response struct {
	Results []struct {
		Plate string  `json:"plate"`
		Score float64 `json:"score"`
	} `json:"results"`
}

func Recognize(token, imagePath, mmc, cameraID string, transactionNo string) (string, float64, error) {
	start := time.Now()

	// Log execution time
	defer func() {
		log.Println("⏱ PlateRecognizer duration:", time.Since(start))
	}()

	file, err := os.Open(imagePath)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add image file
	part, err := writer.CreateFormFile("upload", "image.jpg")
	if err != nil {
		return "", 0, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", 0, err
	}

	// Add extra fields
	timestamp := time.Now().Format(time.RFC3339)
	_ = writer.WriteField("timestamp", timestamp)
	_ = writer.WriteField("mmc", mmc)
	_ = writer.WriteField("camera_id", cameraID)

	_ = writer.Close()

	// Round-robin selection between two endpoints
	// endpoints := [2]string{
	// 	"http://plate-recognizer:8080/v1/plate-reader/",
	// 	"http://plate-recognizer:8081/v1/plate-reader/",
	// }
	// endpoints := [2]string{
	// 	"http://localhost:8080/v1/plate-reader/",
	// 	"http://localhost:8081/v1/plate-reader/",
	// }
	// idx := atomic.AddUint64(&rrCounter, 1) % 2
	// url := endpoints[idx]

	// req, err := http.NewRequest(
	// 	http.MethodPost,
	// 	url,
	// 	&body,
	// )
	// if err != nil {
	// 	return "", 0, err
	// }

	// 1️⃣ get healthy endpoint
	url, err := utils.GetHealthyPlateReaderURL()
	if err != nil {
		return "", 0, err
	}

	log.Println("🚀 Sending request to:", url)

	// 2️⃣ create request (IMPORTANT)
	req, err := http.NewRequest(
		http.MethodPost,
		url,
		&body,
	)
	if err != nil {
		return "", 0, err
	}

	// 3️⃣ set headers AFTER request is created
	req.Header.Set("Authorization", "Token "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// ======================
	// 🔎 LOG REQUEST
	// ======================
	log.Println("🚀 PlateRecognizer REQUEST")
	log.Println("URL       :", req.URL.String())
	log.Println("MMC       :", mmc)
	log.Println("Camera ID :", cameraID)
	log.Println("Timestamp :", timestamp)
	log.Println("Body size :", body.Len(), "bytes")

	// 4️⃣ send request
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}

	// ======================
	// 🔎 LOG RESPONSE
	// ======================
	log.Println("📥 PlateRecognizer RESPONSE")
	log.Println("Status :", resp.Status)
	log.Println("Body   :", string(respBody))

	var result Response
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", 0, err
	}

	if len(result.Results) == 0 {
		return "", 0, fmt.Errorf("no plate detected")
	}

	return result.Results[0].Plate, result.Results[0].Score, nil
}
