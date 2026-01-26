package model

import "time"

type PlateLog struct {
	ID            uint   `gorm:"primaryKey"`
	LocationCode  string `gorm:"type:varchar(50);index"`
	CameraID      string `gorm:"type:varchar(50);index"`
	TransactionNo string `gorm:"type:varchar(100);index"`
	Plate         string `gorm:"type:varchar(20);index"`
	Timestamp     time.Time
	RequestData   string `gorm:"type:text"`
	ResponseData  string `gorm:"type:text"`
	ResponseFinal string `gorm:"type:text"`
	ImageURL      string `gorm:"type:text" json:"image_url"`
	CreatedAt     time.Time
}
