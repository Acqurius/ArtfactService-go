/**
 * Artifact Service Client
 * 
 * A simple JavaScript client for uploading and downloading files using token-based presigned URLs.
 * This client handles all the complexity of token generation and presigned URL management.
 * 
 * @example
 * const client = new ArtifactClient('http://localhost:8080');
 * 
 * // Upload a file
 * const file = document.getElementById('fileInput').files[0];
 * const result = await client.uploadFile(file, { maxUploads: 5 });
 * console.log('Uploaded:', result.uuid);
 * 
 * // Download a file
 * await client.downloadFile(artifactUuid, 'downloaded.txt', { maxDownloads: 3 });
 */
class ArtifactClient {
    /**
     * Create an ArtifactClient instance
     * @param {string} baseUrl - Base URL of the artifact service (e.g., 'http://localhost:8080')
     */
    constructor(baseUrl) {
        this.baseUrl = baseUrl.replace(/\/$/, ''); // Remove trailing slash
    }

    /**
     * Upload a file using token-based presigned URL
     * 
     * @param {File} file - The file to upload (from input[type="file"])
     * @param {Object} options - Upload options
     * @param {number} [options.maxUploads=1] - Maximum number of uploads allowed with this token
     * @param {string} [options.validFrom] - ISO 8601 timestamp when token becomes valid
     * @param {string} [options.validTo] - ISO 8601 timestamp when token expires
     * @param {string} [options.allowedCIDR] - CIDR notation for IP restriction (e.g., '192.168.1.0/24')
     * @param {Function} [options.onProgress] - Progress callback (percent) => void
     * 
     * @returns {Promise<Object>} Upload result with uuid, filename, size, contentType
     * 
     * @example
     * const file = document.getElementById('fileInput').files[0];
     * const result = await client.uploadFile(file, {
     *   maxUploads: 5,
     *   onProgress: (percent) => console.log(`Upload: ${percent}%`)
     * });
     * console.log('File UUID:', result.uuid);
     */
    async uploadFile(file, options = {}) {
        const {
            maxUploads = 1,
            validFrom = null,
            validTo = null,
            allowedCIDR = null,
            onProgress = null
        } = options;

        try {
            // Step 1: Generate upload token
            const tokenResponse = await fetch(`${this.baseUrl}/genUploadPresignedURL`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    max_uploads: maxUploads,
                    valid_from: validFrom,
                    valid_to: validTo,
                    allowed_cidr: allowedCIDR
                })
            });

            if (!tokenResponse.ok) {
                throw new Error(`Failed to generate upload token: ${tokenResponse.statusText}`);
            }

            const { token, upload_url } = await tokenResponse.json();

            // Step 2: Request presigned upload URL
            const presignedResponse = await fetch(upload_url, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    filename: file.name,
                    content_type: file.type || 'application/octet-stream',
                    size: file.size
                })
            });

            if (!presignedResponse.ok) {
                throw new Error(`Failed to get presigned URL: ${presignedResponse.statusText}`);
            }

            const { presigned_url, uuid } = await presignedResponse.json();

            // Step 3: Upload file directly to S3
            await this._uploadToS3(presigned_url, file, file.type, onProgress, options.signal);

            return {
                uuid,
                filename: file.name,
                size: file.size,
                contentType: file.type || 'application/octet-stream',
                token
            };

        } catch (error) {
            throw new Error(`Upload failed: ${error.message}`);
        }
    }

    /**
     * Download a file using token-based presigned URL
     * 
     * @param {string} artifactUuid - UUID of the artifact to download
     * @param {string} [filename] - Optional filename for the downloaded file
     * @param {Object} options - Download options
     * @param {number} [options.maxDownloads=1] - Maximum number of downloads allowed with this token
     * @param {string} [options.validFrom] - ISO 8601 timestamp when token becomes valid
     * @param {string} [options.validTo] - ISO 8601 timestamp when token expires
     * @param {string} [options.allowedCIDR] - CIDR notation for IP restriction
     * 
     * @returns {Promise<void>} Resolves when download completes
     * 
     * @example
     * await client.downloadFile('artifact-uuid-here', 'myfile.pdf', {
     *   maxDownloads: 3
     * });
     */
    async downloadFile(artifactUuid, filename = null, options = {}) {
        const {
            maxDownloads = 1,
            validFrom = null,
            validTo = null,
            allowedCIDR = null
        } = options;

        try {
            // Step 1: Generate download token
            const tokenResponse = await fetch(`${this.baseUrl}/genDownloadPresignedURL`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    artifact_uuid: artifactUuid,
                    max_downloads: maxDownloads,
                    valid_from: validFrom,
                    valid_to: validTo,
                    allowed_cidr: allowedCIDR
                })
            });

            if (!tokenResponse.ok) {
                throw new Error(`Failed to generate download token: ${tokenResponse.statusText}`);
            }

            const { presigned_url } = await tokenResponse.json();

            // Step 2: Download file (browser will follow 302 redirect automatically)
            const response = await fetch(presigned_url);

            if (!response.ok) {
                throw new Error(`Download failed: ${response.statusText}`);
            }

            // Step 3: Trigger browser download
            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = filename || artifactUuid;
            document.body.appendChild(a);
            a.click();
            window.URL.revokeObjectURL(url);
            document.body.removeChild(a);

        } catch (error) {
            throw new Error(`Download failed: ${error.message}`);
        }
    }

    /**
     * Get artifact metadata
     * 
     * @param {string} artifactUuid - UUID of the artifact
     * @returns {Promise<Object>} Artifact metadata
     * 
     * @example
     * const metadata = await client.getArtifactMetadata('artifact-uuid');
     * console.log(metadata.filename, metadata.size);
     */
    async getArtifactMetadata(artifactUuid) {
        const response = await fetch(`${this.baseUrl}/artifact-service/v1/artifacts/`);

        if (!response.ok) {
            throw new Error(`Failed to fetch artifacts: ${response.statusText}`);
        }

        const artifacts = await response.json();
        const artifact = artifacts.find(a => a.uuid === artifactUuid);

        if (!artifact) {
            throw new Error(`Artifact not found: ${artifactUuid}`);
        }

        return artifact;
    }

    /**
     * Internal method to upload file to S3 with progress tracking
     * @private
     */
    async _uploadToS3(presignedUrl, file, contentType, onProgress, signal) {
        return new Promise((resolve, reject) => {
            const xhr = new XMLHttpRequest();

            // Track upload progress
            if (onProgress) {
                xhr.upload.addEventListener('progress', (e) => {
                    if (e.lengthComputable) {
                        const percent = Math.round((e.loaded / e.total) * 100);
                        onProgress(percent);
                    }
                });
            }

            // Handle cancellation
            if (signal) {
                signal.addEventListener('abort', () => {
                    xhr.abort();
                    reject(new DOMException('Upload cancelled', 'AbortError'));
                });
            }

            xhr.addEventListener('load', () => {
                if (xhr.status === 200) {
                    resolve();
                } else {
                    reject(new Error(`S3 upload failed with status ${xhr.status}`));
                }
            });

            xhr.addEventListener('error', () => {
                reject(new Error('S3 upload failed'));
            });

            xhr.open('PUT', presignedUrl);
            xhr.setRequestHeader('Content-Type', contentType);
            xhr.send(file);
        });
    }

    /**
     * Create an upload token (Admin flow)
     * @param {Object} options - Token options
     * @returns {Promise<Object>} { token, upload_url, type }
     */
    async createUploadToken(options = {}) {
        const { maxUploads = 1, validFrom = null, validTo = null, allowedCIDR = null } = options;
        const response = await fetch(`${this.baseUrl}/genUploadPresignedURL`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                max_uploads: maxUploads,
                valid_from: validFrom,
                valid_to: validTo,
                allowed_cidr: allowedCIDR
            })
        });

        if (!response.ok) {
            throw new Error(`Failed to create upload token: ${response.statusText}`);
        }
        return await response.json();
    }

    /**
     * Create a download token (Admin flow)
     * @param {string} artifactUuid
     * @param {Object} options - Token options
     * @returns {Promise<Object>} { token, presigned_url, type }
     */
    async createDownloadToken(artifactUuid, options = {}) {
        const { maxDownloads = 1, validFrom = null, validTo = null, allowedCIDR = null } = options;
        const response = await fetch(`${this.baseUrl}/genDownloadPresignedURL`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                artifact_uuid: artifactUuid,
                max_downloads: maxDownloads,
                valid_from: validFrom,
                valid_to: validTo,
                allowed_cidr: allowedCIDR
            })
        });

        if (!response.ok) {
            throw new Error(`Failed to create download token: ${response.statusText}`);
        }
        return await response.json();
    }

    /**
     * Upload file using an existing token URL (End User flow)
     * @param {string} uploadTokenUrl - The full URL to submit file metadata provided by Admin
     * @param {File} file - The file object
     * @param {Function} [onProgress]
     * @param {Object} [options]
     * @param {AbortSignal} [options.signal] - Signal to abort the upload
     * @returns {Promise<Object>} Upload result { uuid, filename, size, status }
     */
    async uploadFileWithTokenUrl(uploadTokenUrl, file, onProgress = null, options = {}) {
        // Step 1: Request presigned S3 URL
        const presignedResponse = await fetch(uploadTokenUrl, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                filename: file.name,
                content_type: file.type || 'application/octet-stream',
                size: file.size
            }),
            signal: options.signal
        });

        if (!presignedResponse.ok) {
            throw new Error(`Failed to get presigned URL: ${presignedResponse.statusText}`);
        }

        const { presigned_url, uuid } = await presignedResponse.json();

        // Step 2: Upload direct to S3
        await this._uploadToS3(presigned_url, file, file.type, onProgress, options.signal);

        // Step 3: Notify server that upload is complete
        try {
            await this.completeUpload(uuid);
        } catch (err) {
            console.warn('Upload completion notification failed, server worker should handle it:', err);
        }

        return { uuid, filename: file.name, size: file.size };
    }

    /**
     * Notify server that upload is complete
     * @param {string} uuid - Artifact UUID
     */
    async completeUpload(uuid) {
        const response = await fetch(`${this.baseUrl}/artifact-service/v1/artifacts/${uuid}/complete`, {
            method: 'POST'
        });

        if (!response.ok) {
            throw new Error(`Failed to marks upload as complete: ${response.statusText}`);
        }

        return await response.json();
    }

    /**
     * Download file using an existing token URL (End User flow)
     * @param {string} downloadTokenUrl - The full URL provided by Admin
     * @param {string} [filename]
     */
    async downloadFileWithTokenUrl(downloadTokenUrl, filename = null) {
        // Step 1: Request download (follows redirect)
        const response = await fetch(downloadTokenUrl);
        if (!response.ok) {
            throw new Error(`Download failed: ${response.statusText}`);
        }

        // Step 2: Trigger browser download
        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename || 'downloaded-file';
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);
    }
}

// Export for use in modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = ArtifactClient;
}
