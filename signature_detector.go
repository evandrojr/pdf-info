package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// RawSignatureDetector uses byte-level analysis to detect signatures
type RawSignatureDetector struct{}

func NewRawSignatureDetector() *RawSignatureDetector {
	return &RawSignatureDetector{}
}

// DetectSignaturesByteAnalysis performs raw byte analysis for signature detection
func (r *RawSignatureDetector) DetectSignaturesByteAnalysis(filePath string) (bool, int, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return false, 0, err
	}
	
	content := string(data)
	signatureCount := 0
	
	// Look for signature-related patterns in the raw PDF content
	patterns := []string{
		"/Type/Sig",
		"/FT/Sig", 
		"/SigFlags",
		"Adobe.PPKLite",
		"Adobe.PPKMS",
		"PKCS#7",
		"pkcs7",
		"/ByteRange",
		"/Contents<",
		"/SubFilter/adbe.pkcs7.detached",
		"/SubFilter/adbe.pkcs7.sha1",
		"/SubFilter/ETSI.CAdES.detached",
		"/Filter/Adobe.PPKLite",
		"/Filter/Adobe.PPKMS",
	}
	
	fmt.Printf("Debug: Analyzing %d bytes for signature patterns\n", len(data))
	
	for _, pattern := range patterns {
		count := strings.Count(content, pattern)
		if count > 0 {
			fmt.Printf("Debug: Found %d occurrences of pattern: %s\n", count, pattern)
			if pattern == "/Type/Sig" || pattern == "/FT/Sig" {
				signatureCount += count
			} else if signatureCount == 0 {
				signatureCount = 1 // At least one signature indicated
			}
		}
	}
	
	// Additional heuristics for encrypted PDFs
	if signatureCount == 0 {
		// Look for signature dictionaries even in encrypted content
		sigDictPatterns := []string{
			"/Sig",
			"<</Type/Sig",
			"<</FT/Sig",
		}
		
		for _, pattern := range sigDictPatterns {
			if strings.Contains(content, pattern) {
				fmt.Printf("Debug: Found signature dictionary pattern: %s\n", pattern)
				signatureCount = 1
				break
			}
		}
	}
	
	fmt.Printf("Debug: Raw byte analysis found %d signatures\n", signatureCount)
	return signatureCount > 0, signatureCount, nil
}
