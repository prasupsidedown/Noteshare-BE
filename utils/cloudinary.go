package utils

import (
	"context"
	"fmt"
	"log"
	"noteshare-be/config"
	"os"
	"path/filepath"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func getCloudinary() (*cloudinary.Cloudinary, error) {
	cld, err := cloudinary.NewFromURL(fmt.Sprintf("cloudinary://%s:%s@%s",
		config.AppConfig.CloudinaryAPIKey,
		config.AppConfig.CloudinaryAPISecret,
		config.AppConfig.CloudinaryCloudName,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cloudinary: %v", err)
	}
	return cld, nil
}

func getResourceType(filePath string) string {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg":
		return "image"
	case ".mp4", ".mov", ".avi", ".mkv":
		return "video"
	default:
		return "raw"
	}
}

func UploadToCloudinary(ctx context.Context, filePath string, folder string) (string, string, error) {
	client, err := getCloudinary()
	if err != nil {
		return "", "", err
	}

	// Buka file sebagai io.Reader — fix untuk Windows path
	f, err := os.Open(filePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to open file: %v", err)
	}
	defer f.Close()

	resourceType := getResourceType(filePath)
	log.Printf("Uploading to Cloudinary — resourceType: %q", resourceType)

	resp, err := client.Upload.Upload(ctx, f, uploader.UploadParams{
		Folder:       folder,
		ResourceType: resourceType,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to upload to cloudinary: %v", err)
	}

	if resp.Error.Message != "" {
		return "", "", fmt.Errorf("cloudinary error: %s", resp.Error.Message)
	}

	log.Printf("Cloudinary upload success — SecureURL: %q, PublicID: %q", resp.SecureURL, resp.PublicID)

	return resp.SecureURL, resp.PublicID, nil
}

func DeleteFromCloudinary(ctx context.Context, publicID string) error {
	client, err := getCloudinary()
	if err != nil {
		return err
	}

	_, err = client.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete from cloudinary: %v", err)
	}

	return nil
}