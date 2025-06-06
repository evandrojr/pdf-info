# PDF Analysis Program Makefile

.PHONY: build test test-verbose test-coverage bench clean help run-example test-version-bug test-invalid-inputs test-output-format

# Default target
help:
	@echo "PDF Analysis Program - Available commands:"
	@echo ""
	@echo "  make build          Build the pdf-info binary"
	@echo "  make test           Run all integration tests"
	@echo "  make test-verbose   Run tests with verbose output"
	@echo "  make test-coverage  Run tests with coverage analysis"
	@echo "  make bench          Run performance benchmarks"
	@echo "  make run-example    Run program with example PDF"
	@echo "  make clean          Clean build artifacts"
	@echo "  make help           Show this help message"

# Build the program
build:
	@echo "🔧 Building PDF analysis program with static linking..."
	CGO_ENABLED=0 go build -ldflags="-s -w -extldflags '-static'" -a -installsuffix cgo -o pdf-info pdf-info.go
	@echo "✅ Build completed successfully!"

# Run all tests
test: build
	@echo "🧪 Running integration tests..."
	go test ./...

# Run tests with verbose output
test-verbose: build
	@echo "🧪 Running integration tests (verbose)..."
	go test -v ./...

# Run tests with coverage
test-coverage: build
	@echo "📊 Running tests with coverage analysis..."
	go test -cover ./...

# Run benchmarks
bench: build
	@echo "📈 Running performance benchmarks..."
	go test -bench=. -benchmem ./...

# Run program with example PDF
run-example: build
	@echo "🔍 Running PDF analysis on example file..."
	@if [ -f "pdfs/teste.pdf" ]; then \
		./pdf-info pdfs/teste.pdf; \
	else \
		echo "❌ Example PDF (pdfs/teste.pdf) not found"; \
	fi

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f pdf-info
	go clean
	@echo "✅ Cleanup completed!"

# Test specific functionality
test-version-bug: build
	@echo "🐛 Testing PDF version bug fix..."
	go test -v -run TestPDFVersionBugFix ./...

test-invalid-inputs: build
	@echo "❌ Testing invalid input handling..."
	go test -v -run TestInvalidInputs ./...

test-output-format: build
	@echo "📝 Testing output format..."
	go test -v -run TestOutputFormat ./...
