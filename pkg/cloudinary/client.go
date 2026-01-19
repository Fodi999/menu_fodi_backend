package cloudinary

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// Client wraps Cloudinary SDK with recipe-specific logic
type Client struct {
	cld *cloudinary.Cloudinary
}

// NewClient creates a Cloudinary client from environment variables
func NewClient() (*Client, error) {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("cloudinary credentials not set in environment")
	}

	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cloudinary: %w", err)
	}

	return &Client{cld: cld}, nil
}

// UploadRecipeImage uploads a recipe image to Cloudinary
// Returns secure URL and public ID for database storage
func (c *Client) UploadRecipeImage(ctx context.Context, file multipart.File, recipeID string) (*UploadResult, error) {
	// Upload with recipe-specific folder and transformation
	uploadResult, err := c.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:         "recipes",                          // Organized storage
		PublicID:       fmt.Sprintf("recipe_%s", recipeID), // Unique identifier
		Overwrite:      boolPtr(true),                      // Replace if exists
		Transformation: "c_fill,w_1200,h_800,q_auto",       // Optimize: 1200x800, auto quality
		Format:         "webp",                             // Modern format for better compression
		ResourceType:   "image",
	})

	if err != nil {
		return nil, fmt.Errorf("cloudinary upload failed: %w", err)
	}

	return &UploadResult{
		SecureURL: uploadResult.SecureURL,
		PublicID:  uploadResult.PublicID,
		Width:     uploadResult.Width,
		Height:    uploadResult.Height,
		Format:    uploadResult.Format,
		Bytes:     uploadResult.Bytes,
	}, nil
}

// DeleteImage removes an image from Cloudinary
func (c *Client) DeleteImage(ctx context.Context, publicID string) error {
	if publicID == "" {
		return nil // Nothing to delete
	}

	_, err := c.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: "image",
	})

	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	return nil
}

// GenerateThumbnailURL creates a thumbnail URL with Cloudinary transformations
// Example: width=400, height=300, crop=fill
func (c *Client) GenerateThumbnailURL(publicID string, width, height int) string {
	if publicID == "" {
		return ""
	}

	return fmt.Sprintf(
		"https://res.cloudinary.com/%s/image/upload/c_fill,w_%d,h_%d,q_auto,f_webp/%s",
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		width,
		height,
		publicID,
	)
}

// UploadResult contains the result of a successful upload
type UploadResult struct {
	SecureURL string // CDN URL (https://res.cloudinary.com/...)
	PublicID  string // Cloudinary identifier for future operations
	Width     int
	Height    int
	Format    string
	Bytes     int
}

// Helper function to create bool pointer
func boolPtr(b bool) *bool {
	return &b
}
