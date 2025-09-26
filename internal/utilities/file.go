package utilities

import (
	"errors"
	"os"
	"path/filepath"
)

func MkDirParent(destFilePath string) (*os.File, error) {
	_, err := os.Stat(destFilePath)
	if errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(filepath.Dir(destFilePath), os.ModePerm)
		if err != nil {
			return nil, err
		}
	}
	destFile, err := os.Create(destFilePath)
	if err != nil {
		return nil, err
	}
	return destFile, nil
}
