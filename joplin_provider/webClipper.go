package joplin_provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/andrdru/joplin-auto/joplin_provider/entities"
	"github.com/andrdru/joplin-auto/joplin_provider/webClipperClient"
)

type WebClipper struct {
	client *webClipperClient.WebClipper
}

var _ provider = &WebClipper{}

func NewWebClipper(client *webClipperClient.WebClipper) *WebClipper {
	return &WebClipper{
		client: client,
	}
}

func (w *WebClipper) AcquireLock(ctx context.Context, id string) (err error) {
	// noop
	return nil
}

func (w *WebClipper) ReleaseLock(ctx context.Context, id string) (err error) {
	// noop
	return nil
}

// ListNames return file names in <id>.md format
func (w *WebClipper) ListNames(ctx context.Context) (list []string, err error) {
	var page = 1
	for {
		resp, err := w.client.List(ctx, page)
		if err != nil {
			return nil, fmt.Errorf("list: %w", err)
		}

		for _, el := range resp.Items {
			list = append(list, el.ID+".md")
		}

		if !resp.HasMore {
			break
		}

		page++
	}

	return list, nil
}

func (w *WebClipper) Get(ctx context.Context, name string) (file entities.File, err error) {
	id := strings.TrimSuffix(name, ".md")

	note, err := w.client.Get(ctx, id)
	if err != nil {
		return entities.File{}, fmt.Errorf("get: %w", err)
	}

	return entities.File{
		Provider: entities.ProviderWebClipper,
		Name:     name,
		Header:   []byte(note.Title),
		Data:     []byte(note.Body),
		Meta:     [][]byte{[]byte("id: " + note.ID), []byte("parent_id: " + note.ParentID)},
	}, nil
}

func (w *WebClipper) Put(ctx context.Context, file entities.File) (err error) {
	id := strings.TrimSuffix(file.Name, ".md")

	_, err = w.client.Put(ctx, id, string(file.Data))
	return err
}
