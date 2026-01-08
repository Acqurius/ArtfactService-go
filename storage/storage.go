package storage

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var (
	s3Client   *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	bucketName string
)

// InitStorage initializes the S3/Ceph client
func InitStorage() error {
	// Get configuration from environment variables
	accessKey := os.Getenv("CEPH_ACCESS_KEY")
	secretKey := os.Getenv("CEPH_SECRET_KEY")
	endpoint := os.Getenv("CEPH_ENDPOINT")
	bucketName = os.Getenv("CEPH_BUCKET")

	// Set default bucket name if not provided
	if bucketName == "" {
		bucketName = "artifacts"
	}

	// Validate required configuration
	if accessKey == "" || secretKey == "" || endpoint == "" {
		return fmt.Errorf("missing required Ceph configuration: CEPH_ACCESS_KEY, CEPH_SECRET_KEY, and CEPH_ENDPOINT must be set")
	}

	// Create AWS session with Ceph endpoint
	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String("us-east-1"), // Ceph doesn't use regions, but SDK requires it
		S3ForcePathStyle: aws.Bool(true),          // Required for Ceph compatibility
	})
	if err != nil {
		return fmt.Errorf("failed to create S3 session: %w", err)
	}

	// Create S3 client
	s3Client = s3.New(sess)
	uploader = s3manager.NewUploader(sess)
	downloader = s3manager.NewDownloader(sess)

	log.Printf("Storage initialized: endpoint=%s, bucket=%s", endpoint, bucketName)
	return nil
}

// UploadFile uploads a file to Ceph storage
func UploadFile(uuid, filename string, file io.Reader, contentType string, size int64) error {
	// Use UUID as the object key in Ceph
	key := uuid

	// Upload to S3/Ceph
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
		Metadata: map[string]*string{
			"original-filename": aws.String(filename),
			"file-size":         aws.String(fmt.Sprintf("%d", size)),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to Ceph: %w", err)
	}

	log.Printf("File uploaded successfully: uuid=%s, filename=%s, size=%d", uuid, filename, size)
	return nil
}

// DownloadFile downloads a file from Ceph storage
func DownloadFile(uuid string) (io.ReadCloser, error) {
	// Get object from S3/Ceph
	result, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(uuid),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file from Ceph: %w", err)
	}

	return result.Body, nil
}

// DeleteFile deletes a file from Ceph storage
func DeleteFile(uuid string) error {
	_, err := s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(uuid),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from Ceph: %w", err)
	}

	log.Printf("File deleted successfully: uuid=%s", uuid)
	return nil
}

// GetBucketName returns the configured bucket name
func GetBucketName() string {
	return bucketName
}

// GeneratePresignedURL generates a presigned URL for downloading a file from Ceph/S3
func GeneratePresignedURL(uuid string, expirationMinutes int) (string, error) {
	if expirationMinutes <= 0 {
		expirationMinutes = 15 // Default to 15 minutes
	}

	req, _ := s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(uuid),
	})

	urlStr, err := req.Presign(time.Duration(expirationMinutes) * time.Minute)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	log.Printf("Generated presigned URL for uuid=%s, expires in %d minutes", uuid, expirationMinutes)
	return urlStr, nil
}
