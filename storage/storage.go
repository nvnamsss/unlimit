package storage

import (
	"context"
)

const (
	StorerTypeLocal = "local" // has not been implemented yet
	StorerTypeMinio = "minio"
)

// Storer interface provides a unified abstraction for object storage operations
// across different storage backends and cloud providers.
type Storer interface {
	// IsCollectionExist checks if a collection exists in the storage system.
	// Equivalent operations:
	// - S3: HeadBucket
	//
	// - Azure Blob Storage: GetContainerProperties
	//
	// - Google Cloud Storage: Bucket.Attrs()
	//
	// - MinIO: BucketExists
	//
	// - Filesystem: Check if directory exists
	//
	// - Ceph: Check if pool exists
	IsCollectionExist(ctx context.Context, req *IsCollectionExistRequest) (*IsCollectionExistResponse, error)

	// CreateCollection creates a new collection (logical grouping) in the storage system.
	// Equivalent operations:
	// - S3: CreateBucket
	//
	// - Azure Blob Storage: CreateContainer
	//
	// - Google Cloud Storage: CreateBucket
	//
	// - MinIO: MakeBucket
	//
	// - Filesystem: Create directory
	//
	// - Ceph: Create pool
	//
	// - OpenStack Swift: Create container
	CreateCollection(ctx context.Context, req *CreateCollectionRequest) (*CreateCollectionResponse, error)

	// StoreAsset uploads and stores an asset (file/object) in the specified collection.
	// Equivalent operations:
	// - S3: PutObject
	//
	// - Azure Blob Storage: UploadBlob
	//
	// - Google Cloud Storage: Object.NewWriter
	//
	// - MinIO: PutObject
	//
	// - Filesystem: Write file
	//
	// - Ceph: Put object
	//
	// - OpenStack Swift: Put object
	StoreAsset(ctx context.Context, req *StoreAssetRequest) (*StoreAssetResponse, error)

	// GetAssetURL generates a pre-signed or public URL for accessing an asset.
	// Equivalent operations:
	// - S3: PresignedGetObject / GetObjectURL
	//
	// - Azure Blob Storage: GetBlobURL / GenerateSASURL
	//
	// - Google Cloud Storage: SignedURL
	//
	// - MinIO: PresignedGetObject
	//
	// - Filesystem: Generate file path URL
	//
	// - Ceph: Generate object URL
	//
	// - OpenStack Swift: Get temp URL
	GetAssetURL(ctx context.Context, req *GetAssetURLRequest) (*GetAssetURLResponse, error)

	// GetAsset downloads an asset from the storage system and returns its content.
	// Equivalent operations:
	// - S3: GetObject
	//
	// - Azure Blob Storage: DownloadBlob
	//
	// - Google Cloud Storage: Object.NewReader
	//
	// - MinIO: GetObject
	//
	// - Filesystem: Read file
	//
	// - Ceph: Get object
	//
	// - OpenStack Swift: Get object
	GetAsset(ctx context.Context, req *GetAssetRequest) (*GetAssetResponse, error)

	// DeleteAsset removes an asset from the storage system.
	// Equivalent operations:
	// - S3: DeleteObject
	//
	// - Azure Blob Storage: DeleteBlob
	//
	// - Google Cloud Storage: Object.Delete
	//
	// - MinIO: RemoveObject
	//
	// - Filesystem: Delete file
	//
	// - Ceph: Delete object
	//
	// - OpenStack Swift: Delete object
	DeleteAsset(ctx context.Context, req *DeleteAssetRequest) (*DeleteAssetResponse, error)
}
