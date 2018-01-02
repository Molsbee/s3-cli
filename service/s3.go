package service

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/molsbee/s3-cli/model"
	"github.com/molsbee/s3-cli/model/config"
	"github.com/molsbee/s3-cli/service/namedhandler"
	"github.com/molsbee/s3-cli/util"
	"strings"
)

type S3Service interface {
	ListBuckets() ([]model.Bucket, error)
	ListObjects(bucket string) (*model.ObjectResponse, error)
}

type s3Service struct {
	s3 *s3.S3
}

func NewS3Service(conf config.S3ServiceConfig) S3Service {
	creds := credentials.NewStaticCredentials(conf.AccessKey, conf.SecretAccessKey, "")
	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: creds,
		Endpoint:    aws.String(conf.Endpoint),
		Region:      aws.String(conf.Region),
	}))

	s3Serv := s3.New(sess)
	if conf.SignatureVersion == config.V2 {
		s3Serv.Handlers.Sign.Clear()
		s3Serv.Handlers.Sign.PushBackNamed(namedhandler.V2SignRequestHandler)
	}

	return &s3Service{s3Serv}
}

func (s *s3Service) ListBuckets() ([]model.Bucket, error) {
	response, err := s.s3.ListBuckets(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve bucket information - Error: %s", err.Error())
	}

	var buckets []model.Bucket
	for _, bucket := range response.Buckets {
		buckets = append(buckets, model.NewBucket(bucket))
	}
	return buckets, nil
}

func (s *s3Service) ListObjects(bucket string) (*model.ObjectResponse, error) {
	url, err := util.ParseAndValidateBucketURL(bucket)
	if err != nil {
		return nil, err
	}

	request := &s3.ListObjectsInput{
		Bucket: aws.String(url.Host),
		Delimiter: aws.String("/"),
	}

	prefix := strings.TrimPrefix(url.Path, "/")
	finalSlash := strings.LastIndex(prefix, "/") + 1
	if finalSlash != len(prefix) {
		prefix += "/"
	}

	if len(prefix) != 0 {
		request.Prefix = aws.String(prefix)
	}

	response, listErr := s.s3.ListObjects(request)
	if listErr != nil {
		return nil, fmt.Errorf("failed to retrieve objects - Error: %s", listErr.Error())
	}

	return convertToObjectResponse(response, prefix), nil
}

func convertToObjectResponse(output *s3.ListObjectsOutput, folderFilter string) *model.ObjectResponse {
	var prefixes []model.Prefix
	for _, prefix := range output.CommonPrefixes {
		directoryName := strings.TrimPrefix(*prefix.Prefix, folderFilter)
		prefixes = append(prefixes, model.Prefix{Name: directoryName})
	}

	var objects []model.Object
	for _, object := range output.Contents {
		fileName := strings.TrimPrefix(*object.Key, folderFilter)
		objects = append(objects, model.NewObject(object.LastModified, *object.Size, fileName))
	}

	return &model.ObjectResponse{
		Prefixes: prefixes,
		Objects:  objects,
	}
}
