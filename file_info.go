package main

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// getFileInfo extracts basic file information
func (pa *PDFAnalyzer) getFileInfo(filePath string, info *PDFInfo) error {
	stat, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	info.FileName = filepath.Base(filePath)
	info.FilePath = filePath
	info.FileSize = stat.Size()
	info.FileSizeHuman = formatFileSize(stat.Size())
	info.LastModified = stat.ModTime()

	// Calcular hashes
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// MD5
	md5Hash := md5.New()
	if _, err := io.Copy(md5Hash, file); err != nil {
		return err
	}
	info.MD5Hash = fmt.Sprintf("%x", md5Hash.Sum(nil))

	// SHA256
	file.Seek(0, 0)
	sha256Hash := sha256.New()
	if _, err := io.Copy(sha256Hash, file); err != nil {
		return err
	}
	info.SHA256Hash = fmt.Sprintf("%x", sha256Hash.Sum(nil))

	return nil
}
