package utils

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/onedayallday33-a11y/openshannon-go/internal/types"
)

// EncodeImage reads a file from disk and returns an ImageSource
func EncodeImage(path string) (*types.ImageSource, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read image file: %v", err)
	}

	return EncodeImageFromBytes(data, filepath.Ext(path))
}

// EncodeImageFromBytes takes raw bytes and returns an ImageSource
func EncodeImageFromBytes(data []byte, extension string) (*types.ImageSource, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("image data is empty")
	}

	// Detect MIME type
	mimeType := http.DetectContentType(data)
	
	// http.DetectContentType is sometimes generic, let's refine with extension if needed
	if mimeType == "application/octet-stream" || mimeType == "text/plain" {
		switch strings.ToLower(extension) {
		case ".jpg", ".jpeg":
			mimeType = "image/jpeg"
		case ".png":
			mimeType = "image/png"
		case ".gif":
			mimeType = "image/gif"
		case ".webp":
			mimeType = "image/webp"
		}
	}

	if !strings.HasPrefix(mimeType, "image/") {
		return nil, fmt.Errorf("unsupported or invalid image type: %s", mimeType)
	}

	encoded := base64.StdEncoding.EncodeToString(data)

	return &types.ImageSource{
		Type:      "base64",
		MediaType: mimeType,
		Data:      encoded,
	}, nil
}
