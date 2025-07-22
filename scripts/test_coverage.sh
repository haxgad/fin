#!/bin/bash

# Test Coverage Analysis Script for Internal Transfers System

set -e

echo "🧪 Running Test Coverage Analysis"
echo "================================="

# Clean previous coverage files
rm -f coverage.out coverage.html

echo ""
echo "📊 Running tests with coverage..."
go test -v -coverprofile=coverage.out ./...

echo ""
echo "📈 Overall Coverage Summary:"
echo "==========================="
go tool cover -func=coverage.out | grep "total:"

echo ""
echo "📋 Function-by-Function Coverage:"
echo "================================="
go tool cover -func=coverage.out

echo ""
echo "🌐 Generating HTML coverage report..."
go tool cover -html=coverage.out -o coverage.html

echo ""
echo "✅ Coverage analysis complete!"
echo ""
echo "📁 Files generated:"
echo "  - coverage.out  (raw coverage data)"
echo "  - coverage.html (visual HTML report)"
echo ""
echo "🌐 Open coverage.html in your browser to see detailed coverage visualization"
echo ""

# Display coverage percentage for each package
echo "📦 Package Coverage Summary:"
echo "============================"
go test -cover ./... 2>/dev/null | grep -E "coverage:|ok.*coverage" | sed 's/.*\t//' | sort
