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

	// --- Call plate recognizer ---
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

	// --- Request metadata ---
	requestMeta := map[string]string{
		"location_code": locationCode,
		"camera_id":     cameraID,
		"mmc":           mmc,
	}

	// ======================================================
	// ================= MINIO UPLOAD =======================
	// ======================================================

	log.Println("MINIO_ENDPOINT =", os.Getenv("MINIO_ENDPOINT"))
	log.Println("MINIO_BUCKET_IMAGE_LPR =", os.Getenv("MINIO_BUCKET_IMAGE_LPR"))
	log.Println("MINIO_USE_SSL =", os.Getenv("MINIO_USE_SSL"))

	minioBucket := os.Getenv("MINIO_BUCKET_IMAGE_LPR")

	if minioBucket != "" {
		mc, err := minio.New()
		if err != nil {
			log.Printf("MinIO init failed: %v", err)
		} else {
			objName := fmt.Sprintf(
				"%s-%d-%s",
				cameraID,
				time.Now().Unix(),
				filepath.Base(imagePath),
			)

			url, err := mc.UploadFile(
				context.Background(),
				minioBucket,
				objName,
				imagePath,
			)
			if err != nil {
				log.Printf("MinIO upload failed: %v", err)
			} else {
				requestMeta["image_url"] = url
			}
		}
	}

	// ======================================================

	requestJSON, _ := json.Marshal(requestMeta)
	responseFinalJSON, _ := json.Marshal(finalResp)

	plateLog := model.PlateLog{
		LocationCode:  locationCode,
		CameraID:      cameraID,
		Plate:         plate,
		Timestamp:     time.Now(),
		RequestData:   string(requestJSON),
		ResponseData:  "",
		ResponseFinal: string(responseFinalJSON),
		ImageURL:      requestMeta["image_url"],
	}

	if err := db.Create(&plateLog).Error; err != nil {
		return nil, err
	}

	// Update request_data (with image_url if exists)
	if reqJSON2, err := json.Marshal(requestMeta); err == nil {
		db.Model(&plateLog).Update("request_data", string(reqJSON2))
	}

	return &finalResp, nil
}
