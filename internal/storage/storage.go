package storage

import (
	"context"
	"io"
)

type File struct {
	Name        string
	Size        int64
	ContentType string
	Content     io.Reader
}

type UploadedFile struct {
	Key string
	URL string
}

type ImageUploader interface {
	Upload(context.Context, File) (UploadedFile, error)
	Delete(context.Context, []string) error
}
