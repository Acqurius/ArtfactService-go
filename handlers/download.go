package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"file-upload-api/db"
	"file-upload-api/models"
	"file-upload-api/storage"

	"github.com/gin-gonic/gin"
)

// DownloadFile godoc
// @Summary      Download a file
// @Description  Downloads a file by its UUID
// @Tags         files
// @Produce      octet-stream
// @Param        uuid   path      string  true  "File UUID"
// @Success      200  {file}    file
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /artifact-service/v1/artifacts/{uuid}/action/downloadFile [get]
func DownloadFile(c *gin.Context) {
	uuid := c.Param("uuid")

	var metadata models.Artifact
	row := db.DB.QueryRow("SELECT uuid, filename, content_type, size FROM Artifacts WHERE uuid = ?", uuid)
	err := row.Scan(&metadata.UUID, &metadata.Filename, &metadata.ContentType, &metadata.Size)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		} else {
			log.Println("Database error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}


	// Download file from Ceph
	fileReader, err := storage.DownloadFile(uuid)
	if err != nil {
		log.Println("Failed to download file from Ceph:", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "File content not found"})
		return
	}
	defer fileReader.Close()

	// Stream file to response
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+metadata.Filename)
	c.Header("Content-Type", metadata.ContentType)
	c.DataFromReader(http.StatusOK, metadata.Size, metadata.ContentType, fileReader, nil)
}
