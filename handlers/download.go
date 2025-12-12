package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"path/filepath"

	"file-upload-api/db"
	"file-upload-api/models"

	"github.com/gin-gonic/gin"
)

// DownloadFile godoc
// @Summary      Download a file
// @Description  Downloads a file by its ID
// @Tags         files
// @Produce      octet-stream
// @Param        id   path      string  true  "File ID"
// @Success      200  {file}    file
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /artifacts/innerop/{id} [get]
func DownloadFile(c *gin.Context) {
	id := c.Param("id")

	var metadata models.Artifact
	row := db.DB.QueryRow("SELECT uuid, filename, content_type, size FROM Artifacts WHERE uuid = ?", id)
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

	// Construct file path on disk
	// We stored it as UUID + original extension in upload handler
	// But we didn't store the extension in DB?
	// Ah, in upload.go we did: `newFilename := id + ext`
	// But we only stored `filename` (original name) in DB.
	// We need to know the extension to find the file on disk or we should store the disk path/filename in DB.
	// Let's check upload.go again.
	// `metadata.Filename` comes from `file.Filename` (original).
	// We should probably check the disk for files starting with ID or store the storage filename in DB.
	
	// FIX: Let's assume we can glob for the file or we change the DB schema to store `storage_name`.
	// For simplicity, let's look for the file in the uploads directory matching the ID.
	
	matches, err := filepath.Glob(filepath.Join("uploads", id+"*"))
	if err != nil || len(matches) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "File content not found"})
		return
	}
	
	filePath := matches[0]

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+metadata.Filename)
	c.Header("Content-Type", metadata.ContentType)
	c.File(filePath)
}
