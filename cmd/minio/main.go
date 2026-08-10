package main

import (
	"bytes"
	"context"
	"log"

	"github.com/voidforge-studios/unlimit/storage"
)

func main() {
	// Initialize MinIO client
	// endpoint, username, password
	endpoint := "localhost:9000"
	username := "minioadmin"
	password := "minioadmin"
	collection := "my-collection"
	storer, err := storage.NewMinio(endpoint, username, password, false)
	if err != nil {
		log.Fatalf("Failed to create MinIO storer: %v", err)
	}

	ctx := context.Background()
	// _, err = storer.CreateCollection(ctx, &storage.CreateCollectionRequest{
	// 	Name: "my-collection",
	// })
	// if err != nil {
	// 	panic(err)
	// }

	// Use the storer for your operations
	data := []byte("your data here")
	res, err := storer.StoreAsset(ctx, &storage.StoreAssetRequest{
		// Fill in the request fields
		DataReader: bytes.NewReader(data),
		DataSize:   int64(len(data)),
		FileName:   "himom.txt",
		Collection: collection,
	})

	if err != nil {
		panic(err)
	}

	getRes, err := storer.GetAssetURL(ctx, &storage.GetAssetURLRequest{
		ID:         res.ID,
		Collection: collection,
	})
	if err != nil {
		panic(err)
	}

	log.Printf("Stored asset URL: %s", getRes.URL)
}
