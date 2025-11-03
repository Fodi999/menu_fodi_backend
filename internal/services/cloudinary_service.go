package services

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// CloudinaryService сервис для работы с Cloudinary
type CloudinaryService struct {
	CloudName string
	APIKey    string
	APISecret string
	BaseURL   string
}

// CloudinaryUploadResponse ответ от Cloudinary при загрузке
type CloudinaryUploadResponse struct {
	PublicID     string `json:"public_id"`
	Version      int64  `json:"version"`
	Signature    string `json:"signature"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Format       string `json:"format"`
	ResourceType string `json:"resource_type"`
	CreatedAt    string `json:"created_at"`
	Bytes        int64  `json:"bytes"`
	Type         string `json:"type"`
	URL          string `json:"url"`
	SecureURL    string `json:"secure_url"`
}

// NewCloudinaryService создаёт новый сервис Cloudinary
func NewCloudinaryService() *CloudinaryService {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		log.Println("⚠️ Cloudinary credentials not configured")
	}

	return &CloudinaryService{
		CloudName: cloudName,
		APIKey:    apiKey,
		APISecret: apiSecret,
		BaseURL:   fmt.Sprintf("https://api.cloudinary.com/v1_1/%s", cloudName),
	}
}

// generateSignature генерирует подпись для Cloudinary signed upload
func (cs *CloudinaryService) generateSignature(params map[string]string) string {
	// Сортируем ключи
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Строим строку параметров
	var pairs []string
	for _, k := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, params[k]))
	}
	toSign := strings.Join(pairs, "&") + cs.APISecret

	// SHA1 hash
	h := sha1.New()
	h.Write([]byte(toSign))
	return hex.EncodeToString(h.Sum(nil))
}

// UploadImage загружает изображение в Cloudinary
func (cs *CloudinaryService) UploadImage(imageData []byte, fileName string) (*CloudinaryUploadResponse, error) {
	if cs.CloudName == "" {
		return nil, fmt.Errorf("cloudinary not configured")
	}

	log.Printf("[CLOUDINARY] 📸 Uploading image: %s (%d bytes)", fileName, len(imageData))

	// Генерируем параметры для подписи
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	publicID := uuid.New().String()
	folder := "culinary-academy"

	// Параметры для подписи (сортируются автоматически в generateSignature)
	signParams := map[string]string{
		"folder":    folder,
		"public_id": publicID,
		"timestamp": timestamp,
	}

	// Генерируем подпись
	signature := cs.generateSignature(signParams)

	// Создаём multipart форму
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Добавляем файл
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(imageData); err != nil {
		return nil, fmt.Errorf("failed to write file data: %w", err)
	}

	// Добавляем параметры загрузки (signed upload)
	writer.WriteField("api_key", cs.APIKey)
	writer.WriteField("timestamp", timestamp)
	writer.WriteField("signature", signature)
	writer.WriteField("folder", folder)
	writer.WriteField("public_id", publicID)

	writer.Close()

	// Отправляем запрос
	url := fmt.Sprintf("%s/image/upload", cs.BaseURL)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cloudinary error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var uploadResp CloudinaryUploadResponse
	if err := json.Unmarshal(respBody, &uploadResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	log.Printf("[CLOUDINARY] ✅ Image uploaded: %s -> %s", fileName, uploadResp.SecureURL)
	return &uploadResp, nil
}

// UploadImageFromURL загружает изображение по URL
func (cs *CloudinaryService) UploadImageFromURL(imageURL string) (*CloudinaryUploadResponse, error) {
	if cs.CloudName == "" {
		return nil, fmt.Errorf("cloudinary not configured")
	}

	log.Printf("[CLOUDINARY] 📡 Uploading from URL: %s", imageURL)

	// Генерируем параметры для подписи
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	publicID := uuid.New().String()
	folder := "culinary-academy"

	// Параметры для подписи
	signParams := map[string]string{
		"folder":    folder,
		"public_id": publicID,
		"timestamp": timestamp,
	}

	// Генерируем подпись
	signature := cs.generateSignature(signParams)

	// Создаём форму
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("file", imageURL)
	writer.WriteField("api_key", cs.APIKey)
	writer.WriteField("timestamp", timestamp)
	writer.WriteField("signature", signature)
	writer.WriteField("folder", folder)
	writer.WriteField("public_id", publicID)

	writer.Close()

	// Отправляем запрос
	url := fmt.Sprintf("%s/image/upload", cs.BaseURL)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cloudinary error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var uploadResp CloudinaryUploadResponse
	if err := json.Unmarshal(respBody, &uploadResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	log.Printf("[CLOUDINARY] ✅ Image uploaded from URL: %s", uploadResp.SecureURL)
	return &uploadResp, nil
}

// DeleteImage удаляет изображение из Cloudinary
func (cs *CloudinaryService) DeleteImage(publicID string) error {
	if cs.CloudName == "" {
		return fmt.Errorf("cloudinary not configured")
	}

	log.Printf("[CLOUDINARY] 🗑️ Deleting image: %s", publicID)

	url := fmt.Sprintf("%s/image/destroy", cs.BaseURL)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("public_id", publicID)
	writer.Close()

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetBasicAuth(cs.APIKey, cs.APISecret)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cloudinary error (status %d): %s", resp.StatusCode, string(respBody))
	}

	log.Printf("[CLOUDINARY] ✅ Image deleted: %s", publicID)
	return nil
}
