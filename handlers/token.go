package handlers

import (
	"database/sql"
	"log"
	"net"
	"net/http"
	"time"

	"file-upload-api/db"
	"file-upload-api/models"
	"file-upload-api/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GenPresignedURL godoc
// @Summary      Generate a Presigned URL
// @Description  Generates a token for temporary file access with constraints
// @Tags         tokens
// @Accept       json
// @Produce      json
// @Param        request body models.GenTokenRequest true "Token constraints"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /genPresignedURL [post]
func GenPresignedURL(c *gin.Context) {
	var req models.GenTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify artifact exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM Artifacts WHERE uuid = ?)", req.ArtifactUUID).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Artifact not found"})
		return
	}

	// Generate Token
	token := uuid.New().String()

	// Insert into DB
	_, err = db.DB.Exec(`
		INSERT INTO tokens (token, artifact_uuid, valid_from, valid_to, max_downloads, allowed_cidr)
		VALUES (?, ?, ?, ?, ?, ?)`,
		token, req.ArtifactUUID, req.ValidFrom, req.ValidTo, req.MaxDownloads, req.AllowedCIDR)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Generate URL
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	presignedURL := scheme + "://" + c.Request.Host + "/artifacts/" + token

	c.JSON(http.StatusOK, gin.H{
		"token":         token,
		"presigned_url": presignedURL,
	})
}

// DownloadFileWithToken godoc
// @Summary      Download file with Presigned URL
// @Description  Download a file using a token, enforcing constraints. Returns a 302 redirect to S3 presigned URL for direct download.
// @Tags         tokens
// @Produce      octet-stream
// @Param        token path string true "Access Token"
// @Success      302  {string}  string  "Redirect to S3 presigned URL"
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /artifacts/{token} [get]
func DownloadFileWithToken(c *gin.Context) {
	token := c.Param("token")

	var t models.Token
	var filename string
	var contentType string

	// Query token and artifact details
	// We need to fetch basic info + current state
	row := db.DB.QueryRow(`
		SELECT t.token, t.artifact_uuid, t.valid_from, t.valid_to, t.max_downloads, t.current_downloads, t.allowed_cidr,
		       a.filename, a.content_type
		FROM tokens t
		JOIN Artifacts a ON t.artifact_uuid = a.uuid
		WHERE t.token = ?`, token)
	
	err := row.Scan(&t.Token, &t.ArtifactUUID, &t.ValidFrom, &t.ValidTo, &t.MaxDownloads, &t.CurrentDownloads, &t.AllowedCIDR, &filename, &contentType)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid or expired token"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	now := time.Now()

	// 1. Time Validation
	if t.ValidFrom != nil && now.Before(*t.ValidFrom) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Token not yet valid"})
		return
	}
	if t.ValidTo != nil && now.After(*t.ValidTo) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Token expired"})
		return
	}

	// 2. Count Validation
	if t.MaxDownloads != nil && t.CurrentDownloads >= *t.MaxDownloads {
		c.JSON(http.StatusForbidden, gin.H{"error": "Download limit reached"})
		return
	}

	// 3. IP Validation
	if t.AllowedCIDR != "" {
		clientIP := c.ClientIP()
		_, ipNet, err := net.ParseCIDR(t.AllowedCIDR)
		if err != nil {
			// If CIDR in DB is invalid, we probably should block or log error. Blocking for safety.
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid CIDR configuration"})
			return
		}
		ip := net.ParseIP(clientIP)
		if ip == nil || !ipNet.Contains(ip) {
			c.JSON(http.StatusForbidden, gin.H{"error": "IP not allowed"})
			return
		}
	}

	// Increment download count
	_, err = db.DB.Exec("UPDATE tokens SET current_downloads = current_downloads + 1 WHERE token = ?", token)
	if err != nil {
		// Just log error, don't fail download? Or fail? Better fail to enforce limits strictly.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update download stats"})
		return
	}

	// Generate presigned URL for direct S3 download (expires in 15 minutes)
	presignedURL, err := storage.GeneratePresignedURL(t.ArtifactUUID, 15)
	if err != nil {
		log.Println("Failed to generate presigned URL:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate download URL"})
		return
	}

	// Return 302 redirect to the presigned URL
	c.Redirect(http.StatusFound, presignedURL)
}
