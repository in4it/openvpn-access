package storage

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

type azBlob struct {
	StorageIf
	serviceURL azblob.ServiceURL
}

/*
 * cert
 */
func NewAzBlob(accountName, accountKey string) (StorageIf, error) {
	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return nil, err
	}
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})
	u, _ := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", accountName))
	serviceURL := azblob.NewServiceURL(*u, p)

	return &azBlob{
		serviceURL: serviceURL,
	}, nil
}

func (a *azBlob) HeadObject(bucket, item string) error {
	return nil
}
func (a *azBlob) GetObject(container, item string) (bytes.Buffer, error) {
	buffer := bytes.Buffer{}
	ctx := context.Background()
	containerURL := a.serviceURL.NewContainerURL(container)
	blobURL := containerURL.NewBlockBlobURL(item)
	get, err := blobURL.Download(ctx, 0, 0, azblob.BlobAccessConditions{}, false)
	if err != nil {
		return buffer, err
	}

	reader := get.Body(azblob.RetryReaderOptions{})
	buffer.ReadFrom(reader)
	reader.Close()

	return buffer, nil
}
func (a *azBlob) PutObject(container, item, data, kmsArn string) error {
	ctx := context.Background()
	containerURL := a.serviceURL.NewContainerURL(container)
	blobURL := containerURL.NewBlockBlobURL(item)
	_, err := blobURL.Upload(ctx, strings.NewReader(data), azblob.BlobHTTPHeaders{ContentType: "text/plain"}, azblob.Metadata{}, azblob.BlobAccessConditions{})
	if err != nil {
		return err
	}

	return nil
}
func (a *azBlob) DeleteObject(container, item string) error {
	ctx := context.Background()
	containerURL := a.serviceURL.NewContainerURL(container)
	blobURL := containerURL.NewBlockBlobURL(item)
	_, err := blobURL.Delete(ctx, azblob.DeleteSnapshotsOptionNone, azblob.BlobAccessConditions{})
	if err != nil {
		return err
	}
	return nil
}
