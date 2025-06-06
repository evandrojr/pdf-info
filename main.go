package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . <pdf_path>")
		os.Exit(1)
	}

	pdfPath := os.Args[1]
	
	analyzer := &PDFAnalyzer{}
	info, err := analyzer.AnalyzePDF(pdfPath)
	if err != nil {
		log.Fatalf("Error analyzing PDF: %v", err)
	}

	analyzer.PrintReport(info)
}
