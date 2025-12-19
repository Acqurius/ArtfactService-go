# ArtfactService-go

A file upload and download API service with presigned URL support, built with Go and SQLite.

## Features

- 📤 **File Upload**: Upload files with automatic UUID generation
- 📥 **File Download**: Download files by UUID
- 🔐 **Presigned URLs**: Generate time-limited, access-controlled download tokens
- 📊 **Swagger Documentation**: Interactive API documentation at `/swagger/index.html`
- 💾 **SQLite Database**: Lightweight, file-based database for metadata storage
- ☁️ **Ceph Storage**: S3-compatible object storage for scalable file management

## Prerequisites

- Go 1.24.0 or higher
- SQLite (included via modernc.org/sqlite)
- Ceph S3-compatible storage endpoint with access credentials

## Quick Start

### 1. Clone the Repository

```bash
git clone <your-repo-url>
cd ArtfactService-go
```

### 2. Configure Ceph Storage

Set the following environment variables with your Ceph credentials:

```bash
export CEPH_ACCESS_KEY="your-access-key"
export CEPH_SECRET_KEY="your-secret-key"
export CEPH_ENDPOINT="http://your-ceph-endpoint:port"
export CEPH_BUCKET="artifacts"  # Optional, defaults to "artifacts"
```

**Example:**
```bash
export CEPH_ACCESS_KEY="C1L81CIJEQ538MNODAOL"
export CEPH_SECRET_KEY="69Ia5JnXEp4ACyxANIefaZpCBxD3zCZPEWJQMv3l"
export CEPH_ENDPOINT="http://10.188.157.5:80"
export CEPH_BUCKET="artifacts"
```

> [!IMPORTANT]
> Ensure the specified bucket exists in your Ceph storage before running the server.

### 3. Initialize the Database

**Option A: Using the Go script (Recommended)**

```bash
go run scripts/init_db.go
```

**Option B: Using SQLite CLI**

```bash
sqlite3 files.db < schema.sql
```

### 4. Install Dependencies

```bash
go mod download
```

### 5. Run the Server

```bash
go run main.go
```

Or build and run:

```bash
go build -o server .
./server
```

The server will start on port `8080` by default (configurable via `PORT` environment variable).

## Database Schema

The application uses SQLite with two main tables:

### Artifacts Table
Stores file metadata for uploaded artifacts.

| Column | Type | Description |
|--------|------|-------------|
| uuid | TEXT | Primary key, unique identifier |
| filename | TEXT | Original filename |
| content_type | TEXT | MIME type |
| size | BIGINT | File size in bytes |
| created_at | TIMESTAMP | Upload timestamp |

### Tokens Table
Stores presigned URL tokens with access control.

| Column | Type | Description |
|--------|------|-------------|
| token | TEXT | Primary key, unique token |
| artifact_uuid | TEXT | Foreign key to Artifacts |
| valid_from | TIMESTAMP | Token validity start time (optional) |
| valid_to | TIMESTAMP | Token expiration time (optional) |
| max_downloads | BIGINT | Maximum download count (optional) |
| current_downloads | BIGINT | Current download count |
| allowed_cidr | TEXT | IP CIDR restriction (optional) |
| created_at | TIMESTAMP | Token creation time |

## API Endpoints

### Upload File
```http
POST /artifact-service/v1/artifacts/
Content-Type: multipart/form-data

file: <binary>
```

**Response:**
```json
{
  "uuid": "550e8400-e29b-41d4-a716-446655440000",
  "filename": "example.pdf",
  "content_type": "application/pdf",
  "size": 1024,
  "created_at": "2024-01-01T00:00:00Z"
}
```

### List All Artifacts
```http
GET /artifact-service/v1/artifacts/
```

**Response:**
```json
[
  {
    "uuid": "550e8400-e29b-41d4-a716-446655440000",
    "filename": "example.pdf",
    "content_type": "application/pdf",
    "size": 1024,
    "created_at": "2024-01-01T00:00:00Z"
  },
  {
    "uuid": "660e8400-e29b-41d4-a716-446655440001",
    "filename": "document.docx",
    "content_type": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
    "size": 2048,
    "created_at": "2024-01-02T00:00:00Z"
  }
]
```

**Note:** Returns an empty array `[]` if no artifacts exist.

### Download File
```http
GET /artifact-service/v1/artifacts/{uuid}/action/downloadFile
```

**Parameters:**
- `uuid` (path) - The UUID of the artifact to download

**Response:** Binary file content with appropriate headers

### Storage Usage (New)
```http
GET /artifact-service/v1/storage/usage
```

**Response:**
```json
{
  "total_space": 10737418240,
  "used_space": 35930,
  "remaining_space": 10737382310,
  "usage_percent": 0.00033,
  "file_count": 3
}
```

### Generate Presigned URL
```http
POST /genPresignedURL
Content-Type: application/json

{
  "artifact_uuid": "550e8400-e29b-41d4-a716-446655440000",
  "valid_from": "2024-01-01T00:00:00Z",
  "valid_to": "2024-01-02T00:00:00Z",
  "max_downloads": 5,
  "allowed_cidr": "192.168.1.0/24"
}
```

**Response:**
```json
{
  "token": "abc123def456...",
  "download_url": "http://localhost:8080/artifacts/abc123def456..."
}
```

### Download with Token
```http
GET /artifacts/:token
```

### API Documentation
```http
GET /swagger/index.html
```

## Project Structure

```
ArtfactService-go/
├── db/                 # Database initialization and connection
│   └── db.go
├── docs/               # Swagger documentation (auto-generated)
├── handlers/           # HTTP request handlers
│   ├── upload.go
│   ├── download.go
│   └── token.go
├── models/             # Data models
│   ├── file.go
│   └── token.go
├── storage/            # Ceph/S3 storage integration
│   └── storage.go
├── scripts/            # Utility scripts
│   └── init_db.go     # Database initialization script
├── schema.sql          # SQL schema definition
├── main.go             # Application entry point
├── go.mod              # Go module dependencies
└── .gitignore
```

## Development

### Regenerate Swagger Documentation

```bash
swag init
```

### Run Tests

```bash
go test ./...
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | Server port |
| CEPH_ACCESS_KEY | (required) | Ceph S3 access key |
| CEPH_SECRET_KEY | (required) | Ceph S3 secret key |
| CEPH_ENDPOINT | (required) | Ceph S3 endpoint URL |
| CEPH_BUCKET | artifacts | Ceph S3 bucket name |

## Database Management

### Reset Database

To reset the database, simply delete `files.db` and run the initialization script again:

```bash
rm files.db
go run scripts/init_db.go
```

### Backup Database

```bash
cp files.db files.db.backup
```

## Storage Architecture

- **Metadata**: Stored in SQLite database (`files.db`)
- **File Content**: Stored in Ceph S3-compatible object storage
- Files are stored with UUID as the object key in Ceph
- Original filename and metadata are preserved in the database

## Notes

- The `files.db` SQLite database file is excluded from version control (see `.gitignore`)
- Uploaded files are stored in Ceph object storage (not local filesystem)
- Always run the database initialization script before first use
- Ensure Ceph credentials are properly configured before starting the server

## License

[Your License Here]
