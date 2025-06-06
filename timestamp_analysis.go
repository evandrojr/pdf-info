package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// analyzeTimestamp detects and analyzes timestamp information in signatures
func (pa *PDFAnalyzer) analyzeTimestamp(filePath string, sigInfo *DigitalSignatureInfo) {
	// Initialize timestamp fields
	sigInfo.HasTimestamp = false
	sigInfo.TimestampType = ""
	sigInfo.TimestampTime = ""
	sigInfo.TimestampAuthority = ""
	sigInfo.TimestampStatus = "None"

	// Try to detect timestamp by analyzing raw PDF content
	hasTimestamp, timestampInfo := pa.detectTimestampByteAnalysis(filePath)
	if hasTimestamp {
		sigInfo.HasTimestamp = true
		sigInfo.TimestampType = timestampInfo["type"]
		sigInfo.TimestampTime = timestampInfo["time"]
		sigInfo.TimestampAuthority = timestampInfo["authority"]
		sigInfo.TimestampStatus = "Present"
	}
}

// detectTimestampByteAnalysis performs raw byte analysis for timestamp detection
func (pa *PDFAnalyzer) detectTimestampByteAnalysis(filePath string) (bool, map[string]string) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return false, nil
	}

	content := string(data)
	timestampInfo := make(map[string]string)
	
	// Look for timestamp-related patterns
	timestampPatterns := []string{
		"/SubFilter/ETSI.RFC3161",
		"/SubFilter/adbe.pkcs7.detached",
		"/M(D:",  // Timestamp marker
		"/TS",    // Timestamp token
		"1.2.840.113549.1.9.16.1.4",  // RFC3161 timestamp OID
		"TimeStampToken",
		"TSATimeStamp",
		"timestampToken",
		"/Type/TSA",
		"Serpro",  // Serpro timestamp authority
		"Assinador Serpro",  // Serpro signer/timestamp
		"ICP-Brasil",  // ICP-Brasil timestamp
		"AC Timestamping",  // Certificate Authority Timestamp
		"Carimbo",  // Portuguese for timestamp
		"D:20",  // Date timestamp format
		"/ByteRange",  // Signature byte range (often indicates timestamps)
	}

	hasTimestamp := false
	for _, pattern := range timestampPatterns {
		if strings.Contains(content, pattern) {
			hasTimestamp = true
			
			// Determine timestamp type based on pattern
			if strings.Contains(pattern, "RFC3161") {
				timestampInfo["type"] = "RFC3161"
			} else if strings.Contains(pattern, "adbe.pkcs7") {
				timestampInfo["type"] = "PKCS#7"
			} else if strings.Contains(pattern, "TSA") {
				timestampInfo["type"] = "TSA"
			} else if strings.Contains(pattern, "Serpro") || strings.Contains(pattern, "Assinador Serpro") {
				timestampInfo["type"] = "Serpro TSA"
			} else if strings.Contains(pattern, "ICP-Brasil") {
				timestampInfo["type"] = "ICP-Brasil"
			} else if strings.Contains(pattern, "ByteRange") {
				timestampInfo["type"] = "Embedded"
			} else {
				timestampInfo["type"] = "Standard"
			}
			break
		}
	}

	// Try to extract timestamp time if found
	if hasTimestamp {
		// Look for timestamp time patterns
		timePatterns := []string{
			"D:202", // Year 2020+
			"D:201", // Year 2010+
		}
		
		for _, timePattern := range timePatterns {
			if idx := strings.Index(content, timePattern); idx != -1 {
				// Extract timestamp date
				endIdx := idx + 20
				if endIdx > len(content) {
					endIdx = len(content)
				}
				timestampStr := content[idx:endIdx]
				
				// Clean up and format the timestamp string
				if len(timestampStr) >= 15 {
					// Format: D:YYYYMMDDHHmmSS -> YYYY-MM-DD HH:mm:SS
					rawTime := timestampStr
					if strings.HasPrefix(rawTime, "D:") && len(rawTime) >= 15 {
						year := rawTime[2:6]
						month := rawTime[6:8]
						day := rawTime[8:10]
						hour := rawTime[10:12]
						minute := rawTime[12:14]
						second := "00"
						if len(rawTime) >= 16 {
							second = rawTime[14:16]
						}
						timestampInfo["time"] = fmt.Sprintf("%s-%s-%s %s:%s:%s", year, month, day, hour, minute, second)
					} else {
						timestampInfo["time"] = rawTime[:15]
					}
				} else {
					timestampInfo["time"] = timestampStr
				}
				break
			}
		}
		
		// Try to detect timestamp authority
		authorityPatterns := []string{
			"CN=",
			"O=",
			"TSA",
			"TimeStamp",
			"Serpro",
			"Assinador Serpro",
			"ICP-Brasil",
			"AC Timestamping",
			"Autoridade Certificadora",
			"Certificate Authority",
		}
		
		for _, authPattern := range authorityPatterns {
			if idx := strings.Index(content, authPattern); idx != -1 {
				// Extract authority name
				endIdx := idx + 80
				if endIdx > len(content) {
					endIdx = len(content)
				}
				authStr := content[idx:endIdx]
				
				// Clean up authority string
				if newlineIdx := strings.Index(authStr, "\n"); newlineIdx != -1 {
					authStr = authStr[:newlineIdx]
				}
				if nullIdx := strings.Index(authStr, "\x00"); nullIdx != -1 {
					authStr = authStr[:nullIdx]
				}
				
				// Specific cleanup for known authorities
				if strings.Contains(authStr, "Serpro") || strings.Contains(authStr, "Assinador Serpro") {
					timestampInfo["authority"] = "Serpro - Serviço Federal de Processamento de Dados"
				} else if strings.Contains(authStr, "ICP-Brasil") {
					timestampInfo["authority"] = "ICP-Brasil Certificate Authority"
				} else {
					if len(authStr) > 40 {
						authStr = authStr[:40] + "..."
					}
					timestampInfo["authority"] = authStr
				}
				break
			}
		}
		
		// Set default values if not found
		if timestampInfo["time"] == "" {
			timestampInfo["time"] = "Unknown"
		}
		if timestampInfo["authority"] == "" {
			timestampInfo["authority"] = "Unknown"
		}
	}

	return hasTimestamp, timestampInfo
}
