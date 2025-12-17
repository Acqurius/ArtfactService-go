package handlers

import (
	"log"
	"net/http"

	"file-upload-api/db"
	"file-upload-api/models"

	"github.com/gin-gonic/gin"
)

// ListArtifacts godoc
// @Summary      List all artifacts
// @Description  Retrieves a list of all uploaded artifacts with their metadata
// @Tags         files
// @Produce      json
// @Success      200  {array}   models.Artifact
// @Failure      500  {object}  map[string]string
// @Router       /artifact-service/v1/artifacts/ [get]
func ListArtifacts(c *gin.Context) {
	// Query all artifacts from database
	rows, err := db.DB.Query("SELECT uuid, filename, content_type, size, created_at FROM Artifacts ORDER BY created_at DESC")
	if err != nil {
		log.Println("Database query error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve artifacts"})
		return
	}
	defer rows.Close()

	// Collect all artifacts
	var artifacts []models.Artifact
	for rows.Next() {
		var artifact models.Artifact
		err := rows.Scan(&artifact.UUID, &artifact.Filename, &artifact.ContentType, &artifact.Size, &artifact.CreatedAt)
		if err != nil {
			log.Println("Row scan error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse artifacts"})
			return
		}
		artifacts = append(artifacts, artifact)
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		log.Println("Rows iteration error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve artifacts"})
		return
	}

	// Return empty array if no artifacts found (not an error)
	if artifacts == nil {
		artifacts = []models.Artifact{}
	}

	c.JSON(http.StatusOK, artifacts)
}
