package storage

import "errors"

var (
	ErrCollectionIsEmpty  = errors.New("collection is empty")
	ErrCollectionNotFound = errors.New("collection not found")
	ErrCollectionExist    = errors.New("collection already exists")
)
