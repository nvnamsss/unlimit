package storage

import (
	"io"
	"time"
)

type CreateCollectionRequest struct {
	Name string `json:"name"`
}

type CreateCollectionResponse struct {
	Success bool `json:"success"`
}

type StoreAssetRequest struct {
	DataReader io.Reader `json:"-"`
	// DataSize specifies the size of the data to be stored in bytes.
	// For sizes smaller than 16MiB, the storage operation uses a single atomic PUT.
	// For sizes larger than 16MiB, the storage operation uses multipart upload.
	// Use -1 for unknown size, which will read until EOF (maximum 5TiB).
	// WARNING: Using -1 consumes more memory. Provide exact size when possible.
	DataSize int64  `json:"data_size"`
	FileName string `json:"file_name"`
	// ContentType specifies the MIME type of the asset (e.g., "image/png", "application/pdf").
	// If empty, MinIO will attempt to detect based on file extension.
	ContentType string `json:"content_type"`
	// ContentDisposition specifies the Content-Disposition header for downloads.
	// Use "attachment; filename=\"example.png\"" to force download with a specific filename.
	// If empty, the browser may display the file inline.
	ContentDisposition string `json:"content_disposition,omitempty"`
	// Collection is the logical grouping of assets.
	// Equivalent to Bucket in S3, Container in Azure Blob Storage, Pool in Ceph, Directory in Filesystem,
	// Namespace in OpenStack Swift, Share in SMB, Volume in NFS.
	Collection string `json:"collection"`
}

type StoreAssetResponse struct {
	ID          string `json:"id"`
	StoragePath string `json:"storage_path"`
}

type GetAssetURLRequest struct {
	ID         string `json:"id"`
	Collection string `json:"collection"`
	// ResponseContentDisposition overrides the Content-Disposition header in the response.
	// Use "attachment; filename=\"example.png\"" to force download with a specific filename.
	// This is useful for presigned URLs when you want to control the download behavior.
	ResponseContentDisposition string `json:"response_content_disposition,omitempty"`
	// ResponseContentType overrides the Content-Type header in the response.
	// Use this to correct or override the stored content type.
	ResponseContentType string `json:"response_content_type,omitempty"`
	// Expiry specifies how long the presigned URL should be valid.
	// If zero, defaults to 15 minutes.
	Expiry time.Duration `json:"expiry,omitempty"`
}

type GetAssetURLResponse struct {
	URL string `json:"url"`
}

type GetAssetRequest struct {
	ID         string `json:"id"`
	Collection string `json:"collection"`
	// VersionID specifies a specific version of the asset to retrieve.
	// If empty, the latest version is returned.
	// Only applicable for storage systems that support versioning.
	VersionID string `json:"version_id,omitempty"`
}

type GetAssetResponse struct {
	// Data is a reader for the asset content.
	// The caller is responsible for closing this reader when done.
	Data io.ReadCloser `json:"-"`
	// Size is the size of the asset in bytes.
	Size int64 `json:"size"`
	// ContentType is the MIME type of the asset.
	ContentType string `json:"content_type"`
}

type DeleteAssetRequest struct {
	ID          string `json:"id"`
	StoragePath string `json:"storage_path"`
}

type DeleteAssetResponse struct {
	Success bool `json:"success"`
}

type IsCollectionExistRequest struct {
	Name string `json:"name"`
}

type IsCollectionExistResponse struct {
	Exists bool `json:"exists"`
}
