package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"plate-recognizer-api/utils"
	"time"
)

var rrCounter uint64

type Response struct {
	Filename         string  `json:"filename"`
	ImageWidth       int     `json:"image_width"`
	ImageHeight      int     `json:"image_height"`
	PlatesFound      int     `json:"plates_found"`
	ProcessingTimeMS float64 `json:"processing_time_ms"`
	Plates           []Plate `json:"plates"`
}

type Plate struct {
	Text               string      `json:"text"`
	RawText            string      `json:"raw_text"`
	Confidence         float64     `json:"confidence"`
	DetectorConfidence float64     `json:"detector_confidence"`
	BoundingBox        BoundingBox `json:"bounding_box"`
}

type BoundingBox struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

func Recognize(
	LPR_API_KEY,
	imagePath,
	mmc,
	cameraID,
	transactionNo string,
) (string, float64, error) {

	start := time.Now()
	defer func() {
		log.Println("⏱ PlateRecognizer duration:", time.Since(start))
	}()

	file, err := os.Open(imagePath)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	// Detect MIME type
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", 0, err
	}

	contentType := http.DetectContentType(buffer[:n])

	// Reset reader
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", 0, err
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	header := make(textproto.MIMEHeader)
	header.Set(
		"Content-Disposition",
		fmt.Sprintf(
			`form-data; name="file"; filename="%s"`,
			filepath.Base(imagePath),
		),
	)
	header.Set("Content-Type", contentType)

	part, err := writer.CreatePart(header)
	if err != nil {
		return "", 0, err
	}

	if _, err := io.Copy(part, file); err != nil {
		return "", 0, err
	}

	if err := writer.Close(); err != nil {
		return "", 0, err
	}

	url, err := utils.GetHealthyPlateReaderURL()
	if err != nil {
		return "", 0, err
	}

	req, err := http.NewRequest(http.MethodPost, url, &body)
	if err != nil {
		return "", 0, err
	}

	req.Header.Set("X-API-Key", os.Getenv("LPR_API_KEY"))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	log.Println("========== REQUEST ==========")
	log.Println("URL          :", url)
	log.Println("Image        :", imagePath)
	log.Println("Image-Type   :", contentType)
	log.Println("Request-Type :", writer.FormDataContentType())
	log.Println("TOKEN :", os.Getenv("LPR_API_KEY"))
	log.Println("=============================")

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

	log.Println("========== RESPONSE ==========")
	log.Println("Status :", resp.Status)
	log.Println("Body   :", string(respBody))
	log.Println("==============================")

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("%s", string(respBody))
	}

	var result Response
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", 0, err
	}

	log.Printf("Parsed Response: %+v", result)
	log.Printf("plates_found=%d len(plates)=%d",
		result.PlatesFound,
		len(result.Plates),
	)

	if result.PlatesFound == 0 || len(result.Plates) == 0 {
		return "", 0, fmt.Errorf("no plate detected")
	}

	best := result.Plates[0]

	log.Printf(
		"✅ Plate=%s Confidence=%.4f DetectorConfidence=%.4f",
		best.Text,
		best.Confidence,
		best.DetectorConfidence,
	)

	return best.Text, best.Confidence, nil
}
