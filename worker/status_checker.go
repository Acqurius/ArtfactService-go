package worker

import (
	"log"
	"time"

	"ArtifactService/db"
	"ArtifactService/storage"
)

// StartStatusChecker starts a background worker that periodically checks the status of pending artifacts
func StartStatusChecker(interval time.Duration) {
	log.Printf("Starting Status Check Worker with interval %v", interval)
	ticker := time.NewTicker(interval)
	
	// Run immediately on start
	go checkPendingArtifacts()

	go func() {
		for range ticker.C {
			checkPendingArtifacts()
		}
	}()
}

func checkPendingArtifacts() {
	// Find artifacts that are PENDING
	// We assume items created recently might not be uploaded yet, but we check anyway.
	rows, err := db.DB.Query("SELECT uuid, created_at FROM Artifacts WHERE status = 'PENDING'")
	if err != nil {
		log.Println("Worker: Failed to query pending artifacts:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var uuid string
		var createdAt time.Time
		if err := rows.Scan(&uuid, &createdAt); err != nil {
			log.Println("Worker: Failed to scan row:", err)
			continue
		}

		// Check if file exists in S3/Ceph
		exists, err := storage.CheckFileExists(uuid)
		if err != nil {
			// If generic error (network, auth), log and skip
			// We don't change status on error
			log.Printf("Worker: CheckFileExists error for %s: %v", uuid, err)
			continue
		}

		if exists {
			// File found! Update status to UPLOADED
			_, err := db.DB.Exec("UPDATE Artifacts SET status = 'UPLOADED' WHERE uuid = ?", uuid)
			if err != nil {
				log.Printf("Worker: Failed to update status for %s: %v", uuid, err)
			} else {
				log.Printf("Worker: Artifact %s status updated to UPLOADED", uuid)
			}
		} else {
			// File not found. Check if it has been pending for too long.
			// Default presigned URL expiry is usually around 15 minutes.
			// We give it a buffer, say 30 minutes.
			if time.Since(createdAt) > 30*time.Minute {
				// Mark as EXPIRED or FAILED
				_, err := db.DB.Exec("UPDATE Artifacts SET status = 'EXPIRED' WHERE uuid = ?", uuid)
				if err != nil {
					log.Printf("Worker: Failed to mark %s as EXPIRED: %v", uuid, err)
				} else {
					log.Printf("Worker: Artifact %s marked as EXPIRED (timeout)", uuid)
				}
			}
		}
	}
}
