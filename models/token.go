package models

import (
	"time"
)

type Token struct {
	Token            string    `json:"token"`
	ArtifactUUID     string    `json:"artifact_uuid"`
	ValidFrom        *time.Time `json:"valid_from"` // Optional
	ValidTo          *time.Time `json:"valid_to"`   // Optional
	MaxDownloads     *int64    `json:"max_downloads"` // Optional
	CurrentDownloads int64     `json:"current_downloads"`
	AllowedCIDR      string    `json:"allowed_cidr"` // Optional
	CreatedAt        time.Time `json:"created_at"`
}

type GenTokenRequest struct {
	ArtifactUUID string     `json:"artifact_uuid" binding:"required"`
	ValidFrom    *time.Time `json:"valid_from"`
	ValidTo      *time.Time `json:"valid_to"`
	MaxDownloads *int64     `json:"max_downloads"`
	AllowedCIDR  string     `json:"allowed_cidr"`
}

type GenUploadTokenRequest struct {
	ValidFrom    *time.Time `json:"valid_from"`
	ValidTo      *time.Time `json:"valid_to"`
	MaxUploads   *int       `json:"max_uploads"`
	AllowedCIDR  string     `json:"allowed_cidr"`
}
