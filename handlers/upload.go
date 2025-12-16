package handlers

import (
	"log"
	"net/http"

	"file-upload-api/db"
	"file-upload-api/models"
	"file-upload-api/storage"

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
	
	// Open uploaded file
	fileReader, err := file.Open()
	if err != nil {
		log.Println("Failed to open uploaded file:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to process file"})
		return
	}
	defer fileReader.Close()

	// Upload file to Ceph
	if err := storage.UploadFile(uuid, file.Filename, fileReader, file.Header.Get("Content-Type"), file.Size); err != nil {
		log.Println("Failed to upload file to Ceph:", err)
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
