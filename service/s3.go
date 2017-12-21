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
	ListObjects(bucket string) (model.ObjectResponse, error)
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

func (s *s3Service) ListBuckets() (buckets []model.Bucket, err error) {
	response, bErr := s.s3.ListBuckets(nil)
	if bErr != nil {
		err = fmt.Errorf("failed to retrieve bucket information - Error: %s", err.Error())
		return
	}

	for _, bucket := range response.Buckets {
		buckets = append(buckets, model.NewBucket(bucket))
	}
	return
}

func (s *s3Service) ListObjects(bucket string) (objectResponse model.ObjectResponse, err error) {
	url, err := util.ParseAndValidateBucketURL(bucket)
	if err != nil {
		return
	}

	request := &s3.ListObjectsInput{
		Bucket: aws.String(url.Host),
	}

	if len(url.Path) == 0 {
		request.Delimiter = aws.String("/")
	}

	response, listErr := s.s3.ListObjects(request)
	if listErr != nil {
		err = fmt.Errorf("failed to retrieve objects - Error: %s", err.Error())
		return
	}

	folder := strings.TrimPrefix(url.Path, "/")
	objectResponse = convertToObjectResponse(response, folder)
	return
}

func convertToObjectResponse(output *s3.ListObjectsOutput, folderFilter string) model.ObjectResponse {
	var prefixes []model.Prefix
	for _, prefix := range output.CommonPrefixes {
		prefixes = append(prefixes, model.Prefix{Name: *prefix.Prefix})
	}

	folderFilterLength := len(folderFilter)
	if folderFilterLength != 0 {
		finalSlash := strings.LastIndex(folderFilter, "/") + 1
		if finalSlash != folderFilterLength {
			folderFilter += "/"
		}
	}

	var objects []model.Object
	for _, object := range output.Contents {
		if len(folderFilter) == 0 {
			objects = append(objects, model.NewObjectFromS3Object(object))
		} else {
			if strings.HasPrefix(*object.Key, folderFilter) {
				key := strings.TrimPrefix(*object.Key, folderFilter)
				if strings.Contains(key, "/") {
					fmt.Println("Directory")
				}


				objects = append(objects, model.NewObject(object.LastModified, *object.Size, key))
			}
		}
	}

	return model.ObjectResponse{
		Prefixes: prefixes,
		Objects:  objects,
	}
}
