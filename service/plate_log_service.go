package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"plate-recognizer-api/internal/minio"
	"plate-recognizer-api/model"

	"gorm.io/gorm"
)

type FinalResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Code    string      `json:"code"`
}

func RecognizeAndSavePlateLog(
	db *gorm.DB,
	token string,
	imagePath string,
	locationCode string,
	cameraID string,
	mmc string,
) (*FinalResponse, error) {

	// ✅ Call directly (NO import service)
	plate, score, err := Recognize(
		token,
		imagePath,
		mmc,
		cameraID,
	)
	if err != nil {
		return nil, err
	}

	plate = strings.ToUpper(plate)

	finalResp := FinalResponse{
		Status:  200,
		Message: "plate recognized successfully",
		Code:    "SUCCESS",
		Data: map[string]interface{}{
			"plate": plate,
			"score": score,
		},
	}

	// Prepare request metadata
	requestMeta := map[string]string{
		"location_code": locationCode,
		"camera_id":     cameraID,
		"mmc":           mmc,
	}

	// Try to upload image to MinIO (optional)
	minioBucket := os.Getenv("MINIO_BUCKET_IMAGE_LPR")
	if minioBucket != "" {
		if mc, err := minio.New(); err == nil {
			objName := fmt.Sprintf("%s-%d-%s", cameraID, time.Now().Unix(), filepath.Base(imagePath))
			if url, err := mc.UploadFile(context.Background(), minioBucket, objName, imagePath); err == nil {
				requestMeta["image_url"] = url
			} else {
				log.Printf("MinIO upload failed: %v", err)
			}
		} else {
			log.Printf("MinIO client init failed: %v", err)
		}
		log.Println("MINIO_ENDPOINT =", os.Getenv("MINIO_ENDPOINT"))
	}

	requestJSON, _ := json.Marshal(requestMeta)

	responseFinalJSON, _ := json.Marshal(finalResp)

	log := model.PlateLog{
		LocationCode:  locationCode,
		CameraID:      cameraID,
		Plate:         plate,
		Timestamp:     time.Now(),
		RequestData:   string(requestJSON),
		ResponseData:  "",
		ResponseFinal: string(responseFinalJSON),
		ImageURL:      requestMeta["image_url"],
	}

	if err := db.Create(&log).Error; err != nil {
		return nil, err
	}

	// Update request_data with potential image_url
	if reqJSON2, err := json.Marshal(requestMeta); err == nil {
		db.Model(&log).Update("request_data", string(reqJSON2))
	}

	return &finalResp, nil
}
