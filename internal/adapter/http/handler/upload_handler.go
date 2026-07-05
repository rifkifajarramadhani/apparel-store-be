package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"mime/multipart"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/storage"
)

const (
	maxImageSize  = 8 * 1024 * 1024
	maxBatchFiles = 64
)

type UploadHandler struct {
	uploader storage.ImageUploader
	logger   *slog.Logger
}

func NewUploadHandler(uploader storage.ImageUploader, logger *slog.Logger) *UploadHandler {
	return &UploadHandler{uploader: uploader, logger: logger}
}

type uploadMetadata struct {
	ClientID string `json:"clientId"`
}

type uploadedImageResponse struct {
	ClientID string `json:"clientId"`
	Key      string `json:"key"`
	URL      string `json:"url"`
}

func (h *UploadHandler) ProductImages(c fiber.Ctx) error {
	if h.uploader == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "image uploads are not configured"})
	}

	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid multipart upload"})
	}

	files := form.File["files"]
	var metadata []uploadMetadata
	if err := json.Unmarshal([]byte(c.FormValue("metadata")), &metadata); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid upload metadata"})
	}

	if len(files) == 0 || len(files) != len(metadata) || len(files) > maxBatchFiles {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "files and metadata must contain the same number of entries"})
	}

	seen := make(map[string]struct{}, len(metadata))
	for i, file := range files {
		if strings.TrimSpace(metadata[i].ClientID) == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "every file requires a clientId"})
		}

		if _, exists := seen[metadata[i].ClientID]; exists {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "clientId values must be unique"})
		}

		seen[metadata[i].ClientID] = struct{}{}
		if err := validateImage(file); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
	}

	result := make([]uploadedImageResponse, 0, len(files))
	keys := make([]string, 0, len(files))
	for i, header := range files {
		file, err := header.Open()
		if err != nil {
			h.rollback(c.Context(), keys)
			return h.uploadError(c, fmt.Errorf("open %s: %w", header.Filename, err))
		}

		uploaded, uploadErr := h.uploader.Upload(c.Context(), storage.File{
			Name: header.Filename, Size: header.Size,
			ContentType: header.Header.Get("Content-Type"), Content: file,
		})
		closeErr := file.Close()
		if uploadErr != nil {
			h.rollback(c.Context(), keys)
			return h.uploadError(c, uploadErr)
		}

		if closeErr != nil {
			h.rollback(c.Context(), append(keys, uploaded.Key))
			return h.uploadError(c, closeErr)
		}

		keys = append(keys, uploaded.Key)
		result = append(result, uploadedImageResponse{
			ClientID: metadata[i].ClientID, Key: uploaded.Key, URL: uploaded.URL,
		})
	}

	return c.JSON(fiber.Map{"files": result})
}

func validateImage(file *multipart.FileHeader) error {
	contentType := strings.ToLower(file.Header.Get("Content-Type"))
	if !strings.HasPrefix(contentType, "image/") {
		return fmt.Errorf("%s is not an image", file.Filename)
	}

	if file.Size <= 0 || file.Size > maxImageSize {
		return fmt.Errorf("%s must be between 1 byte and 8 MB", file.Filename)
	}

	return nil
}

func (h *UploadHandler) rollback(ctx context.Context, keys []string) {
	if err := h.uploader.Delete(ctx, keys); err != nil {
		h.logger.ErrorContext(ctx, "failed to roll back UploadThing batch", "error", err, "keys", keys)
	}
}

func (h *UploadHandler) uploadError(c fiber.Ctx, err error) error {
	h.logger.ErrorContext(c.Context(), "product image upload failed", "error", err)
	return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "image upload failed"})
}
