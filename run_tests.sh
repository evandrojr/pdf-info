#!/bin/bash

# PDF Analysis Program - Test Runner
# This script builds the program and runs all tests

set -e  # Exit on any error

echo "🔧 Building PDF analysis program..."
go build -o pdf-info pdf-info.go

echo "✅ Build completed successfully!"
echo ""

echo "🧪 Running integration tests..."
echo "=================================="

# Run tests with verbose output
go test -v ./...

echo ""
echo "📊 Running benchmarks..."
echo "========================"

# Run benchmarks
go test -bench=. -benchmem ./...

echo ""
echo "📈 Running test coverage analysis..."
echo "===================================="

# Run tests with coverage (note: this measures unit test coverage, not integration test coverage)
go test -cover ./...

echo ""
echo "🎉 All tests completed!"
echo ""
echo "Available test commands:"
echo "  go test -v                    # Run all tests with verbose output"
echo "  go test -run TestPDFVersion   # Run only PDF version tests"
echo "  go test -run TestInvalidInputs # Run only invalid input tests"
echo "  go test -bench=.              # Run benchmarks"
echo "  go test -cover                # Run with coverage analysis"
echo ""
echo "To run the program manually:"
echo "  ./pdf-info pdfs/teste.pdf"
