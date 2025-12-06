package utils

import "os"

func ReadAll(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return content, nil
}
