package dto

type UploadedImageResponse struct {
	ClientID string `json:"clientId"`
	Key      string `json:"key"`
	URL      string `json:"url"`
}

type UploadBatchResponse struct {
	Files []UploadedImageResponse `json:"files"`
}
