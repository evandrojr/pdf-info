package main

import (
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
)

// analyzeLedongthuc performs PDF analysis using the ledongthuc/pdf library
func (pa *PDFAnalyzer) analyzeLedongthuc(filePath string, info *PDFInfo) error {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	totalTextLength := 0
	var fontsUsed []string
	imagesCount := 0

	// Extrair texto de todas as páginas
	for i := 1; i <= r.NumPage(); i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}
		
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		
		textLen := len(strings.TrimSpace(text))
		totalTextLength += textLen
		
		// Atualizar informação da página se ela existir
		if i-1 < len(info.Pages) {
			info.Pages[i-1].TextLength = textLen
		}
	}	
	info.TotalTextLength = totalTextLength
	info.FontsUsed = fontsUsed
	info.ImagesCount = imagesCount
	
	return nil
}

// Helper function to read all content from a ReadCloser
func readAll(rc io.ReadCloser) ([]byte, error) {
	defer rc.Close()
	return io.ReadAll(rc)
}
