package db

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func uploadFileToS3(bucket, item, filePath string) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-north-1"),
	})
	if err != nil {
		return fmt.Errorf("failed to create session, %v", err)
	}

	uploader := s3manager.NewUploader(sess)

	file, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %q, %v", filePath, err)
	}

	// Upload the file to S3.
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(item),
		Body:   bytes.NewReader(file),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}

	log.Printf("file uploaded to, %s/%s\n", bucket, item)
	return nil
}
