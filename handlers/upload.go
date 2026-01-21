package handlers

import (
	"log"
	"net/http"

	"ArtifactService/db"
	"ArtifactService/logger"
	"ArtifactService/models"
	"ArtifactService/storage"

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
// @Router       /artifact-service/v1/artifacts/ [post]
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

	logger.Record(logger.ActionUpload, uuid, c.ClientIP(), "", "SUCCESS", "Standard upload")
}

// CompleteUpload godoc
// @Summary      Mark upload as complete
// @Description  Allows client to notify server that upload to S3 is complete. Server verifies file existence and updates status.
// @Tags         files
// @Accept       json
// @Produce      json
// @Param        uuid path string true "Artifact UUID"
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /artifact-service/v1/artifacts/{uuid}/complete [post]
func CompleteUpload(c *gin.Context) {
	uuid := c.Param("uuid")

	// Check current status
	var status string
	err := db.DB.QueryRow("SELECT status FROM Artifacts WHERE uuid = ?", uuid).Scan(&status)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Artifact not found"})
		return
	}

	// Idempotency: If already uploaded, return success immediately
	if status == "UPLOADED" {
		c.JSON(http.StatusOK, gin.H{
			"message": "Upload already completed",
			"status":  "UPLOADED",
		})
		return
	}

	// Verify file existence in S3/Ceph
	exists, err := storage.CheckFileExists(uuid)
	if err != nil {
		log.Printf("Failed to check file existence for %s: %v", uuid, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify storage"})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error":  "File verification failed: file not found in storage",
			"status": status,
		})
		return
	}

	// Update status to UPLOADED
	_, err = db.DB.Exec("UPDATE Artifacts SET status = 'UPLOADED' WHERE uuid = ?", uuid)
	if err != nil {
		log.Printf("Failed to update status for %s: %v", uuid, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Upload verification successful",
		"status":  "UPLOADED",
	})
	
	logger.Record(logger.ActionUpload, uuid, c.ClientIP(), "", "SUCCESS", "Presigned upload completed")
}
