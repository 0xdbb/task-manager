package util

import (
	"path/filepath"
	"strings"
)

func FileName(path string) string {
	return filepath.Base(path)
}

func FileExt(path string) string {
	return filepath.Ext(FileName(path))
}

func FilenameWithoutExt(path string) string {
	filename := FileName(path)
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

func FileBasePathWithoutExt(path string) string {
	return strings.TrimSuffix(path, FileExt(path))
}
