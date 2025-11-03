package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/services"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
)

// UploadImageRequest запрос на загрузку изображения
type UploadImageRequest struct {
	ImageURL string `json:"imageUrl"` // URL изображения для загрузки
}

// UploadImageHandler POST /api/upload/image - загрузка изображения
func UploadImageHandler(w http.ResponseWriter, r *http.Request) {
	// Максимальный размер файла: 10MB
	r.ParseMultipartForm(10 << 20)

	cloudinary := services.NewCloudinaryService()

	// Проверяем, есть ли файл в multipart form
	file, header, err := r.FormFile("image")
	if err == nil {
		// Загружаем файл напрямую
		defer file.Close()

		imageData, err := io.ReadAll(file)
		if err != nil {
			utils.RespondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
				"status":  "error",
				"message": "Failed to read image file",
			})
			return
		}

		log.Printf("[UPLOAD] 📤 Uploading file: %s (%d bytes)", header.Filename, len(imageData))

		uploadResp, err := cloudinary.UploadImage(imageData, header.Filename)
		if err != nil {
			log.Printf("[UPLOAD] ❌ Error: %v", err)
			utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"status":  "error",
				"message": "Failed to upload image to Cloudinary",
				"error":   err.Error(),
			})
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"status": "ok",
			"data": map[string]interface{}{
				"url":       uploadResp.SecureURL,
				"publicId":  uploadResp.PublicID,
				"width":     uploadResp.Width,
				"height":    uploadResp.Height,
				"format":    uploadResp.Format,
				"bytes":     uploadResp.Bytes,
				"createdAt": uploadResp.CreatedAt,
			},
		})
		return
	}

	// Если файла нет, проверяем JSON с URL
	var req UploadImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Invalid request. Provide either 'image' file or 'imageUrl' in JSON",
		})
		return
	}

	if req.ImageURL == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "imageUrl is required",
		})
		return
	}

	log.Printf("[UPLOAD] 🌐 Uploading from URL: %s", req.ImageURL)

	uploadResp, err := cloudinary.UploadImageFromURL(req.ImageURL)
	if err != nil {
		log.Printf("[UPLOAD] ❌ Error: %v", err)
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Failed to upload image from URL",
			"error":   err.Error(),
		})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"data": map[string]interface{}{
			"url":       uploadResp.SecureURL,
			"publicId":  uploadResp.PublicID,
			"width":     uploadResp.Width,
			"height":    uploadResp.Height,
			"format":    uploadResp.Format,
			"bytes":     uploadResp.Bytes,
			"createdAt": uploadResp.CreatedAt,
		},
	})
}
