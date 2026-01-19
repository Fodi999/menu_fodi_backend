package http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/cloudinary"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/go-chi/chi/v5"
)

// UploadRecipeImage handles image upload for a recipe
// POST /api/admin/recipes/:id/image
func (h *AdminHandlers) UploadRecipeImage(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")
	if recipeID == "" {
		utils.WriteError(w, http.StatusBadRequest, "Recipe ID is required")
		return
	}

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	// Get file from form
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "File is required")
		return
	}
	defer file.Close()

	// Validate file size (max 5MB)
	if fileHeader.Size > 5*1024*1024 {
		utils.WriteError(w, http.StatusBadRequest, "File too large (max 5MB)")
		return
	}

	// Validate file type
	contentType := fileHeader.Header.Get("Content-Type")
	allowedTypes := []string{"image/jpeg", "image/png", "image/webp"}
	if !contains(allowedTypes, contentType) {
		utils.WriteError(w, http.StatusBadRequest, "Invalid file type. Allowed: JPEG, PNG, WebP")
		return
	}

	// Check if recipe exists
	var recipe models.Recipe
	if err := h.service.DB().Where("id = ?", recipeID).First(&recipe).Error; err != nil {
		utils.WriteError(w, http.StatusNotFound, "Recipe not found")
		return
	}

	// Initialize Cloudinary client
	cldClient, err := cloudinary.NewClient()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to initialize image service")
		return
	}

	// Delete old image if exists (before uploading new one)
	// Note: We don't fail the request if old image deletion fails
	// because the new image will overwrite it via Overwrite:true in UploadParams
	oldImagePublicId := recipe.ImagePublicId
	if oldImagePublicId != "" {
		if err := cldClient.DeleteImage(r.Context(), oldImagePublicId); err != nil {
			// Log warning but continue - Cloudinary will overwrite
			fmt.Printf("WARNING: Failed to delete old image (will be overwritten): PublicID=%s, Error=%v\n", oldImagePublicId, err)
		}
	}

	// Upload to Cloudinary
	uploadResult, err := cldClient.UploadRecipeImage(r.Context(), file, recipeID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Image upload failed: %v", err))
		return
	}

	// Update recipe in database
	recipe.ImageUrl = uploadResult.SecureURL
	recipe.ImagePublicId = uploadResult.PublicID
	if err := h.service.DB().Save(&recipe).Error; err != nil {
		// CRITICAL: Transactional integrity - cleanup uploaded image
		// Cloudinary upload succeeded but DB save failed
		// Must delete orphaned image to prevent storage bloat
		if cleanupErr := cldClient.DeleteImage(r.Context(), uploadResult.PublicID); cleanupErr != nil {
			// Log cleanup failure but don't expose to user
			fmt.Printf("CRITICAL: Failed to cleanup orphaned image after DB save failure. PublicID: %s, Error: %v\n", uploadResult.PublicID, cleanupErr)
		}
		utils.WriteError(w, http.StatusInternalServerError, "Failed to save image URL to database")
		return
	}

	// Generate thumbnail URLs for response
	thumbnails := map[string]string{
		"small":  cldClient.GenerateThumbnailURL(uploadResult.PublicID, 200, 150),
		"medium": cldClient.GenerateThumbnailURL(uploadResult.PublicID, 400, 300),
		"large":  cldClient.GenerateThumbnailURL(uploadResult.PublicID, 800, 600),
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"imageUrl":   uploadResult.SecureURL,
			"publicId":   uploadResult.PublicID,
			"width":      uploadResult.Width,
			"height":     uploadResult.Height,
			"format":     uploadResult.Format,
			"size":       uploadResult.Bytes,
			"thumbnails": thumbnails,
		},
	})
}

// DeleteRecipeImage removes the image from a recipe
// DELETE /api/admin/recipes/:id/image
func (h *AdminHandlers) DeleteRecipeImage(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")
	if recipeID == "" {
		utils.WriteError(w, http.StatusBadRequest, "Recipe ID is required")
		return
	}

	// Check if recipe exists
	var recipe models.Recipe
	if err := h.service.DB().Where("id = ?", recipeID).First(&recipe).Error; err != nil {
		utils.WriteError(w, http.StatusNotFound, "Recipe not found")
		return
	}

	if recipe.ImagePublicId == "" {
		utils.WriteError(w, http.StatusBadRequest, "Recipe has no image")
		return
	}

	// Initialize Cloudinary client
	cldClient, err := cloudinary.NewClient()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to initialize image service")
		return
	}

	// Delete from Cloudinary
	if err := cldClient.DeleteImage(r.Context(), recipe.ImagePublicId); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete image: %v", err))
		return
	}

	// Update recipe in database
	recipe.ImageUrl = ""
	recipe.ImagePublicId = ""
	if err := h.service.DB().Save(&recipe).Error; err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to update recipe")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Image deleted successfully",
	})
}

// Helper function to check if slice contains a value
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}
