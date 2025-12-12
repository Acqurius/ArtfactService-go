package models

import (
	"time"
)

type Artifact struct {
	UUID        string    `json:"uuid"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
}
