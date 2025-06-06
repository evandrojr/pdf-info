package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// analyzeDigitalSignatures analyzes digital signatures in the PDF
func (pa *PDFAnalyzer) analyzeDigitalSignatures(filePath string, ctx *model.Context, info *PDFInfo) {
	// First, try to detect signature fields directly from the PDF structure
	// This works even for encrypted PDFs in many cases
	hasSignatureFields := pa.detectSignatureFields(ctx, info)
	
	// If structural analysis fails, try raw byte analysis
	if !hasSignatureFields {
		hasRawSignatures, rawCount, err := pa.detectSignaturesByteAnalysis(filePath)
		if err != nil {
			// Silent error - continue with no signatures detected
		} else if hasRawSignatures {
			info.HasDigitalSignatures = true
			info.SignatureCount = rawCount
			hasSignatureFields = true
		}
	}
	
	// Try to validate signatures using pdfcpu (this may fail for encrypted PDFs)
	results, err := api.ValidateSignatures(filePath, true, nil) // all=true
	if err != nil {
		fmt.Printf("Warning: error validating signatures: %v", err)
		// If validation fails but we detected signature fields, still report them
		if hasSignatureFields {
			info.HasDigitalSignatures = true
			// Keep the signature count from field detection
			return
		}
		info.HasDigitalSignatures = false
		info.SignatureCount = 0
		return
	}

	if len(results) == 0 {
		// If no validation results but we found signature fields, report the fields
		if hasSignatureFields {
			info.HasDigitalSignatures = true
			// Keep the signature count from field detection
			return
		}
		info.HasDigitalSignatures = false
		info.SignatureCount = 0
		return
	}

	// We have successful validation results
	info.HasDigitalSignatures = true
	info.SignatureCount = len(results)
	info.Signatures = make([]DigitalSignatureInfo, 0, len(results))

	// Process each validation result
	for _, result := range results {
		sigInfo := DigitalSignatureInfo{
			FieldName:   result.Details.FieldName,
			SubFilter:   result.Details.SubFilter,
			SignerName:  result.Details.SignerName,
			SigningTime: formatTime(result.Details.SigningTime),
			Location:    result.Details.Location,
			Reason:      result.Details.Reason,
			ContactInfo: result.Details.ContactInfo,
			IsCertified: result.Certified(),
			IsValid:     result.Status == 1, // SignatureStatusValid
		}

		// Determine signature status
		switch result.Status {
		case 1: // model.SignatureStatusValid
			sigInfo.Status = "Valid"
		case 2: // model.SignatureStatusInvalid
			sigInfo.Status = "Invalid"
		default: // model.SignatureStatusUnknown or others
			sigInfo.Status = "Unknown"
		}

		// Add validation errors if any exist
		if len(result.Problems) > 0 {
			sigInfo.ValidationErrors = result.Problems
		}

		// Determine signature type
		if result.Certified() {
			sigInfo.Type = "Certified"
		} else {
			sigInfo.Type = "Approval"
		}

		// Analyze timestamp information
		pa.analyzeTimestamp(filePath, &sigInfo)

		info.Signatures = append(info.Signatures, sigInfo)
	}
}

// detectSignatureFields detects signature fields in the PDF structure
func (pa *PDFAnalyzer) detectSignatureFields(ctx *model.Context, info *PDFInfo) bool {
	if ctx == nil || ctx.RootDict == nil {
		return false
	}
	
	signatureFound := false
	
	// Method 1: Look for AcroForm dictionary
	acroFormObj, found, _ := ctx.RootDict.Entry("AcroForm", "", false)
	if found && acroFormObj != nil {
		if count := pa.processAcroForm(ctx, acroFormObj, info); count > 0 {
			signatureFound = true
		}
	}
	
	// Method 2: Search for signature objects in the PDF structure
	// Look through all objects for signature-related entries
	signatureCount := 0
	
	// Check if there are any objects with /Type/Sig
	if ctx.XRefTable != nil {
		for i := 1; i <= *ctx.XRefTable.Size; i++ {
			entry, _ := ctx.XRefTable.FindTableEntry(i, 0)
			if entry == nil || entry.Object == nil {
				continue
			}
			
			if dict, ok := entry.Object.(types.Dict); ok {
				// Check for signature type
				if typeObj, found, _ := dict.Entry("Type", "", false); found {
					if nameObj, ok := typeObj.(types.Name); ok && string(nameObj) == "Sig" {
						signatureCount++
					}
				}
				
				// Check for signature field type
				if ftObj, found, _ := dict.Entry("FT", "", false); found {
					if nameObj, ok := ftObj.(types.Name); ok && string(nameObj) == "Sig" {
						signatureCount++
					}
				}
				
				// Check for signature value (V) entry that points to signature dict
				if vObj, found, _ := dict.Entry("V", "", false); found && vObj != nil {
					if indRef, ok := vObj.(types.IndirectRef); ok {
						// Try to dereference the signature value
						if sigDict, err := ctx.Dereference(indRef); err == nil {
							if sigDictTyped, ok := sigDict.(types.Dict); ok {
								// Check if this looks like a signature dictionary
								if filterObj, found, _ := sigDictTyped.Entry("Filter", "", false); found {
									if nameObj, ok := filterObj.(types.Name); ok {
										filter := string(nameObj)
										if filter == "Adobe.PPKLite" || filter == "Adobe.PPKMS" || strings.Contains(filter, "PKCS") {
											signatureCount++
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	
	// Method 3: Simple brute force search for signature patterns
	// This is a fallback when the PDF structure is not easily accessible
	if !signatureFound && signatureCount == 0 {
		// Look for signature indicators in the document catalog
		if pa.hasSignatureIndicators(ctx) {
			signatureCount = 1 // Assume at least one signature
		}
	}
	
	if signatureCount > 0 || signatureFound {
		info.HasDigitalSignatures = true
		if signatureCount > info.SignatureCount {
			info.SignatureCount = signatureCount
		}
		return true
	}
	
	return false
}

// hasSignatureIndicators checks for general signature indicators
func (pa *PDFAnalyzer) hasSignatureIndicators(ctx *model.Context) bool {
	// Check for SigFlags in the document catalog or AcroForm
	if ctx.RootDict != nil {
		// Look for any reference to signature-related entries
		if acroFormObj, found, _ := ctx.RootDict.Entry("AcroForm", "", false); found && acroFormObj != nil {
			// Even if we can't process the AcroForm fully, check for SigFlags
			if indRef, ok := acroFormObj.(types.IndirectRef); ok {
				if resolved, err := ctx.Dereference(indRef); err == nil {
					if dict, ok := resolved.(types.Dict); ok {
						if sigFlagsObj, found, _ := dict.Entry("SigFlags", "", false); found && sigFlagsObj != nil {
							return true
						}
					}
				}
			} else if dict, ok := acroFormObj.(types.Dict); ok {
				if sigFlagsObj, found, _ := dict.Entry("SigFlags", "", false); found && sigFlagsObj != nil {
					return true
				}
			}
		}
	}
	return false
}

// processAcroForm processes the AcroForm dictionary to find signature fields
func (pa *PDFAnalyzer) processAcroForm(ctx *model.Context, acroFormObj types.Object, info *PDFInfo) int {
	// TODO: Implement AcroForm processing for signature field detection
	// This is a placeholder implementation
	return 0
}

// detectSignaturesByteAnalysis performs raw byte analysis for signature detection
func (pa *PDFAnalyzer) detectSignaturesByteAnalysis(filePath string) (bool, int, error) {
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
	
	for _, pattern := range patterns {
		count := strings.Count(content, pattern)
		if count > 0 {
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
				signatureCount = 1
				break
			}
		}
	}
	
	return signatureCount > 0, signatureCount, nil
}

// resolveFieldDict resolves a field reference to its dictionary
func (pa *PDFAnalyzer) resolveFieldDict(ctx *model.Context, fieldRef types.Object) types.Dict {
	if indRef, ok := fieldRef.(types.IndirectRef); ok {
		if resolved, err := ctx.Dereference(indRef); err == nil {
			if dict, ok := resolved.(types.Dict); ok {
				return dict
			}
		}
	} else if dict, ok := fieldRef.(types.Dict); ok {
		return dict
	}
	return nil
}

// isSignatureField checks if a field dictionary represents a signature field
func (pa *PDFAnalyzer) isSignatureField(fieldDict types.Dict) bool {
	if fieldDict == nil {
		return false
	}
	
	// Check field type
	if ftObj, found, _ := fieldDict.Entry("FT", "", false); found {
		if nameObj, ok := ftObj.(types.Name); ok && string(nameObj) == "Sig" {
			return true
		}
	}
	
	return false
}

// formatTime formats a time.Time to a standard string format
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}
