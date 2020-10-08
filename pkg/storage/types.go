package storage

import "bytes"

type StorageIf interface {
	HeadObject(bucket, item string) error
	GetObject(bucket, item string) (bytes.Buffer, error)
	PutObject(bucket, item, data, kmsArn string) error
}
