package main

import (
	"fmt"
	"os"
	"pdfparser"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: pdfparser <path_to_pdf_file>")
		return
	}

	filePath := os.Args[1]
	text, err := pdfparser.ExtractTextFromPDF(filePath)
	if err != nil {
		fmt.Printf("Error extracting text from PDF: %v\n", err)
		return
	}
	fmt.Println("Extracted text:")
	fmt.Println(text)
}
