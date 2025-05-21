package pdfparser

import (
	"os"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// ExtractTextFromPDF extracts text from all pages of a PDF file.
func ExtractTextFromPDF(filePath string) (string, error) {
	// Open the PDF file.
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Get the file size.
	fi, err := f.Stat()
	if err != nil {
		return "", err
	}
	fileSize := fi.Size()

	// Create a ReadSeeker.
	rs := f

	// Create a PDF context.
	conf := model.NewDefaultConfiguration()
	ctx, err := api.ReadContext(rs, fileSize, conf)
	if err != nil {
		return "", err
	}

	// Extract text from all pages.
	var text strings.Builder
	for _, pageNumber := range ctx.PageNumberSequence() {
		pageText, err := api.ExtractText(ctx, []string{string(pageNumber)}, conf)
		if err != nil {
			return "", err
		}
		text.WriteString(pageText)
	}

	return text.String(), nil
}
