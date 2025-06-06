package main

import (
	"fmt"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// getStringFromDict extracts string values from PDF dictionary objects
func getStringFromDict(dict types.Dict, key string) string {
	if obj, found := dict.Find(key); found && obj != nil {
		if str, ok := obj.(types.StringLiteral); ok {
			return str.Value()
		}
		if name, ok := obj.(types.Name); ok {
			return name.Value()
		}
		// Também tentar como HexLiteral
		if hex, ok := obj.(types.HexLiteral); ok {
			return hex.Value()
		}
	}
	return ""
}

// formatFileSize formats file size in human-readable format
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// boolToYesNo converts boolean to "Yes"/"No" string
func boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// printIfNotEmpty prints label: value only if value is not empty
func printIfNotEmpty(label, value string) {
	if value != "" {
		fmt.Printf("%s: %s\n", label, value)
	}
}
