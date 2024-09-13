package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"golang.org/x/mod/sumdb/dirhash"
	"io"
	"os"
)

func ComputeHash(t interface{}) (string, error) {
	data, err := json.Marshal(t)
	if err != nil {
		return "", err
	}

	h := sha256.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil)), nil
}

func ComputeDirHash(path string) (string, error) {
	return dirhash.HashDir(path, "", dirhash.DefaultHash)
}

func ComputeFileHash(filePath string) (string, error) {
	return dirhash.Hash1([]string{filePath}, func(name string) (io.ReadCloser, error) {
		return os.Open(name)
	})
}
