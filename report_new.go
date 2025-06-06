package main

import (
	"fmt"
	"strings"
	"time"
)

// PrintReport generates and prints a comprehensive PDF analysis report
func (pa *PDFAnalyzer) PrintReport(info *PDFInfo) {
	fmt.Println("=" + strings.Repeat("=", 80))
	fmt.Println("                        PDF ANALYSIS REPORT")
	fmt.Println("=" + strings.Repeat("=", 80))

	// File information
	pa.printFileInformation(info)

	// Document information
	pa.printDocumentMetadata(info)

	// Technical information
	pa.printTechnicalInformation(info)

	// Security information
	if info.IsEncrypted {
		pa.printSecurityInformation(info)
	}

	// Content information
	pa.printContentInformation(info)

	// Page information
	if len(info.Pages) > 0 {
		pa.printPageInformation(info)
	}

	// Bookmarks
	if len(info.Bookmarks) > 0 {
		pa.printBookmarks(info)
	}

	// Attachments
	if len(info.Attachments) > 0 {
		pa.printAttachments(info)
	}

	// Digital signatures - always visible section
	pa.printDigitalSignatures(info)

	// Footer
	pa.printReportFooter()
}

// printFileInformation prints basic file information
func (pa *PDFAnalyzer) printFileInformation(info *PDFInfo) {
	fmt.Println("\n📁 FILE INFORMATION")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("File name: %s\n", info.FileName)
	fmt.Printf("Path: %s\n", info.FilePath)
	fmt.Printf("Size: %s (%d bytes)\n", info.FileSizeHuman, info.FileSize)
	fmt.Printf("Last modified: %s\n", info.LastModified.Format("2006-01-02 15:04:05"))
	fmt.Printf("MD5: %s\n", info.MD5Hash)
	fmt.Printf("SHA256: %s\n", info.SHA256Hash)
}

// printDocumentMetadata prints PDF document metadata
func (pa *PDFAnalyzer) printDocumentMetadata(info *PDFInfo) {
	fmt.Println("\n📄 DOCUMENT METADATA")
	fmt.Println(strings.Repeat("-", 50))
	printIfNotEmpty("Title", info.Title)
	printIfNotEmpty("Author", info.Author)
	printIfNotEmpty("Subject", info.Subject)
	printIfNotEmpty("Keywords", info.Keywords)
	printIfNotEmpty("Creator", info.Creator)
	printIfNotEmpty("Producer", info.Producer)
	printIfNotEmpty("Creation date", info.CreationDate)
	printIfNotEmpty("Modification date", info.ModDate)
}

// printTechnicalInformation prints technical PDF information
func (pa *PDFAnalyzer) printTechnicalInformation(info *PDFInfo) {
	fmt.Println("\n⚙️  TECHNICAL INFORMATION")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("PDF version: %s\n", info.PDFVersion)
	fmt.Printf("Number of pages: %d\n", info.PageCount)
	fmt.Printf("Is encrypted: %s\n", boolToYesNo(info.IsEncrypted))
	fmt.Printf("Is linearized: %s\n", boolToYesNo(info.IsLinearized))
	fmt.Printf("Is tagged (accessible): %s\n", boolToYesNo(info.IsTagged))
	fmt.Printf("Has bookmarks: %s\n", boolToYesNo(info.HasBookmarks))
	fmt.Printf("Has attachments: %s\n", boolToYesNo(info.HasAttachments))
	fmt.Printf("Has forms: %s\n", boolToYesNo(info.HasForms))
	fmt.Printf("Has JavaScript: %s\n", boolToYesNo(info.HasJavaScript))
	fmt.Printf("Has annotations: %s\n", boolToYesNo(info.HasAnnotations))
	fmt.Printf("Has digital signatures: %s\n", boolToYesNo(info.HasDigitalSignatures))
	if info.HasDigitalSignatures {
		fmt.Printf("Number of signatures: %d\n", info.SignatureCount)
	}
}

// printSecurityInformation prints security and permissions information
func (pa *PDFAnalyzer) printSecurityInformation(info *PDFInfo) {
	fmt.Println("\n🔒 SECURITY INFORMATION")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("User password set: %s\n", boolToYesNo(info.UserPasswordSet))
	fmt.Printf("Owner password set: %s\n", boolToYesNo(info.OwnerPasswordSet))
	fmt.Printf("Printing allowed: %s\n", boolToYesNo(info.PrintAllowed))
	fmt.Printf("Modification allowed: %s\n", boolToYesNo(info.ModifyAllowed))
	fmt.Printf("Copy allowed: %s\n", boolToYesNo(info.CopyAllowed))
	fmt.Printf("Add notes allowed: %s\n", boolToYesNo(info.AddNotesAllowed))
	fmt.Printf("Fill forms allowed: %s\n", boolToYesNo(info.FillFormsAllowed))
	fmt.Printf("Accessibility access: %s\n", boolToYesNo(info.AccessibilityAllowed))
	fmt.Printf("Document assembly allowed: %s\n", boolToYesNo(info.AssembleAllowed))
	fmt.Printf("High quality printing: %s\n", boolToYesNo(info.PrintHighQualityAllowed))
}

// printContentInformation prints content analysis information
func (pa *PDFAnalyzer) printContentInformation(info *PDFInfo) {
	fmt.Println("\n📝 CONTENT INFORMATION")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("Total text characters: %d\n", info.TotalTextLength)
	fmt.Printf("Number of images: %d\n", info.ImagesCount)
	if len(info.FontsUsed) > 0 {
		fmt.Printf("Fonts used: %s\n", strings.Join(info.FontsUsed, ", "))
	}
}

// printPageInformation prints information about PDF pages
func (pa *PDFAnalyzer) printPageInformation(info *PDFInfo) {
	fmt.Println("\n📖 PAGE INFORMATION")
	fmt.Println(strings.Repeat("-", 50))
	for i, page := range info.Pages {
		if i < 5 { // Show only the first 5 pages
			fmt.Printf("Page %d: %.1f x %.1f pts, rotation: %d°, text: %d chars\n",
				page.Number, page.Width, page.Height, page.Rotation, page.TextLength)
		}
	}
	if len(info.Pages) > 5 {
		fmt.Printf("... and %d more pages\n", len(info.Pages)-5)
	}
}

// printBookmarks prints bookmark information
func (pa *PDFAnalyzer) printBookmarks(info *PDFInfo) {
	fmt.Println("\n🔖 BOOKMARKS")
	fmt.Println(strings.Repeat("-", 50))
	for _, bookmark := range info.Bookmarks {
		indent := strings.Repeat("  ", bookmark.Level-1)
		fmt.Printf("%s- %s (page %d)\n", indent, bookmark.Title, bookmark.Page)
	}
}

// printAttachments prints attachment information
func (pa *PDFAnalyzer) printAttachments(info *PDFInfo) {
	fmt.Println("\n📎 ATTACHMENTS")
	fmt.Println(strings.Repeat("-", 50))
	for _, attachment := range info.Attachments {
		fmt.Printf("- %s (%s, %s)\n", attachment.Name, attachment.Type, formatFileSize(attachment.Size))
	}
}

// printDigitalSignatures prints digital signature information
func (pa *PDFAnalyzer) printDigitalSignatures(info *PDFInfo) {
	fmt.Println("\n🔐 DIGITAL SIGNATURES")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("Document has signatures: %s\n", boolToYesNo(info.HasDigitalSignatures))
	fmt.Printf("Number of signatures: %d\n", info.SignatureCount)

	if info.HasDigitalSignatures && len(info.Signatures) > 0 {
		fmt.Println("\nSignature details:")
		for i, sig := range info.Signatures {
			fmt.Printf("\n  Signature %d:\n", i+1)
			if sig.FieldName != "" {
				fmt.Printf("    Field: %s\n", sig.FieldName)
			}
			fmt.Printf("    Type: %s\n", sig.Type)
			if sig.SubFilter != "" {
				fmt.Printf("    SubFilter: %s\n", sig.SubFilter)
			}
			fmt.Printf("    Status: %s\n", sig.Status)
			fmt.Printf("    Valid: %s\n", boolToYesNo(sig.IsValid))
			fmt.Printf("    Certified: %s\n", boolToYesNo(sig.IsCertified))
			if sig.SignerName != "" {
				fmt.Printf("    Signer: %s\n", sig.SignerName)
			}
			if sig.SigningTime != "" {
				fmt.Printf("    Signing date/time: %s\n", sig.SigningTime)
			}
			if sig.Location != "" {
				fmt.Printf("    Location: %s\n", sig.Location)
			}
			if sig.Reason != "" {
				fmt.Printf("    Reason: %s\n", sig.Reason)
			}
			if sig.ContactInfo != "" {
				fmt.Printf("    Contact: %s\n", sig.ContactInfo)
			}
			
			// Timestamp information
			fmt.Printf("    Has timestamp: %s\n", boolToYesNo(sig.HasTimestamp))
			if sig.HasTimestamp {
				if sig.TimestampType != "" {
					fmt.Printf("    Timestamp type: %s\n", sig.TimestampType)
				}
				if sig.TimestampTime != "" {
					fmt.Printf("    Timestamp time: %s\n", sig.TimestampTime)
				}
				if sig.TimestampAuthority != "" {
					fmt.Printf("    Timestamp authority: %s\n", sig.TimestampAuthority)
				}
				if sig.TimestampStatus != "" {
					fmt.Printf("    Timestamp status: %s\n", sig.TimestampStatus)
				}
			}
			
			if len(sig.ValidationErrors) > 0 {
				fmt.Printf("    Validation issues:\n")
				for _, err := range sig.ValidationErrors {
					fmt.Printf("      - %s\n", err)
				}
			}
		}
	} else if info.HasDigitalSignatures {
		fmt.Println("\nDigital signature(s) detected in document.")
		fmt.Printf("Found %d signature(s), but detailed validation failed due to encryption or security restrictions.\n", info.SignatureCount)
		fmt.Println("\nSignature validation requires:")
		fmt.Println("  • Document decryption (if encrypted)")
		fmt.Println("  • Access to signing certificates")
		fmt.Println("  • Valid certificate chain")
		fmt.Println("  • Trusted certificate authority (CA)")
	} else {
		fmt.Println("\nThis document does not have digital signatures.")
		fmt.Println("To digitally sign a PDF, you can use:")
		fmt.Println("  • Adobe Acrobat")
		fmt.Println("  • LibreOffice")
		fmt.Println("  • Online signature tools")
		fmt.Println("  • ICP-Brasil digital certificates")
	}
}

// printReportFooter prints the report footer
func (pa *PDFAnalyzer) printReportFooter() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("Analysis completed at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("=", 80))
}
