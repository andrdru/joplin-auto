package managers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/andrdru/joplin-auto/joplin_provider/entities"
)

type (
	joplinProvider interface {
		WaitForLockRealised(ctx context.Context, waitFor time.Duration) (err error)
		AcquireLock(ctx context.Context, id string) (err error)
		ReleaseLock(ctx context.Context, id string) (err error)
		ListNames(ctx context.Context, prefix *string) (list []string, err error)
		Get(ctx context.Context, name string) (file entities.File, err error)
		Put(ctx context.Context, file entities.File) (err error)
	}

	Todo struct {
		noteID   string
		parentID string
		appID    string

		provider         joplinProvider
		previousDataHash []byte
	}

	todoRecord struct {
		priority int
		data     []byte
	}
)

func NewTodo(provider joplinProvider, appID string, noteID string, parentID string) *Todo {
	return &Todo{
		appID:    appID,
		noteID:   noteID,
		parentID: parentID,
		provider: provider,
	}
}

func (t *Todo) UpdateTodo(ctx context.Context) error {
	err := t.provider.WaitForLockRealised(ctx, 30*time.Second)
	if err != nil {
		return fmt.Errorf("waitForLockRealised: %w", err)
	}

	err = t.provider.AcquireLock(ctx, t.appID)
	if err != nil {
		return fmt.Errorf("acquireLock: %w", err)
	}

	defer func() {
		err = t.provider.ReleaseLock(ctx, t.appID)
		if err != nil {
			slog.Default().Error("cannot realise lock", slog.String("err", err.Error()))
		}
	}()

	children, note, err := t.listChildren(ctx)
	if err != nil {
		return fmt.Errorf("listMDFiles: %w", err)
	}

	data := t.formatTodoData(children)

	err = t.updateNote(ctx, note, data)
	if err != nil {
		return fmt.Errorf("updateNote: %w", err)
	}

	return nil
}

func (t *Todo) listChildren(ctx context.Context) (list []entities.File, note entities.File, err error) {
	from, err := t.provider.ListNames(ctx, nil)
	if err != nil {
		return nil, entities.File{}, fmt.Errorf("list: %w", err)
	}

	var file entities.File
	for _, name := range from {
		if !strings.HasSuffix(name, ".md") {
			continue
		}

		file, err = t.provider.Get(ctx, name)
		if err != nil {
			return nil, entities.File{}, fmt.Errorf("get: %w", err)
		}

		if name == t.noteID+".md" {
			file.SplitRaw()
			note = file

			continue
		}

		if !t.isChild(file) {
			continue
		}

		file.SplitRaw()
		list = append(list, file)
	}

	return list, note, nil
}

func (t *Todo) isChild(file entities.File) bool {
	sub := []byte("parent_id: " + t.parentID)

	return bytes.Contains(file.Raw, sub)
}

func (t *Todo) formatTodoData(files []entities.File) (data []byte) {
	var list []todoRecord

	for _, el := range files {
		if el.MetaData("id") == t.noteID {
			continue
		}

		tmp := getMarkedTodos(el)
		if len(tmp) == 0 {
			continue
		}

		list = append(list, tmp...)
	}

	slices.SortFunc(list, func(a, b todoRecord) int { return -(a.priority - b.priority) })

	for _, el := range list {
		data = append(data, el.data...)
	}

	return data
}

func (t *Todo) updateNote(ctx context.Context, file entities.File, data []byte) (err error) {
	key := hashBytes(data)
	if slices.Equal(t.previousDataHash, key) {
		return nil
	}

	data = append(data, []byte("\n\nDO NOT EDIT\ngenerated at "+time.Now().Format(time.RFC3339))...)

	err = file.SetData(data)
	if err != nil {
		return fmt.Errorf("setData: %w", err)
	}

	err = t.provider.Put(ctx, file)
	if err != nil {
		return fmt.Errorf("put: %w", err)
	}

	t.previousDataHash = key

	return nil
}

func getMarkedTodos(file entities.File) (to []todoRecord) {
	parts := bytes.Split(file.Data, []byte("\n"))

	prefixes := [][]byte{
		[]byte("- [ ] !"),
		[]byte("- [x] !"),
	}

	for _, el := range parts {
		el = bytes.TrimSpace(el)

		for _, p := range prefixes {
			if !bytes.HasPrefix(el, p) {
				continue
			}

			el = bytes.TrimPrefix(el, p)
			priority := 1

			for i := range el {
				if el[i] == '!' {
					priority++
					continue
				}

				el = el[i:]
				break
			}

			b := bytes.NewBuffer(p[:len(p)-2])
			b.WriteString(" **")
			b.Write(el)
			b.WriteString("** (")
			b.Write(file.Header)
			b.WriteString(")\n")

			to = append(to,
				todoRecord{
					priority: priority,
					data:     b.Bytes(),
				},
			)

			break
		}
	}

	return to
}

func hashBytes(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}
