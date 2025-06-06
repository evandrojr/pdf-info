# PDF Info Tool

This project provides tools and utilities for extracting and analyzing information from PDF files, with a focus on digital signatures and certificate validation. It is designed for use in environments where digital document integrity and authenticity are critical, such as legal, governmental, and enterprise settings.

## Features

- Extract metadata and information from PDF files
- Analyze digital signatures and certificate chains
- Support for ICP-Brasil signed documents
- Command-line interface for easy integration
- Example certificates and test PDFs included

## Project Structure

```
app.go                  # Main application source code
debug_version.go        # Debugging and versioning utilities
go.mod, go.sum          # Go module dependencies
cert/                   # Example certificates for testing
pdfs/                   # Sample PDF files for analysis
teste.txt               # Test file
```

## Getting Started

### Prerequisites

- [Go](https://golang.org/dl/) 1.18 or higher

### Installation

Clone the repository:

```bash
git clone <repository-url>
cd pdf-info
```

Install dependencies:

```bash
go mod tidy
```

### Usage

To run the main application:

```bash
go run app.go
```

You can place your PDF files in the `pdfs/` directory and use the tool to extract information or validate signatures.

### Example

```bash
go run app.go pdfs/teste.pdf
```

## Directory Details

- `cert/`: Contains test certificates (PEM format) for signature validation.
- `pdfs/`: Contains various sample PDF files, including signed and unsigned documents.

## Contributing

Contributions are welcome! Please open issues or submit pull requests for improvements, bug fixes, or new features.

## License

This project is licensed under the MIT License.

## Author

Developed by [Your Name].
