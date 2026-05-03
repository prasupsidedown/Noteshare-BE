package utils

import (
	"context"
	"fmt"
	"noteshare-be/config"

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

// UploadToCloudinary uploads a file to Cloudinary and returns the URL and PublicID
func UploadToCloudinary(ctx context.Context, filePath string, folder string) (string, string, error) {
	if cld == nil {
		return "", "", fmt.Errorf("cloudinary not initialized")
	}

	resp, err := cld.Upload.Upload(ctx, filePath, uploader.UploadParams{
		Folder: folder,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to upload to cloudinary: %v", err)
	}

	return resp.SecureURL, resp.PublicID, nil
}

// DeleteFromCloudinary deletes a file from Cloudinary
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
