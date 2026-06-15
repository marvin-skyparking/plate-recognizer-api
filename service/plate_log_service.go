package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	// "net/http"
	"os"
	"path/filepath"
	"plate-recognizer-api/internal/minio"
	"plate-recognizer-api/model"
	"strings"
	"time"

	"gorm.io/gorm"
)

// type MemberCheckResponse struct {
// 	Data struct {
// 		Category string `json:"category"`
// 	} `json:"data"`
// }

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
	transactionNo string,
	cameraID string,
	mmc string,
) (*FinalResponse, error) {

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

	// --- Call plate recognizer ---
	plate, score, err := Recognize(
		token,
		imagePath,
		mmc,
		cameraID,
		transactionNo,
	)

	requestJSON, _ := json.Marshal(requestMeta)
	var finalResp FinalResponse
	var plateLog model.PlateLog

	// Handle recognition result (success or failure)
	if err != nil {
		log.Printf("Plate recognition failed: %v", err)
		finalResp = FinalResponse{
			Status:  500,
			Message: fmt.Sprintf("plate recognition failed: %v", err),
			Code:    "RECOGNITION_FAILED",
			Data:    nil,
		}
		plateLog = model.PlateLog{
			LocationCode:  locationCode,
			CameraID:      cameraID,
			Plate:         "",
			TransactionNo: transactionNo,
			Timestamp:     time.Now(),
			RequestData:   string(requestJSON),
			Accuracy:      "0",
			ResponseData:  "",
			ResponseFinal: "",
			ImageURL:      requestMeta["image_url"],
		}
	} else {
		plate = strings.ToUpper(plate)

		finalResp = FinalResponse{
			Status:  200,
			Message: "plate recognized successfully",
			Code:    "SUCCESS",
			Data: map[string]interface{}{
				"plate": plate,
				"score": score,
			},
		}
		plateLog = model.PlateLog{
			LocationCode:  locationCode,
			CameraID:      cameraID,
			Plate:         plate,
			TransactionNo: transactionNo,
			Timestamp:     time.Now(),
			RequestData:   string(requestJSON),
			Accuracy:      fmt.Sprintf("%.2f", score),
			ResponseData:  "",
			ResponseFinal: "",
			ImageURL:      requestMeta["image_url"],
		}
	}

	// Set final response for both success and failure cases
	responseFinalJSON, _ := json.Marshal(finalResp)
	plateLog.ResponseFinal = string(responseFinalJSON)

	// Save to database regardless of success or failure
	if err := db.Create(&plateLog).Error; err != nil {
		log.Printf("Database save failed: %v", err)
		// Even if save fails, return the response status
		return &finalResp, err
	}

	// Update request_data (with image_url if exists)
	if reqJSON2, err := json.Marshal(requestMeta); err == nil {
		db.Model(&plateLog).Update("request_data", string(reqJSON2))
	}

	return &finalResp, nil
}
