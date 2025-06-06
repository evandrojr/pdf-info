package main

import (
	"fmt"
)

// AnalyzePDF performs comprehensive analysis of a PDF file
func (pa *PDFAnalyzer) AnalyzePDF(filePath string) (*PDFInfo, error) {
	info := &PDFInfo{}
	
	// Basic file information
	if err := pa.getFileInfo(filePath, info); err != nil {
		return nil, fmt.Errorf("error getting file information: %v", err)
	}

	// Analysis using pdfcpu
	if err := pa.analyzePDFCPU(filePath, info); err != nil {
		fmt.Printf("Warning: error in pdfcpu analysis: %v\n", err)
	}

	// Analysis using ledongthuc/pdf
	if err := pa.analyzeLedongthuc(filePath, info); err != nil {
		fmt.Printf("Warning: error in ledongthuc analysis: %v\n", err)
	}

	return info, nil
}
