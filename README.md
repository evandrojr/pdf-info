# PDF Analysis Tool

A comprehensive PDF analysis tool written in Go that provides detailed information about PDF files, including metadata, technical specifications, security features, and digital signature analysis.

## Features

- **File Information**: Basic file details (size, modification date, checksums)
- **PDF Metadata**: Title, author, creation date, and other document properties
- **Technical Analysis**: PDF version, page count, encryption status, linearization
- **Security Features**: Encryption details, permission restrictions
- **Digital Signatures**: Detection and basic validation of digital signatures
- **Content Analysis**: Text extraction, image counting, page dimensions
- **Multi-language Support**: Full English output with proper error handling

## Installation

### Prerequisites

- Go 1.19 or later
- Git

### Build from Source

```bash
git clone <repository-url>
cd pdf-info
go mod tidy
go build -o pdf-info pdf-info.go
```

Or use the Makefile:

```bash
make build
```

## Usage

### Basic Usage

```bash
./pdf-info path/to/your/document.pdf
```

### Examples

```bash
# Analyze a simple PDF
./pdf-info pdfs/simple-test.pdf

# Analyze an encrypted PDF
./pdf-info pdfs/readonly.pdf

# Get help
./pdf-info
```

## Testing

This project includes comprehensive integration tests that verify all major functionality.

### Running Tests

```bash
# Run all tests
make test

# Run tests with verbose output
make test-verbose

# Run tests with coverage analysis
make test-coverage

# Run performance benchmarks
make bench

# Run specific test categories
make test-version-bug      # Test PDF version detection fix
make test-invalid-inputs   # Test error handling
make test-output-format    # Test output structure
```

### Using the Test Script

```bash
# Make the script executable (if not already)
chmod +x run_tests.sh

# Run all tests and benchmarks
./run_tests.sh
```

### Test Coverage

The test suite includes:

1. **PDF Analysis Tests**: Verify correct analysis of different PDF types
   - Simple PDFs (version 1.3)
   - Complex multi-page PDFs
   - Encrypted PDFs (version 1.6)
   - PDFs with special characters in filenames

2. **Error Handling Tests**: Test program behavior with invalid inputs
   - Missing arguments
   - Non-existent files
   - Non-PDF files (shows warnings but continues)

3. **Output Format Tests**: Verify report structure and English translation
   - All sections present
   - No Portuguese text remaining
   - Proper formatting

4. **Bug Regression Tests**: Specifically test the PDF version bug fix
   - Ensure PDF 1.3 shows as "1.3" not "3.0"
   - Ensure PDF 1.6 shows as "1.6" not "6.0"

5. **Performance Benchmarks**: Measure analysis speed

### Test PDFs

The `pdfs/` directory contains test files:

- `simple-test.pdf`: Simple PDF 1.3, single page, no encryption
- `complex-document.pdf`: PDF 1.3, multiple pages, metadata
- `readonly-signed-icp-brazil.pdf`: PDF 1.6, encrypted, digitally signed
- `readonly.pdf`: PDF 1.6, encrypted
- `pdf-version-test.pdf`: PDF 1.3, test file for version verification
- `multiple-icp-brasil-signtures.pdf`: PDF with multiple digital signatures

## Development

### Project Structure

```
pdf-info/
├── pdf-info.go              # Main source code
├── pdf_analysis_test.go     # Integration tests
├── go.mod                   # Go module definition
├── go.sum                   # Go module checksums
├── Makefile                 # Build and test automation
├── run_tests.sh             # Test execution script
├── README.md                # This file
├── cert/                    # Test certificates
├── pdfs/                    # Test PDF files
└── debug/                   # Debug utilities
```

### Key Components

- **File Analysis**: SHA256/MD5 hashing, file metadata extraction
- **PDF Processing**: Uses `pdfcpu` and `ledongthuc/pdf` libraries
- **Security Analysis**: Encryption detection, permission analysis
- **Signature Detection**: Digital signature presence and basic validation
- **Content Extraction**: Text and image analysis
- **Error Handling**: Graceful handling of corrupted or invalid files

### Dependencies

- `github.com/pdfcpu/pdfcpu`: PDF processing and manipulation
- `github.com/ledongthuc/pdf`: Alternative PDF reading library

## Bug Fixes

### PDF Version Display Fix

**Issue**: PDF versions were displayed incorrectly (e.g., "3.0" instead of "1.3")

**Cause**: The `ctx.HeaderVersion` field contains only the decimal part of the PDF version (3 for PDF 1.3, 6 for PDF 1.6), but the code was dividing by 10.

**Solution**: Changed from:
```go
version := float64(*ctx.HeaderVersion) / 10.0
info.PDFVersion = fmt.Sprintf("%.1f", version)
```

To:
```go
info.PDFVersion = fmt.Sprintf("1.%d", *ctx.HeaderVersion)
```

**Tests**: Specific regression tests verify this fix in `TestPDFVersionBugFix`.

### Complete English Translation

All program output has been translated from Portuguese to English:
- Error messages
- Report headers and sections  
- Field labels and values
- Boolean responses (Yes/No instead of Sim/Não)
- Status messages and help text

## Available Commands

### Makefile

| Command              | Description                                 |
|----------------------|---------------------------------------------|
| make build           | Build the pdf-info binary                   |
| make test            | Run all integration tests                   |
| make test-verbose    | Run tests with verbose output               |
| make test-coverage   | Run tests with coverage analysis            |
| make bench           | Run performance benchmarks                  |
| make run-example     | Run program with example PDF                |
| make clean           | Clean build artifacts                       |
| make help            | Show help message                           |
| make test-version-bug| Test PDF version bug fix                    |
| make test-invalid-inputs| Test invalid input handling              |
| make test-output-format | Test output format                      |

### Direct Go Commands

```bash
# Build
go build -o pdf-info pdf-info.go

# Run tests
go test -v                    # All tests with verbose output
go test -run TestPDFVersion   # Only PDF version tests
go test -bench=.              # Benchmarks
go test -cover                # Coverage analysis

# Run program
./pdf-info pdfs/simple-test.pdf
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass: `make test`
5. Submit a pull request

## License

[Add your license information here]

## Changelog

### Latest Version
- ✅ Fixed PDF version display bug
- ✅ Complete English translation
- ✅ Comprehensive integration test suite
- ✅ Performance benchmarks
- ✅ Improved error handling
- ✅ Enhanced documentation
