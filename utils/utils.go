package utils

import (
	"mime"
	"path/filepath"
	"strings"
)

func CleanPath(path string) string {
	path = strings.Split(path, "?")[0]
	path = strings.ReplaceAll(path, "\\", "/")
	return strings.ReplaceAll(path, "//", "/")
}

func GetMimeType(filePath string) string {
	ext := filepath.Ext(filePath)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		// maybe we should return an error in the future
		return ""
	}
	return mimeType
}
