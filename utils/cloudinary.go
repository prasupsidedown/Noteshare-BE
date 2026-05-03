// Fix: tambahkan ResourceType berdasarkan ekstensi file
package utils

import (
	"context"
	"fmt"
	"noteshare-be/config"
	"path/filepath"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

var cld *cloudinary.Cloudinary

func init() {
	var err error
	cld, err = cloudinary.NewFromURL(fmt.Sprintf("cloudinary://%s:%s@%s",
		config.AppConfig.CloudinaryAPIKey,
		config.AppConfig.CloudinaryAPISecret,
		config.AppConfig.CloudinaryCloudName,
	))
	if err != nil {
		fmt.Printf("Warning: Failed to initialize Cloudinary: %v\n", err)
	}
}

func getResourceType(filePath string) string {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg":
		return "image"
	case ".mp4", ".mov", ".avi", ".mkv":
		return "video"
	default:
		// PDF, DOC, DOCX, PPT, PPTX, TXT, dll
		return "raw"
	}
}

func UploadToCloudinary(ctx context.Context, filePath string, folder string) (string, string, error) {
	if cld == nil {
		return "", "", fmt.Errorf("cloudinary not initialized")
	}

	resourceType := getResourceType(filePath)

	resp, err := cld.Upload.Upload(ctx, filePath, uploader.UploadParams{
		Folder:       folder,
		ResourceType: resourceType,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to upload to cloudinary: %v", err)
	}

	return resp.SecureURL, resp.PublicID, nil
}

func DeleteFromCloudinary(ctx context.Context, publicID string) error {
	if cld == nil {
		return fmt.Errorf("cloudinary not initialized")
	}

	_, err := cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete from cloudinary: %v", err)
	}

	return nil
}