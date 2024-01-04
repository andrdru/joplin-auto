package s3client

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type (
	S3Client struct {
		Bucket string

		s3sess *session.Session
		s3     *s3.S3
	}

	Config struct {
		Region   string
		Endpoint string
		Key      string
		Secret   string
		Bucket   string
	}

	File struct {
		Data []byte
		Name string
	}
)

const (
	stripeSize = 2 << 25 // 64Mb
)

func NewS3Client(cfg *Config) (*S3Client, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(cfg.Region),
		S3ForcePathStyle: aws.Bool(true),

		Endpoint:    aws.String(cfg.Endpoint),
		Credentials: credentials.NewStaticCredentials(cfg.Key, cfg.Secret, ""),
	})
	if err != nil {
		return nil, fmt.Errorf("new session: %w", err)
	}

	return &S3Client{
		s3sess: sess,
		s3:     s3.New(sess),
		Bucket: cfg.Bucket,
	}, nil
}

func (c *S3Client) File(ctx context.Context, name string) (file File, err error) {
	resp, err := c.s3.GetObjectWithContext(ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(c.Bucket),
			Key:    aws.String(name),
		},
	)

	if err != nil {
		return File{}, fmt.Errorf("GetObjectWithContext: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	file.Data, err = io.ReadAll(resp.Body)
	if err != nil {
		return File{}, fmt.Errorf("ReadAll: %w", err)
	}

	file.Name = name

	return file, nil
}

func (c *S3Client) ListNames(ctx context.Context, prefix *string, startAfter *string) (list []string, err error) {
	resp, err := c.s3.ListObjectsV2WithContext(ctx,
		&s3.ListObjectsV2Input{
			Bucket:     aws.String(c.Bucket),
			Prefix:     prefix,
			StartAfter: startAfter,
		})
	if err != nil {
		return nil, fmt.Errorf("ListObjectsV2WithContext: %w", err)
	}

	if len(resp.Contents) == 0 {
		return nil, nil
	}

	list = make([]string, 0, len(resp.Contents))
	for _, v := range resp.Contents {
		if v.Key == nil {
			continue
		}

		list = append(list, *v.Key)
	}

	return list, nil
}

func (c *S3Client) Upload(ctx context.Context, r io.Reader, fullFileName string) (err error) {
	u := s3manager.NewUploader(c.s3sess)

	_, err = u.UploadWithContext(ctx,
		&s3manager.UploadInput{
			Body:   r,
			Bucket: aws.String(c.Bucket),
			Key:    aws.String(fullFileName),
		},
		func(uploader *s3manager.Uploader) {
			uploader.PartSize = stripeSize
		},
	)

	if err != nil {
		return fmt.Errorf("UploadWithContext: %w", err)
	}

	return nil
}

func (c *S3Client) Delete(ctx context.Context, fullFileName string) (err error) {
	_, err = c.s3.DeleteObjectWithContext(ctx,
		&s3.DeleteObjectInput{
			Bucket: aws.String(c.Bucket),
			Key:    aws.String(fullFileName),
		})

	if err != nil {
		return fmt.Errorf("DeleteObjectWithContext: %w", err)
	}

	return nil
}
