package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"

	"file-upload-api/db"

	"github.com/gin-gonic/gin"
)

type StorageUsage struct {
	TotalSpace     int64   `json:"total_space"`
	UsedSpace      int64   `json:"used_space"`
	RemainingSpace int64   `json:"remaining_space"`
	UsagePercent   float64 `json:"usage_percent"`
	FileCount      int64   `json:"file_count"`
}

// GetStorageUsage godoc
// @Summary      Get storage usage statistics
// @Description  Retrieves current storage usage including total space, used space, remaining space, and file count.
// @Tags         storage
// @Produce      json
// @Success      200  {object}  StorageUsage
// @Failure      500  {object}  map[string]string
// @Router       /artifact-service/v1/storage/usage [get]
func GetStorageUsage(c *gin.Context) {
	// 1. Get Quota (Total Space) from Env
	quotaStr := os.Getenv("STORAGE_QUOTA")
	var totalSpace int64
	if quotaStr == "" {
		// Default to 10GB if not set, for demonstration purposes. 
		// Or strictly 0 if we want to show "unlimited". 
		// Let's default to 10GB (10 * 1024 * 1024 * 1024) to make it look realistic for the API request.
		totalSpace = 10 * 1024 * 1024 * 1024 
	} else {
		var err error
		totalSpace, err = strconv.ParseInt(quotaStr, 10, 64)
		if err != nil {
			log.Printf("Invalid STORAGE_QUOTA value: %s, defaulting to 10GB", quotaStr)
			totalSpace = 10 * 1024 * 1024 * 1024
		}
	}

	// 2. Query Database for Used Space and File Count
	var usedSpace int64
	var fileCount int64

	// calculating used space and file count
	// COALESCE(SUM(size), 0) handles the case where the table is empty (returns 0 instead of NULL)
	row := db.DB.QueryRow("SELECT COUNT(*), COALESCE(SUM(size), 0) FROM Artifacts")
	err := row.Scan(&fileCount, &usedSpace)
	if err != nil {
		if err == sql.ErrNoRows {
			// Should be handled by COALESCE/COUNT logic, but just in case
			fileCount = 0
			usedSpace = 0
		} else {
			log.Println("Database query error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve storage usage"})
			return
		}
	}

	// 3. Calculate Derived Metrics
	remainingSpace := totalSpace - usedSpace
	if remainingSpace < 0 {
		remainingSpace = 0
	}

	var usagePercent float64
	if totalSpace > 0 {
		usagePercent = (float64(usedSpace) / float64(totalSpace)) * 100
	} else {
		usagePercent = 0 // Avoid division by zero if totalSpace is 0
	}

	// 4. Return Response
	response := StorageUsage{
		TotalSpace:     totalSpace,
		UsedSpace:      usedSpace,
		RemainingSpace: remainingSpace,
		UsagePercent:   usagePercent,
		FileCount:      fileCount,
	}

	c.JSON(http.StatusOK, response)
}
