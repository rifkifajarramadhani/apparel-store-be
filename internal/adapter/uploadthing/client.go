package uploadthing

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/storage"
)

const apiURL = "https://api.uploadthing.com"

type tokenPayload struct {
	APIKey  string   `json:"apiKey"`
	AppID   string   `json:"appId"`
	Regions []string `json:"regions"`
}

type Client struct {
	apiKey     string
	appID      string
	httpClient *http.Client
}

func NewClient(token string, httpClient *http.Client) (*Client, error) {
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(token))
	if err != nil {
		return nil, fmt.Errorf("decode UploadThing token: %w", err)
	}

	var payload tokenPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil, fmt.Errorf("parse UploadThing token: %w", err)
	}

	if !strings.HasPrefix(payload.APIKey, "sk_") || payload.AppID == "" || len(payload.Regions) == 0 {
		return nil, fmt.Errorf("invalid UploadThing token payload")
	}

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{apiKey: payload.APIKey, appID: payload.AppID, httpClient: httpClient}, nil
}

func (c *Client) Upload(ctx context.Context, file storage.File) (storage.UploadedFile, error) {
	presigned, err := c.prepare(ctx, file)
	if err != nil {
		return storage.UploadedFile{}, err
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", file.Name)
	if err != nil {
		return storage.UploadedFile{}, fmt.Errorf("create upload body: %w", err)
	}

	if _, err := io.Copy(part, file.Content); err != nil {
		return storage.UploadedFile{}, fmt.Errorf("read upload file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return storage.UploadedFile{}, fmt.Errorf("close upload body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, presigned.URL, &body)
	if err != nil {
		return storage.UploadedFile{}, fmt.Errorf("create UploadThing upload request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := c.httpClient.Do(req)
	if err != nil {
		return storage.UploadedFile{}, fmt.Errorf("upload to UploadThing: %w", err)
	}

	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return storage.UploadedFile{}, fmt.Errorf("UploadThing upload failed (%d): %s", response.StatusCode, strings.TrimSpace(string(message)))
	}

	return storage.UploadedFile{
		Key: presigned.Key,
		URL: "https://" + c.appID + ".ufs.sh/f/" + url.PathEscape(presigned.Key),
	}, nil
}

type prepareResponse struct {
	Key string `json:"key"`
	URL string `json:"url"`
}

func (c *Client) prepare(ctx context.Context, file storage.File) (prepareResponse, error) {
	payload, err := json.Marshal(map[string]any{
		"fileName":           file.Name,
		"fileSize":           file.Size,
		"fileType":           file.ContentType,
		"contentDisposition": "inline",
		"acl":                "public-read",
	})
	if err != nil {
		return prepareResponse{}, fmt.Errorf("encode prepare upload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL+"/v7/prepareUpload", bytes.NewReader(payload))
	if err != nil {
		return prepareResponse{}, fmt.Errorf("create prepare upload request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-uploadthing-api-key", c.apiKey)

	response, err := c.httpClient.Do(req)
	if err != nil {
		return prepareResponse{}, fmt.Errorf("prepare UploadThing upload: %w", err)
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return prepareResponse{}, fmt.Errorf("prepare UploadThing upload failed (%d): %s", response.StatusCode, strings.TrimSpace(string(message)))
	}

	var result prepareResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return prepareResponse{}, fmt.Errorf("decode prepare upload response: %w", err)
	}

	return result, nil
}

func (c *Client) Delete(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	payload, err := json.Marshal(map[string]any{"fileKeys": keys})
	if err != nil {
		return fmt.Errorf("encode UploadThing deletion: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL+"/v6/deleteFiles", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create UploadThing deletion request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-uploadthing-api-key", c.apiKey)

	response, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete UploadThing files: %w", err)
	}

	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("delete UploadThing files failed (%d): %s", response.StatusCode, strings.TrimSpace(string(message)))
	}

	return nil
}
