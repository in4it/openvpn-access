package storage

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type s3Struct struct {
	StorageIf
	sess *session.Session
}

/*
 * cert
 */
func NewS3() (StorageIf, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	return &s3Struct{sess: sess}, nil
}
func (s *s3Struct) HeadObject(bucket, item string) error {

	svc := s3.New(session.New())
	_, err := svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(item),
	})
	if err == nil {
		return nil
	} else {
		return err
	}
}

func (s *s3Struct) GetObject(bucket, item string) (bytes.Buffer, error) {
	var out bytes.Buffer
	var err error

	requestInput := s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(item),
	}

	buf := aws.NewWriteAtBuffer([]byte{})
	downloader := s3manager.NewDownloader(s.sess)
	_, err = downloader.Download(buf, &requestInput)
	if err != nil {
		return out, fmt.Errorf("S3 Download error %s", err)
	}

	if len(buf.Bytes()) == 0 {
		return out, fmt.Errorf("Unable to download item %q, bucket %q (0 bytes downloaded)", item, bucket)
	}

	_, err = io.Copy(&out, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return out, fmt.Errorf("Unable to output buffer")
	}
	return out, nil
}
func (s *s3Struct) PutObject(bucket, item, data, kmsArn string) error {
	reader := strings.NewReader(data)

	uploader := s3manager.NewUploader(s.sess)
	uploadInput := s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(item),
		Body:   reader,
	}
	if kmsArn != "" {
		uploadInput.SSEKMSKeyId = aws.String(kmsArn)
		uploadInput.ServerSideEncryption = aws.String("aws:kms")
	}
	_, err := uploader.Upload(&uploadInput)
	if err != nil {
		return fmt.Errorf("Unable to upload %q to %q, %v", item, bucket, err)
	}

	return nil
}
func (s *s3Struct) DeleteObject(bucket, item string) error {
	return fmt.Errorf("Not implemented")
}
