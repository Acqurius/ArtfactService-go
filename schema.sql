-- ArtfactService Database Schema
-- SQLite Database Initialization Script

-- Create Artifacts table
CREATE TABLE IF NOT EXISTS Artifacts (
    uuid TEXT PRIMARY KEY,
    filename TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size BIGINT NOT NULL,
    status TEXT DEFAULT 'UPLOADED',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create tokens table
CREATE TABLE IF NOT EXISTS tokens (
    token TEXT PRIMARY KEY,
    artifact_uuid TEXT NOT NULL,
    valid_from TIMESTAMP,
    valid_to TIMESTAMP,
    max_downloads BIGINT,
    current_downloads BIGINT DEFAULT 0,
    allowed_cidr TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(artifact_uuid) REFERENCES Artifacts(uuid)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_artifacts_created_at ON Artifacts(created_at);
CREATE INDEX IF NOT EXISTS idx_tokens_artifact_uuid ON tokens(artifact_uuid);
CREATE INDEX IF NOT EXISTS idx_tokens_valid_to ON tokens(valid_to);
