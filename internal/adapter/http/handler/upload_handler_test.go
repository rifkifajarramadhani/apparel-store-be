package handler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/storage"
)

type uploaderFake struct {
	uploads int
	failAt  int
	deleted []string
}

func (f *uploaderFake) Upload(_ context.Context, file storage.File) (storage.UploadedFile, error) {
	f.uploads++
	if f.uploads == f.failAt {
		return storage.UploadedFile{}, errors.New("storage unavailable")
	}

	return storage.UploadedFile{Key: "key-" + file.Name, URL: "https://app.ufs.sh/f/key-" + file.Name}, nil
}

func (f *uploaderFake) Delete(_ context.Context, keys []string) error {
	f.deleted = append(f.deleted, keys...)
	return nil
}

func uploadRequest(t *testing.T, metadata string, files ...string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("metadata", metadata); err != nil {
		t.Fatal(err)
	}

	for _, name := range files {
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", `form-data; name="files"; filename="`+name+`"`)
		header.Set("Content-Type", "image/png")
		part, err := writer.CreatePart(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write([]byte("image")); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/upload", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}

func TestUploadHandlerProductImages(t *testing.T) {
	uploader := &uploaderFake{}
	app := fiber.New()
	app.Post("/upload", NewUploadHandler(uploader, slog.New(slog.NewTextHandler(io.Discard, nil))).ProductImages)

	response, err := app.Test(uploadRequest(t, `[{"clientId":"one"},{"clientId":"two"}]`, "one.png", "two.png"))
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != fiber.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("status = %d, body = %s", response.StatusCode, body)
	}

	body, _ := io.ReadAll(response.Body)
	if !strings.Contains(string(body), `"clientId":"one"`) || !strings.Contains(string(body), `"key":"key-two.png"`) {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestUploadHandlerRollsBackPartialBatch(t *testing.T) {
	uploader := &uploaderFake{failAt: 2}
	app := fiber.New()
	app.Post("/upload", NewUploadHandler(uploader, slog.New(slog.NewTextHandler(io.Discard, nil))).ProductImages)

	response, err := app.Test(uploadRequest(t, `[{"clientId":"one"},{"clientId":"two"}]`, "one.png", "two.png"))
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != fiber.StatusBadGateway {
		t.Fatalf("status = %d", response.StatusCode)
	}
	if len(uploader.deleted) != 1 || uploader.deleted[0] != "key-one.png" {
		t.Fatalf("deleted = %v", uploader.deleted)
	}
}
