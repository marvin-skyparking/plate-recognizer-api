package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"plate-recognizer-api/internal/minio"
	"plate-recognizer-api/model"
	"strings"
	"time"

	"gorm.io/gorm"
)

type MemberCheckResponse struct {
	Data struct {
		Category string `json:"category"`
	} `json:"data"`
}

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

	// --- Call plate recognizer ---
	plate, score, err := Recognize(
		token,
		imagePath,
		mmc,
		cameraID,
		transactionNo,
	)
	if err != nil {
		return nil, err
	}

	plate = strings.ToUpper(plate)

	// --- Call member service ---
	resp, err := http.Get("http://backend-app-local:5000/api/members/check-plat/" + plate)
	if err != nil {
		log.Println("Error checking member status:", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Println("Member service HTTP status:", resp.StatusCode)

	// --- Check HTTP status ---
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("member service returned %d", resp.StatusCode)
	}

	// ======================
	// Decode JSON response
	// ======================
	var memberResp MemberCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&memberResp); err != nil {
		log.Println("JSON decode error:", err)
		return nil, err
	}

	if memberResp.Data.Category == "" {
		memberResp.Data.Category = "CASUAL"
	}

	finalResp := FinalResponse{
		Status:  200,
		Message: "plate recognized successfully",
		Code:    "SUCCESS",
		Data: map[string]interface{}{
			"plate":         plate,
			"score":         score,
			"status_member": memberResp.Data.Category,
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
		TransactionNo: transactionNo,
		Timestamp:     time.Now(),
		RequestData:   string(requestJSON),
		Accuracy:      fmt.Sprintf("%.2f", score),
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
