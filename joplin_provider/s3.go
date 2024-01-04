package joplin_provider

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/andrdru/joplin-auto/joplin_provider/entities"
	"github.com/andrdru/joplin-auto/joplin_provider/s3client"
)

type (
	S3 struct {
		client *s3client.S3Client
	}
)

var (
	ErrOtherLockWasNotRealised = errors.New("other lock was not realised")
)

func NewS3(client *s3client.S3Client) *S3 {
	return &S3{
		client: client,
	}
}

// WaitForLockRealised wait for other locks realised
func (s3 *S3) WaitForLockRealised(ctx context.Context, waitFor time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, waitFor)
	defer cancel()

	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ErrOtherLockWasNotRealised
		case <-timer.C:

		}

		prefix := "locks/"
		list, err := s3.client.ListNames(ctx, &prefix, nil)
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}

		if len(list) == 0 {
			break
		}

		timer.Reset(1 * time.Second)
	}

	return nil
}

// AcquireLock exclusive lock file
// lock is exclusive, app set as joplin desktop
func (s3 *S3) AcquireLock(ctx context.Context, id string) (err error) {
	file := entities.File{
		Name: s3.lockFileName(id),
		Raw:  []byte(fmt.Sprintf(`{"type":2,"clientType":1,"clientId":"%s"}`, id)),
	}

	err = s3.Put(ctx, file)
	if err != nil {
		return fmt.Errorf("put: %w", err)
	}

	return nil
}

// ReleaseLock remove lock file
func (s3 *S3) ReleaseLock(ctx context.Context, id string) (err error) {
	return s3.client.Delete(ctx, s3.lockFileName(id))
}

// ListNames list all files names
func (s3 *S3) ListNames(ctx context.Context, prefix *string) (list []string, err error) {
	var startAfter *string

	for {
		tmp, err := s3.client.ListNames(ctx, prefix, startAfter)
		if err != nil {
			return nil, fmt.Errorf("list: %w", err)
		}

		if len(tmp) == 0 {
			break
		}

		list = append(list, tmp...)

		startAfter = &tmp[len(tmp)-1]
	}

	return list, nil
}

// Get load file contents
func (s3 *S3) Get(ctx context.Context, name string) (file entities.File, err error) {
	ret, err := s3.client.File(ctx, name)
	if err != nil {
		return entities.File{}, err
	}

	return entities.File{
		Name: ret.Name,
		Raw:  ret.Data,
	}, nil
}

// Put store file contents
func (s3 *S3) Put(ctx context.Context, file entities.File) (err error) {
	b := bytes.NewBuffer(file.Raw)

	return s3.client.Upload(ctx, b, file.Name)
}

func (s3 *S3) lockFileName(id string) string {
	return fmt.Sprintf("locks/2_1_%s.json", id)
}
