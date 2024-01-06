package joplin_provider

import (
	"context"

	"github.com/andrdru/joplin-auto/joplin_provider/entities"
)

type (
	provider interface {
		AcquireLock(ctx context.Context, id string) (err error)
		ReleaseLock(ctx context.Context, id string) (err error)
		ListNames(ctx context.Context) (list []string, err error)
		Get(ctx context.Context, name string) (file entities.File, err error)
		Put(ctx context.Context, file entities.File) (err error)
	}
)
