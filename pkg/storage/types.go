package storage

import "bytes"

//StorageIf implements an interface for the different types of supported storage
type StorageIf interface {
	HeadObject(bucket, item string) error
	GetObject(bucket, item string) (bytes.Buffer, error)
	PutObject(bucket, item, data, kmsArn string) error
	DeleteObject(bucket, item string) error
}
