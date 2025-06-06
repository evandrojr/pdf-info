package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestPDFAnalysis executes integration tests for the PDF analysis program
func TestPDFAnalysis(t *testing.T) {
	// Ensure the binary exists
	binaryPath := "./pdf-info"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Fatal("Binary pdf-info not found. Please run 'go build -o pdf-info pdf-info.go' first")
	}

	// Test cases with expected results
	testCases := []struct {
		name           string
		pdfFile        string
		expectedInOutput []string
		notExpectedInOutput []string
	}{
		{
			name:    "Simple PDF 1.3",
			pdfFile: "pdfs/simple-test.pdf",
			expectedInOutput: []string{
				"PDF ANALYSIS REPORT",
				"FILE INFORMATION",
				"PDF version: 1.3",
				"File name: simple-test.pdf",
				"Size:",
				"Is encrypted: No",
			},
			notExpectedInOutput: []string{
				"Error analyzing PDF",
				"PDF version: 3.0", // This was the bug we fixed
				"Is encrypted: Yes",
			},
		},
		{
			name:    "Complex PDF with multiple pages",
			pdfFile: "pdfs/complex-document.pdf",
			expectedInOutput: []string{
				"PDF ANALYSIS REPORT",
				"FILE INFORMATION",
				"PDF version: 1.3",
				"File name: complex-document.pdf",
				"Number of pages: 2",
				"Is encrypted: No",
			},
			notExpectedInOutput: []string{
				"Error analyzing PDF",
				"Number of pages: 1",
				"Is encrypted: Yes",
			},
		},
		{
			name:    "Encrypted PDF",
			pdfFile: "pdfs/readonly-signed-icp-brazil.pdf",
			expectedInOutput: []string{
				"PDF ANALYSIS REPORT",
				"FILE INFORMATION",
				"PDF version: 1.6",
				"File name: readonly-signed-icp-brazil.pdf",
				"Is encrypted: Yes",
				"SECURITY INFORMATION",
			},
			notExpectedInOutput: []string{
				"Error analyzing PDF",
				"PDF version: 1.3",
				"Is encrypted: No",
			},
		},
		{
			name:    "Another encrypted PDF",
			pdfFile: "pdfs/readonly.pdf",
			expectedInOutput: []string{
				"PDF ANALYSIS REPORT",
				"FILE INFORMATION",
				"PDF version: 1.6",
				"File name: readonly.pdf",
				"Is encrypted: Yes",
				"SECURITY INFORMATION",
			},
			notExpectedInOutput: []string{
				"Error analyzing PDF",
				"PDF version: 1.3",
				"Is encrypted: No",
			},
		},
		{
			name:    "PDF with special characters in name",
			pdfFile: "pdfs/pdf-version-test.pdf",
			expectedInOutput: []string{
				"PDF ANALYSIS REPORT",
				"FILE INFORMATION",
				"PDF version: 1.3",
				"File name: pdf-version-test.pdf",
				"Is encrypted: No",
			},
			notExpectedInOutput: []string{
				"Error analyzing PDF",
				"PDF version: 3.0",
				"Is encrypted: Yes",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Check if the PDF file exists
			if _, err := os.Stat(tc.pdfFile); os.IsNotExist(err) {
				t.Skipf("PDF file %s not found, skipping test", tc.pdfFile)
			}

			// Execute the binary with the PDF file
			cmd := exec.Command(binaryPath, tc.pdfFile)
			output, err := cmd.CombinedOutput()
			
			if err != nil {
				t.Fatalf("Failed to execute binary: %v\nOutput: %s", err, string(output))
			}

			outputStr := string(output)

			// Check expected strings are present
			for _, expected := range tc.expectedInOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain '%s', but it didn't.\nFull output:\n%s", expected, outputStr)
				}
			}

			// Check unwanted strings are not present
			for _, notExpected := range tc.notExpectedInOutput {
				if strings.Contains(outputStr, notExpected) {
					t.Errorf("Expected output NOT to contain '%s', but it did.\nFull output:\n%s", notExpected, outputStr)
				}
			}
		})
	}
}

// TestInvalidInputs tests the program's behavior with invalid inputs
func TestInvalidInputs(t *testing.T) {
	binaryPath := "./pdf-info"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Fatal("Binary pdf-info not found. Please run 'go build -o pdf-info pdf-info.go' first")
	}

	testCases := []struct {
		name     string
		args     []string
		expectError bool
	}{
		{
			name:     "No arguments",
			args:     []string{},
			expectError: true,
		},
		{
			name:     "Non-existent file",
			args:     []string{"non_existent_file.pdf"},
			expectError: true,
		},
		{
			name:     "Non-PDF file",
			args:     []string{"pdf-info.go"},
			expectError: false, // Program doesn't fail, but shows warnings
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tc.args...)
			output, err := cmd.CombinedOutput()
			
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected command to fail, but it succeeded.\nOutput: %s", string(output))
				}
				// Check that usage message is shown for no arguments
				if len(tc.args) == 0 && !strings.Contains(string(output), "Usage:") {
					t.Errorf("Expected usage message when no arguments provided.\nOutput: %s", string(output))
				}
			} else {
				if err != nil {
					t.Errorf("Expected command to succeed, but it failed: %v\nOutput: %s", err, string(output))
				}
			}
		})
	}
}

// TestOutputFormat tests the overall structure and format of the output
func TestOutputFormat(t *testing.T) {
	binaryPath := "./pdf-info"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Fatal("Binary pdf-info not found. Please run 'go build -o pdf-info pdf-info.go' first")
	}

	// Use a simple PDF for this test
	pdfFile := "pdfs/simple-test.pdf"
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("PDF file pdfs/simple-test.pdf not found, skipping test")
	}

	cmd := exec.Command(binaryPath, pdfFile)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		t.Fatalf("Failed to execute binary: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)

	// Test the overall structure of the output
	expectedSections := []string{
		"PDF ANALYSIS REPORT",
		"FILE INFORMATION",
		"DOCUMENT METADATA",
		"TECHNICAL INFORMATION",
		"CONTENT INFORMATION",
		"DIGITAL SIGNATURES",
	}

	for _, section := range expectedSections {
		if !strings.Contains(outputStr, section) {
			t.Errorf("Expected output to contain section '%s', but it didn't.\nFull output:\n%s", section, outputStr)
		}
	}

	// Test that all output is in English (no Portuguese words)
	portugueseWords := []string{
		"Erro",
		"Versão",
		"Número",
		"páginas",
		"Sim",
		"Não",
		"Informações",
		"Arquivo",
		"Documento",
		"Técnicas",
		"Assinaturas",
		"Digitais",
		"Segurança",
		"Características",
	}

	for _, word := range portugueseWords {
		if strings.Contains(outputStr, word) {
			t.Errorf("Found Portuguese word '%s' in output. All text should be in English.\nFull output:\n%s", word, outputStr)
		}
	}
}

// TestPDFVersionBugFix specifically tests that the PDF version bug is fixed
func TestPDFVersionBugFix(t *testing.T) {
	binaryPath := "./pdf-info"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Fatal("Binary pdf-info not found. Please run 'go build -o pdf-info pdf-info.go' first")
	}

	// Test multiple PDFs to ensure version detection is correct
	testCases := []struct {
		pdfFile         string
		expectedVersion string
		buggyVersion    string
	}{
		{
			pdfFile:         "pdfs/simple-test.pdf",
			expectedVersion: "PDF version: 1.3",
			buggyVersion:    "PDF version: 3.0",
		},
		{
			pdfFile:         "pdfs/complex-document.pdf",
			expectedVersion: "PDF version: 1.3",
			buggyVersion:    "PDF version: 3.0",
		},
		{
			pdfFile:         "pdfs/readonly-signed-icp-brazil.pdf",
			expectedVersion: "PDF version: 1.6",
			buggyVersion:    "PDF version: 6.0",
		},
	}

	for _, tc := range testCases {
		t.Run(filepath.Base(tc.pdfFile), func(t *testing.T) {
			if _, err := os.Stat(tc.pdfFile); os.IsNotExist(err) {
				t.Skipf("PDF file %s not found, skipping test", tc.pdfFile)
			}

			cmd := exec.Command(binaryPath, tc.pdfFile)
			output, err := cmd.CombinedOutput()
			
			if err != nil {
				t.Fatalf("Failed to execute binary: %v\nOutput: %s", err, string(output))
			}

			outputStr := string(output)

			// Check that the correct version is shown
			if !strings.Contains(outputStr, tc.expectedVersion) {
				t.Errorf("Expected to find '%s' in output, but didn't.\nFull output:\n%s", tc.expectedVersion, outputStr)
			}

			// Check that the buggy version is NOT shown
			if strings.Contains(outputStr, tc.buggyVersion) {
				t.Errorf("Found buggy version '%s' in output. The PDF version bug should be fixed.\nFull output:\n%s", tc.buggyVersion, outputStr)
			}
		})
	}
}

// TestNonPDFFileHandling tests how the program handles non-PDF files
func TestNonPDFFileHandling(t *testing.T) {
	binaryPath := "./pdf-info"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Fatal("Binary pdf-info not found. Please run 'go build -o pdf-info pdf-info.go' first")
	}

	// Test with a non-PDF file
	cmd := exec.Command(binaryPath, "pdf-info.go")
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		t.Fatalf("Failed to execute binary: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)

	// Check that appropriate warnings are shown
	expectedWarnings := []string{
		"Warning: error in pdfcpu analysis",
		"Warning: error in ledongthuc analysis",
		"not a PDF file",
	}

	for _, warning := range expectedWarnings {
		if !strings.Contains(outputStr, warning) {
			t.Errorf("Expected warning '%s' in output for non-PDF file, but didn't find it.\nFull output:\n%s", warning, outputStr)
		}
	}

	// Should still generate a report, but with empty/default values
	if !strings.Contains(outputStr, "PDF ANALYSIS REPORT") {
		t.Error("Expected PDF ANALYSIS REPORT header even for non-PDF files")
	}

	// PDF version should be empty for non-PDF files
	if !strings.Contains(outputStr, "PDF version: \n") && !strings.Contains(outputStr, "PDF version: ") {
		t.Error("Expected empty PDF version for non-PDF files")
	}
}

// BenchmarkPDFAnalysis benchmarks the performance of PDF analysis
func BenchmarkPDFAnalysis(b *testing.B) {
	binaryPath := "./pdf-info"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		b.Fatal("Binary pdf-info not found. Please run 'go build -o pdf-info pdf-info.go' first")
	}

	pdfFile := "pdfs/simple-test.pdf"
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		b.Skip("PDF file pdfs/simple-test.pdf not found, skipping benchmark")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := exec.Command(binaryPath, pdfFile)
		_, err := cmd.CombinedOutput()
		if err != nil {
			b.Fatalf("Failed to execute binary: %v", err)
		}
	}
}
