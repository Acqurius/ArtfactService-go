# ArtfactService-go

A file upload and download API service with presigned URL support, built with Go and SQLite.

## Features

- 📤 **File Upload**: Upload files with automatic UUID generation
- 📥 **File Download**: Download files by UUID
- 🔐 **Presigned URLs**: Generate time-limited, access-controlled download tokens
- 📊 **Swagger Documentation**: Interactive API documentation at `/swagger/index.html`
- 💾 **SQLite Database**: Lightweight, file-based database for metadata storage

## Prerequisites

- Go 1.24.0 or higher
- SQLite (included via modernc.org/sqlite)

## Quick Start

### 1. Clone the Repository

```bash
git clone <your-repo-url>
cd ArtfactService-go
```

### 2. Initialize the Database

**Option A: Using the Go script (Recommended)**

```bash
go run scripts/init_db.go
```

**Option B: Using SQLite CLI**

```bash
sqlite3 files.db < schema.sql
```

### 3. Install Dependencies

```bash
go mod download
```

### 4. Run the Server

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
POST /artifacts/innerop/upload
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

### Download File
```http
GET /artifacts/innerop/:id
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

## Notes

- The `files.db` SQLite database file is excluded from version control (see `.gitignore`)
- Uploaded files are stored in the `uploads/` directory (also excluded from git)
- Always run the database initialization script before first use

## License

[Your License Here]
