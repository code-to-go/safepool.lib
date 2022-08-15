package exchanges

import (
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sirupsen/logrus"
)

type S3Config struct {
	Region     string `json:"region" yaml:"region"`
	Endpoint   string `json:"endpoint" yaml:"endpoint"`
	Bucket     string `json:"bucket" yaml:"bucket"`
	AccessKey  string `json:"accessKey" yaml:"accessKey"`
	Secret     string `json:"secret" yaml:"secret"`
	DisableSSL bool   `json:"disableSSL" yaml:"disableSSL"`
}

type S3 struct {
	uploader *s3manager.Uploader
	svc      *s3.S3
	bucket   string
	url      string
}

func getAWSConfig(c S3Config) *aws.Config {
	s3c := aws.Config{}
	if c.Region != "" {
		s3c.Region = aws.String(c.Region)
	}
	if c.AccessKey != "" && c.Secret != "" {
		s3c.Credentials = credentials.NewStaticCredentials(
			c.AccessKey,
			c.Secret,
			"",
		)
	}
	if c.Endpoint != "" {
		s3c.Endpoint = aws.String(c.Endpoint)
	}
	s3c.DisableSSL = aws.Bool(c.DisableSSL)
	return &s3c
}

func NewS3(c S3Config) (Exchanger, error) {
	sess := session.New(getAWSConfig(c))
	url := fmt.Sprintf("s3://%s@%s/%s#region-%s", c.AccessKey, c.Endpoint, c.Bucket, c.Region)
	s := &S3{
		uploader: s3manager.NewUploader(sess),
		svc:      s3.New(sess),
		url:      url,
		bucket:   c.Bucket,
	}
	err := s.createBucketIfNeeded()
	return s, err
}

func (s *S3) createBucketIfNeeded() error {
	_, err := s.svc.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err == nil {
		return err
	}

	_, err = s.svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		logrus.Errorf("cannot create bucket %s: %v", s.bucket, err)
	}

	return err
}

func (s *S3) Read(name string, dest io.Writer) error {
	rawObject, err := s.svc.GetObject(
		&s3.GetObjectInput{
			Bucket: &s.bucket,
			Key:    &name,
		})
	if err != nil {
		logrus.Errorf("cannot read %s/%s: %v", s, name, err)
		return err
	}

	// b, err := io.ReadAll(rawObject.Body)
	// dest.Write(b)
	io.Copy(dest, rawObject.Body)
	// print(n)
	rawObject.Body.Close()
	return nil
}

func (s *S3) Write(name string, source io.Reader) error {

	_, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket: &s.bucket,
		Key:    &name,
		Body:   source,
	})
	if err != nil {
		logrus.Errorf("cannot write %s/%s: %v", s.String(), name, err)
	}
	return err
}

func (s *S3) Concat(name string, source []Source) error {
	return nil
}

func (s *S3) ReadDir(prefix string, opts ListOption) ([]fs.FileInfo, error) {
	input := &s3.ListObjectsInput{
		Bucket:    aws.String(s.bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	}

	result, err := s.svc.ListObjects(input)
	if err != nil {
		logrus.Errorf("cannot write %s/%s: %v", s.String(), prefix, err)
		return nil, err
	}

	var infos []fs.FileInfo
	for _, item := range result.Contents {
		cut := strings.LastIndex(prefix, "/")
		name := (*item.Key)[cut+1:]

		infos = append(infos, simpleFileInfo{
			name:    name,
			size:    *item.Size,
			isDir:   false,
			modTime: *item.LastModified,
		})

	}

	return infos, nil
}

func (s *S3) Stat(name string) (fs.FileInfo, error) {
	return nil, nil
}

func (s *S3) Delete(name string) error {
	_, err := s.svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &name,
	})
	return err
}

func (s *S3) Close() error {
	return nil
}

func (s *S3) String() string {
	return s.url
}
