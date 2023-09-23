package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"

	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Uploader struct {
	Context         context.Context
	s3Client        *s3.Client
	Bucket          string
	prefix          string
	localTrimPrefix string
}

type S3FileUploader interface {
	UploadFile(string, chan<- int)
}

func DefaultS3Client() *s3.Client {
	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("Unable to create AWS configuration: %s", err)
		return nil
	}
	cachedCredentialProvider := aws.NewCredentialsCache(cfg.Credentials)
	return s3.New(s3.Options{
		Credentials: cachedCredentialProvider,
		Region:      "us-west-2",
	})
}

/*
	NewS3Uploader

Create a new uploader where the bucketUrl is a valid, schemed, url like
s3://bucket/prefix/deeperPrefix
*/
func NewS3Uploader(s3Client *s3.Client, bucketUrl string) *S3Uploader {
	url, err := url.Parse(bucketUrl)
	if err != nil {
		log.Printf("%s isn't a valid S3 bucket URL: %s", bucketUrl, err)
		return nil
	}

	return &S3Uploader{
		Context:  context.TODO(),
		s3Client: s3Client,
		Bucket:   url.Host,
		prefix:   url.Path,
	}
}

/*
	TrimLocalPrefix: Set the local prefix to trim from video files

before uploading them
*/
func (u *S3Uploader) TrimLocalPrefix(prefix string) {
	u.localTrimPrefix = prefix
}

const (
	ErrorOpeningVideoFile = iota
	ErrorUploadingVideoFile
	StartUploadVideoFile
	DoneUploadVideoFile
)

func (u *S3Uploader) UploadFile(filepath string, status chan<- int) {
	if flagVerbose {
		log.Printf("Uploading %s", filepath)
	}
	body, err := os.Open(filepath)
	if err != nil {
		log.Printf("Error opening video file %s: %s", filepath, err)
		status <- ErrorOpeningVideoFile
		return
	}
	defer body.Close()

	sensorVideoPath := strings.TrimPrefix(filepath, u.localTrimPrefix)
	key := strings.TrimPrefix(fmt.Sprintf("%s/%s", u.prefix, sensorVideoPath), "/")
	input := &s3.PutObjectInput{
		Bucket:       &u.Bucket,
		Key:          &key,
		Body:         body,
		StorageClass: types.StorageClassIntelligentTiering,
	}
	status <- StartUploadVideoFile
	_, err = u.s3Client.PutObject(u.Context, input)
	if err != nil {
		log.Printf("Error uploading video file %s to %s: %s", filepath, key, err)
		status <- ErrorUploadingVideoFile
		return
	}
	if flagDebug {
		log.Printf("Uploaded %s to %s", filepath, key)
	}
	status <- DoneUploadVideoFile
}
