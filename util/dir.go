package util

import (
	"fmt"
	"os"
)

func CreateDir(path string, perm ...os.FileMode) error {
	var permission os.FileMode = 0755
	if len(perm) > 0 {
		permission = perm[0]
	}
	err := os.MkdirAll(path, permission)
	if err != nil {
		return fmt.Errorf("failed to create directory %q: %w", path, err)
	}
	return nil
}
