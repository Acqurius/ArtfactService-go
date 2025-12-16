package main

import (
	"log"
	"os"

	"file-upload-api/db"
	_ "file-upload-api/docs"
	"file-upload-api/handlers"
	"file-upload-api/storage"

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

	r := gin.Default()

	// Routes
	r.POST("/artifacts/innerop/upload", handlers.UploadFile)
	r.GET("/artifacts/innerop/:id", handlers.DownloadFile)
	
	// Presigned URL routes
	r.POST("/genPresignedURL", handlers.GenPresignedURL)
	r.GET("/artifacts/:token", handlers.DownloadFileWithToken)
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
