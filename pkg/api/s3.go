package api

import (
	"bytes"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type s3Struct struct {
	sess *session.Session
}

/*
 * cert
 */
func NewS3() *s3Struct {
	sess, _ := session.NewSession()
	return &s3Struct{sess: sess}
}

func (s *s3Struct) getObject(bucket, item string) (bytes.Buffer, error) {
	var out bytes.Buffer

	requestInput := s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(item),
	}

	buf := aws.NewWriteAtBuffer([]byte{})
	downloader := s3manager.NewDownloader(s.sess)
	downloader.Download(buf, &requestInput)

	if len(buf.Bytes()) == 0 {
		return out, fmt.Errorf("Unable to download item %q (0 bytes downloaded)", item)
	}

	_, err := io.Copy(&out, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return out, fmt.Errorf("Unable to output buffer")
	}
	return out, nil
}
