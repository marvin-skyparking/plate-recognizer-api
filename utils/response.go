package utils

import "github.com/gofiber/fiber/v2"

type APIResponse struct {
	Status  int         `json:"status"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(c *fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(APIResponse{
		Status:  status,
		Code:    "SUCCESS",
		Message: message,
		Data:    data,
	})
}

func Error(c *fiber.Ctx, status int, code string, message string) error {
	return c.Status(status).JSON(APIResponse{
		Status:  status,
		Code:    code,
		Message: message,
	})
}
