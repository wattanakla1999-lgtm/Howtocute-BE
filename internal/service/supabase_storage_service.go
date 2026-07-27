package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	neturl "net/url"
	"path"
	"strings"
	"time"
)

const maxSlipUploadBytes = 15 * 1024 * 1024

type SlipUploader interface {
	UploadSlip(ctx context.Context, bookingID uint, dataURL string) (string, error)
}

type ImageUploader interface {
	UploadImage(ctx context.Context, folder string, entityID uint, dataURL string) (string, error)
}

type SupabaseStorageClient struct {
	baseURL    string
	serviceKey string
	bucket     string
	httpClient *http.Client
}

func NewSupabaseStorageClient(baseURL, serviceKey, bucket string) *SupabaseStorageClient {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	serviceKey = strings.TrimSpace(serviceKey)
	bucket = strings.TrimSpace(bucket)
	if baseURL == "" || serviceKey == "" || bucket == "" {
		return nil
	}
	return &SupabaseStorageClient{
		baseURL:    baseURL,
		serviceKey: serviceKey,
		bucket:     bucket,
		httpClient: &http.Client{Timeout: 20 * time.Second},
	}
}

func (c *SupabaseStorageClient) UploadSlip(ctx context.Context, bookingID uint, dataURL string) (string, error) {
	mediaType, data, err := decodeDataURL(dataURL)
	if err != nil {
		return "", err
	}
	if len(data) > maxSlipUploadBytes {
		return "", fmt.Errorf("slip image exceeds %d bytes", maxSlipUploadBytes)
	}

	objectPath, err := slipObjectPath(bookingID, mediaType)
	if err != nil {
		return "", err
	}
	uploadURL := c.baseURL + "/storage/v1/object/" + escapePath(c.bucket) + "/" + escapePath(objectPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.serviceKey)
	req.Header.Set("apikey", c.serviceKey)
	req.Header.Set("Content-Type", mediaType)
	req.Header.Set("Cache-Control", "3600")
	req.Header.Set("x-upsert", "false")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("supabase storage upload failed: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return c.baseURL + "/storage/v1/object/public/" + escapePath(c.bucket) + "/" + escapePath(objectPath), nil
}

func (c *SupabaseStorageClient) UploadImage(ctx context.Context, folder string, entityID uint, dataURL string) (string, error) {
	mediaType, data, err := decodeDataURL(dataURL)
	if err != nil {
		return "", err
	}
	if len(data) > maxSlipUploadBytes {
		return "", fmt.Errorf("image exceeds %d bytes", maxSlipUploadBytes)
	}

	objectPath, err := imageObjectPath(folder, entityID, mediaType)
	if err != nil {
		return "", err
	}
	uploadURL := c.baseURL + "/storage/v1/object/" + escapePath(c.bucket) + "/" + escapePath(objectPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.serviceKey)
	req.Header.Set("apikey", c.serviceKey)
	req.Header.Set("Content-Type", mediaType)
	req.Header.Set("Cache-Control", "3600")
	req.Header.Set("x-upsert", "false")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("supabase storage upload failed: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return c.baseURL + "/storage/v1/object/public/" + escapePath(c.bucket) + "/" + escapePath(objectPath), nil
}

func decodeDataURL(value string) (string, []byte, error) {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "data:") {
		return "", nil, errors.New("slip image must be a data URL")
	}
	header, encoded, ok := strings.Cut(value, ",")
	if !ok || !strings.Contains(header, ";base64") {
		return "", nil, errors.New("slip image data URL must be base64 encoded")
	}
	mediaType := strings.TrimPrefix(strings.Split(header, ";")[0], "data:")
	if !strings.HasPrefix(mediaType, "image/") {
		return "", nil, errors.New("slip image must be an image data URL")
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", nil, err
	}
	return mediaType, decoded, nil
}

func slipObjectPath(bookingID uint, mediaType string) (string, error) {
	random := make([]byte, 6)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	extension := extensionForMediaType(mediaType)
	fileName := fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), hex.EncodeToString(random), extension)
	if bookingID == 0 {
		return path.Join("booking-slips", "pending", fileName), nil
	}
	return path.Join("booking-slips", fmt.Sprint(bookingID), fileName), nil
}

func imageObjectPath(folder string, entityID uint, mediaType string) (string, error) {
	random := make([]byte, 6)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	extension := extensionForMediaType(mediaType)
	fileName := fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), hex.EncodeToString(random), extension)
	folder = strings.Trim(strings.TrimSpace(folder), "/")
	if folder == "" {
		folder = "images"
	}
	if entityID == 0 {
		return path.Join(folder, "pending", fileName), nil
	}
	return path.Join(folder, fmt.Sprint(entityID), fileName), nil
}

func extensionForMediaType(mediaType string) string {
	switch mediaType {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	}
	extensions, err := mime.ExtensionsByType(mediaType)
	if err != nil || len(extensions) == 0 {
		return ".jpg"
	}
	return extensions[0]
}

func escapePath(value string) string {
	segments := strings.Split(value, "/")
	for i, segment := range segments {
		segments[i] = neturl.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}
