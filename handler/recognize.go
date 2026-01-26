package handler

import (
	"fmt"
	"os"

	"plate-recognizer-api/service"
	"plate-recognizer-api/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type RecognizeHandler struct {
	Token string
	DB    *gorm.DB
}

func NewRecognizeHandler(token string, db *gorm.DB) *RecognizeHandler {
	return &RecognizeHandler{
		Token: token,
		DB:    db,
	}
}

func (h *RecognizeHandler) Recognize(c *fiber.Ctx) error {
	// ==========================
	// Validate form-data
	// ==========================
	file, err := c.FormFile("image")
	if err != nil {
		return utils.Error(
			c,
			fiber.StatusBadRequest,
			"BAD_REQUEST",
			"image is required",
		)
	}

	locationCode := c.FormValue("location_code")
	cameraID := c.FormValue("camera_id")
	transactionNo := c.FormValue("transaction_no")
	mmc := c.FormValue("mmc")

	if locationCode == "" || cameraID == "" {
		return utils.Error(
			c,
			fiber.StatusBadRequest,
			"BAD_REQUEST",
			"location_code and camera_id are required",
		)
	}

	// ==========================
	// Save temp image
	// ==========================
	tmp, err := os.CreateTemp("", "plate-*.jpg")
	if err != nil {
		return utils.Error(
			c,
			fiber.StatusInternalServerError,
			"INTERNAL_ERROR",
			"failed to create temp file",
		)
	}
	defer os.Remove(tmp.Name())

	if err := c.SaveFile(file, tmp.Name()); err != nil {
		return utils.Error(
			c,
			fiber.StatusInternalServerError,
			"INTERNAL_ERROR",
			"failed to save image",
		)
	}

	// ==========================
	// CALL SERVICE (SAVE TO DB)
	// ==========================
	fmt.Print(transactionNo)
	resp, err := service.RecognizeAndSavePlateLog(
		h.DB,
		h.Token,
		tmp.Name(),
		locationCode,
		transactionNo,
		cameraID,
		mmc,
	)
	if err != nil {
		return utils.Error(
			c,
			fiber.StatusBadRequest,
			"PROCESS_FAILED",
			err.Error(),
		)
	}

	// ==========================
	// SUCCESS RESPONSE
	// ==========================
	return utils.Success(
		c,
		fiber.StatusOK,
		resp.Message,
		resp.Data,
	)
}
