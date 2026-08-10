package storage

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorer struct {
	client *minio.Client
}

func NewMinio(endpoint string, accessKey string, secretKey string, secure bool) (Storer, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
	})

	if err != nil {
		return nil, err
	}

	return &MinioStorer{
		client: client,
	}, nil
}

func (m *MinioStorer) IsCollectionExist(ctx context.Context, req *IsCollectionExistRequest) (*IsCollectionExistResponse, error) {
	if req.Name == "" {
		return nil, ErrCollectionIsEmpty
	}

	exists, err := m.client.BucketExists(ctx, req.Name)
	if err != nil {
		return nil, m.mapError(err)
	}

	return &IsCollectionExistResponse{
		Exists: exists,
	}, nil
}

func (m *MinioStorer) StoreAsset(ctx context.Context, req *StoreAssetRequest) (*StoreAssetResponse, error) {
	if req.DataSize == 0 {
		req.DataSize = -1
	}

	opts := minio.PutObjectOptions{}
	if req.ContentType != "" {
		opts.ContentType = req.ContentType
	}
	if req.ContentDisposition != "" {
		opts.ContentDisposition = req.ContentDisposition
	}

	info, err := m.client.PutObject(ctx, req.Collection, req.FileName, req.DataReader, req.DataSize, opts)

	if err != nil {
		return nil, m.mapError(err)
	}

	return &StoreAssetResponse{
		ID:          info.Key,
		StoragePath: info.Bucket + "/" + info.Key,
	}, nil
}

func (m *MinioStorer) GetAssetURL(ctx context.Context, req *GetAssetURLRequest) (*GetAssetURLResponse, error) {
	// Set default expiry to 15 minutes if not specified
	expiry := req.Expiry
	if expiry == 0 {
		expiry = time.Minute * 15
	}

	// Build request parameters for response header overrides
	reqParams := url.Values{}
	if req.ResponseContentDisposition != "" {
		reqParams.Set("response-content-disposition", req.ResponseContentDisposition)
	}
	if req.ResponseContentType != "" {
		reqParams.Set("response-content-type", req.ResponseContentType)
	}

	presignedURL, err := m.client.PresignedGetObject(ctx, req.Collection, req.ID, expiry, reqParams)
	if err != nil {
		return nil, m.mapError(err)
	}
	return &GetAssetURLResponse{
		URL: presignedURL.String(),
	}, nil
}

func (m *MinioStorer) GetAsset(ctx context.Context, req *GetAssetRequest) (*GetAssetResponse, error) {
	opts := minio.GetObjectOptions{}
	if req.VersionID != "" {
		opts.VersionID = req.VersionID
	}

	obj, err := m.client.GetObject(ctx, req.Collection, req.ID, opts)
	if err != nil {
		return nil, m.mapError(err)
	}

	// Get object info to retrieve size and content type
	info, err := obj.Stat()
	if err != nil {
		obj.Close()
		return nil, m.mapError(err)
	}

	return &GetAssetResponse{
		Data:        obj,
		Size:        info.Size,
		ContentType: info.ContentType,
	}, nil
}

func (m *MinioStorer) DeleteAsset(ctx context.Context, req *DeleteAssetRequest) (*DeleteAssetResponse, error) {
	// Parse collection and object ID from storage path if ID is empty
	collection := ""
	objectID := req.ID

	if req.StoragePath != "" {
		parts := strings.SplitN(req.StoragePath, "/", 2)
		if len(parts) == 2 {
			collection = parts[0]
			objectID = parts[1]
		}
	}

	if collection == "" {
		return nil, ErrCollectionIsEmpty
	}

	err := m.client.RemoveObject(ctx, collection, objectID, minio.RemoveObjectOptions{})
	if err != nil {
		return nil, m.mapError(err)
	}

	return &DeleteAssetResponse{
		Success: true,
	}, nil
}

func (m *MinioStorer) CreateCollection(ctx context.Context, req *CreateCollectionRequest) (*CreateCollectionResponse, error) {
	// TODO: Implement actual MinIO create collection logic

	// get collection
	if req.Name == "" {
		return nil, ErrCollectionIsEmpty
	}

	// check if collection already exists
	exists, err := m.client.BucketExists(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrCollectionExist
	}

	if err := m.client.MakeBucket(ctx, req.Name, minio.MakeBucketOptions{}); err != nil {
		return nil, err
	}

	return &CreateCollectionResponse{
		Success: true,
	}, nil
}

func (m *MinioStorer) mapError(err error) error {
	var (
		errRes minio.ErrorResponse
		ok     bool
	)

	errRes, ok = err.(minio.ErrorResponse)
	if !ok {
		return m.mapErrorString(err)
	}

	switch errRes.Code {
	case minio.NoSuchBucket:
		return ErrCollectionNotFound
	case minio.InvalidBucketName:
		return ErrCollectionIsEmpty
	}

	return err
}

func (m *MinioStorer) mapErrorString(err error) error {
	v := err.Error()
	if strings.Contains(v, "Bucket") {
		v = strings.ReplaceAll(v, "Bucket", "Collection")
		return errors.New(v)
	}

	return err
}
