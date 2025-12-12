package handlers

import (
	"log"
	"net/http"
	"path/filepath"

	"file-upload-api/db"
	"file-upload-api/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadFile godoc
// @Summary      Upload a file
// @Description  Uploads a file and saves metadata to the database
// @Tags         files
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "File to upload"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /artifacts/innerop/upload [post]
func UploadFile(c *gin.Context) {
	// Single file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file is received"})
		return
	}

	// Generate UUID
	uuid := uuid.New().String()
	ext := filepath.Ext(file.Filename)
	newFilename := uuid + ext
	uploadPath := filepath.Join("uploads", newFilename)

	// Save file to disk
	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		log.Println("Failed to save file:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to save file"})
		return
	}

	// Save metadata to DB
	metadata := models.Artifact{
		UUID:        uuid,
		Filename:    file.Filename,
		ContentType: file.Header.Get("Content-Type"),
		Size:        file.Size,
	}

	_, err = db.DB.Exec("INSERT INTO Artifacts (uuid, filename, content_type, size) VALUES (?, ?, ?, ?)",
		metadata.UUID, metadata.Filename, metadata.ContentType, metadata.Size)
	if err != nil {
		log.Println("Failed to insert metadata:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Create download link
	// Assuming the server is running on the Host header address
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	downloadURL := scheme + "://" + c.Request.Host + "/artifacts/innerop/" + uuid

	c.JSON(http.StatusOK, gin.H{
		"message":      "File uploaded successfully",
		"uuid":         uuid,
		"download_url": downloadURL,
	})
}
