package main

import (
	"log"
	"os"
	"time"

	"ArtifactService/db"
	_ "ArtifactService/docs"
	"ArtifactService/handlers"
	"ArtifactService/storage"
	"ArtifactService/worker"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// @title           File Upload API
	// @version         1.0
	// @description     This is a sample server for uploading and downloading files.
	// @host            localhost:8080
	// @BasePath        /

	// Initialize Database
	db.InitDB()
	
	// Initialize Storage (Ceph/S3)
	if err := storage.InitStorage(); err != nil {
		log.Fatal("Failed to initialize storage: ", err)
	}

	// Start Background Workers
	// Check for pending uploads every 60 seconds
	worker.StartStatusChecker(60 * time.Second)

	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Routes
	r.POST("/artifact-service/v1/artifacts/", handlers.UploadFile)
	r.GET("/artifact-service/v1/artifacts/", handlers.ListArtifacts)
	r.GET("/artifact-service/v1/artifacts/:uuid/action/downloadFile", handlers.DownloadFile)
	r.DELETE("/artifact-service/v1/artifacts/:uuid", handlers.DeleteArtifact)
	r.GET("/artifact-service/v1/storage/usage", handlers.GetStorageUsage)
	
	// Token generation routes
	r.POST("/genDownloadPresignedURL", handlers.GenDownloadPresignedURL)
	r.POST("/genUploadPresignedURL", handlers.GenUploadPresignedURL)
	
	// Token-based file access routes
	r.GET("/artifacts/:token", handlers.DownloadFileWithToken)
	r.POST("/artifacts/upload/:token", handlers.UploadFileWithToken)
	
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Run server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
