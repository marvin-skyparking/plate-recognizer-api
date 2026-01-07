package service

import (
	"errors"
	"plate-recognizer-api/model"

	"gorm.io/gorm"
)

// CreateUser creates a new user with hashed password
func CreateUser(db *gorm.DB, username, password string) (*model.User, error) {
	if username == "" || password == "" {
		return nil, errors.New("username and password are required")
	}

	// Check for existing username
	var existing model.User
	if err := db.Where("username = ?", username).First(&existing).Error; err == nil {
		return nil, errors.New("username already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	user := &model.User{
		Username: username,
		IsActive: true,
	}

	if err := user.SetPassword(password); err != nil {
		return nil, err
	}

	if err := db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}
