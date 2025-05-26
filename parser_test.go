package pdfparser

import (
	"os"
	"strings"
	"testing"
)

// TestExtractTextFromPDF_Success tests the successful extraction of text from a PDF.
func TestExtractTextFromPDF_Success(t *testing.T) {
	// Define the path to the test PDF file.
	// IMPORTANT: This test assumes 'testdata/test_document.pdf' exists and is a valid PDF.
	// You will need to create this file manually with known text content.
	// For example, the PDF could contain the text "Hello, this is a test PDF document."
	filePath := "testdata/test_document.pdf"
	expectedText := "Hello, this is a test PDF document."

	// Check if the test file exists. If not, skip the test.
	// This is a workaround because we cannot create a PDF file programmatically with current tools.
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Skipf("Skipping test because test PDF file '%s' does not exist. Please create it manually.", filePath)
	}

	extractedText, err := ExtractTextFromPDF(filePath)
	if err != nil {
		t.Fatalf("ExtractTextFromPDF returned an unexpected error: %v", err)
	}

	// Normalize extracted text by trimming whitespace and removing extra newlines often introduced by PDF extractors.
	normalizedExtractedText := strings.TrimSpace(strings.ReplaceAll(extractedText, "\n", " "))
	normalizedExpectedText := strings.TrimSpace(strings.ReplaceAll(expectedText, "\n", " "))

	// For now, we'll check if the expected text is contained within the extracted text.
	// This is a more robust check against minor formatting differences from pdfcpu.
	if !strings.Contains(normalizedExtractedText, normalizedExpectedText) {
		t.Errorf("Extracted text does not contain the expected string.\nExpected to contain: '%s'\nGot: '%s'", normalizedExpectedText, normalizedExtractedText)
	}
}

// TestExtractTextFromPDF_FileNotFound tests the behavior of ExtractTextFromPDF when the file is not found.
func TestExtractTextFromPDF_FileNotFound(t *testing.T) {
	filePath := "non_existent_document.pdf"
	_, err := ExtractTextFromPDF(filePath)
	if err == nil {
		t.Fatalf("ExtractTextFromPDF was expected to return an error for a non-existent file, but it returned nil")
	}
}
