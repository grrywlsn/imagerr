package db

import (
	"time"
)

type Image struct {
    ID               int64     `json:"id"`
    OriginalFilename string    `json:"original_filename"`
    UUIDFilename     string    `json:"uuid_filename"`
    Description      string    `json:"description"`
    URL              string    `json:"url"`
    Tags             []string  `json:"tags" gorm:"type:text[]"`
    StoragePath      string    `json:"storage_path"`
    CreatedAt        time.Time `json:"created_at"`
    ViewCount        int       `json:"view_count" gorm:"default:0"`
}