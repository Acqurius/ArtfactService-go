package handlers

import (
	"log"
	"net/http"

	"ArtifactService/db"
	"ArtifactService/storage"

	"github.com/gin-gonic/gin"
)

// DeleteArtifact godoc
// @Summary      Delete an artifact
// @Description  Deletes an artifact by its UUID from both database and storage
// @Tags         files
// @Produce      json
// @Param        uuid   path      string  true  "File UUID"
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /artifact-service/v1/artifacts/{uuid} [delete]
func DeleteArtifact(c *gin.Context) {
	uuid := c.Param("uuid")

	// Check if artifact exists in DB
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM Artifacts WHERE uuid = ?)", uuid).Scan(&exists)
	if err != nil {
		log.Println("Database error checking existence:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Artifact not found"})
		return
	}

	// Delete from S3/Ceph
	err = storage.DeleteFile(uuid)
	if err != nil {
		log.Println("Failed to delete file from storage:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file content"})
		return
	}

	// Delete from DB
	_, err = db.DB.Exec("DELETE FROM Artifacts WHERE uuid = ?", uuid)
	if err != nil {
		log.Println("Failed to delete file record from database:", err)
		// Note: The file is already deleted from storage at this point.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Artifact deleted successfully", "uuid": uuid})
}
