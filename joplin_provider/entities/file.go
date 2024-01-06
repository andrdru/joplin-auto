package entities

import (
	"bytes"
	"errors"
	"fmt"
	"time"
)

type (
	File struct {
		Provider string

		Name string
		Raw  []byte

		Header []byte
		Data   []byte
		Meta   [][]byte
	}
)

const (
	ProviderS3         = "s3"
	ProviderWebClipper = "web_clipper"
)

var (
	ErrUnexpectedMetadata = errors.New("unexpected metadata")
)

func (f *File) SplitRaw() {
	if f.Provider != ProviderS3 {
		return
	}

	parts := bytes.Split(f.Raw, []byte("\n"))

	f.Header = parts[0]

	if len(parts) < 2 {
		return
	}

	for i := 1; i < len(parts)-1; i++ {
		if !bytes.HasPrefix(parts[i], []byte("id: ")) {
			continue
		}

		f.Data = bytes.TrimSpace(bytes.Join(parts[1:i], []byte("\n")))
		f.Meta = parts[i:]
		break
	}
}

func (f *File) FormatRaw() {
	if f.Provider != ProviderS3 {
		return
	}

	buf := bytes.NewBuffer(f.Header)
	buf.WriteString("\n\n")
	buf.Write(f.Data)
	buf.WriteString("\n\n")
	buf.Write(bytes.Join(f.Meta, []byte("\n")))

	f.Raw = buf.Bytes()
}

func (f *File) MetaData(name string) (data string) {
	needle := []byte(name)

	if !bytes.HasSuffix(needle, []byte(": ")) {
		needle = append(needle, []byte(": ")...)
	}

	for _, el := range f.Meta {
		if !bytes.HasPrefix(el, needle) {
			continue
		}

		return string(el[len(needle):])
	}

	return ""
}

func (f *File) SetMetaData(name string, data string) (ok bool) {
	needle := []byte(name)
	if !bytes.HasSuffix(needle, []byte(": ")) {
		needle = append(needle, []byte(": ")...)
	}

	for i := range f.Meta {
		if !bytes.HasPrefix(f.Meta[i], needle) {
			continue
		}

		f.Meta[i] = append(needle, data...)
		return true
	}

	return false
}

func (f *File) SetData(data []byte) error {
	nowTime := time.Now().In(time.UTC).Format("2006-01-02T15:04:05.999Z07:00")

	f.Data = data

	if f.Provider != ProviderS3 {
		return nil
	}

	if !f.SetMetaData("updated_time", nowTime) {
		return fmt.Errorf("no updated_time: %w", ErrUnexpectedMetadata)
	}

	if !f.SetMetaData("user_updated_time", nowTime) {
		return fmt.Errorf("no user_updated_time: %w", ErrUnexpectedMetadata)
	}

	f.FormatRaw()

	return nil
}
