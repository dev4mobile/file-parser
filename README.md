# Go PDF Parser

A simple command-line tool written in Go to extract text from PDF files.

## Dependencies

This tool uses the `pdfcpu` library: [https://github.com/pdfcpu/pdfcpu](https://github.com/pdfcpu/pdfcpu)

## Build Instructions

To build the parser, navigate to the project directory and run:

```sh
go build
```

This will create an executable named `pdfparser` (or `pdfparser.exe` on Windows).

## Usage

To extract text from a PDF file, run the executable with the path to your PDF file as an argument:

```sh
./pdfparser /path/to/your/document.pdf
```

The extracted text will be printed to the console.

Example:
```sh
./pdfparser my_resume.pdf
```

## Running Tests

To run the unit tests:

```sh
go test
```

**Note:** For the `TestExtractTextFromPDF_Success` test to run completely, you need to manually create a file named `test_document.pdf` inside a `testdata` directory in the project root. This file should contain some simple text. For example: "Hello, this is a test PDF document."
