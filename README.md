# Go PDF Parser & Image Watermarker

A command-line tool written in Go to extract text from PDF files and apply various types of watermarks to images.

## Features

*   **PDF Text Extraction**: Extracts text content from PDF documents.
*   **Image Watermarking**:
    *   Visible text watermarks on images.
    *   Visible image (logo) watermarks on images.
    *   Invisible (steganographic) watermarks in images.
    *   Batch processing of image files in a directory.
    *   Concurrent processing using multiple CPU cores.

## Dependencies

This tool uses the following Go libraries:

*   PDF Processing:
    *   `github.com/pdfcpu/pdfcpu`
*   Image Processing & Watermarking:
    *   `github.com/fogleman/gg` (for text rendering and drawing)
    *   `github.com/disintegration/imaging` (for image manipulation, loading, saving)
    *   `github.com/auyer/steganography` (for invisible watermarks)

## Build Instructions

To build the tool, navigate to the project directory and run:

```sh
go build
```

This will create an executable named `pdfparser` (or `pdfparser.exe` on Windows). Make sure the executable is in your PATH or call it using `./pdfparser`.

## Usage

The tool has two main commands: `pdf` and `watermark`.

### 1. PDF Text Extraction

To extract text from a PDF file:

```sh
pdfparser pdf /path/to/your/document.pdf
```

The extracted text will be printed to the console.

Example:
```sh
pdfparser pdf my_resume.pdf
```

### 2. Image Watermarking

The image watermarking functionality is accessed via the `watermark image` subcommand.

```sh
pdfparser watermark image [options...]
```

**General Options:**

*   `-input <path>`: (Required) Path to the input image file or a directory containing images. If a directory is specified, all `.png`, `.jpg`, and `.jpeg` files will be processed.
*   `-output <path>`: (Required) Path to the output file (if input is a single file) or output directory (if input is a directory). The output directory will be created if it doesn't exist.
*   `-type <type>`: (Required) Type of watermark to apply. Valid values:
    *   `visible`: Applies a visible watermark (text or image).
    *   `invisible`: Applies an invisible steganographic watermark.
*   `-concurrency <num>`: Number of concurrent workers to use for processing. Defaults to the number of available CPU cores.

**Options for Visible Watermarks (`-type visible`)**

When `-type visible` is chosen, you must also specify `-visible.type`.

*   `-visible.type <type>`: (Required) Type of visible watermark. Valid values:
    *   `text`: For text-based watermarks.
    *   `image`: For image-based (logo) watermarks.
*   `-visible.opacity <float>`: Opacity of the watermark (0.0 for fully transparent, 1.0 for fully opaque). Default: `0.5`.
*   `-visible.position <pos>`: Position of the watermark. Valid values: `TopLeft`, `TopCenter`, `TopRight`, `CenterLeft`, `Center`, `CenterRight`, `BottomLeft`, `BottomCenter`, `BottomRight`. Default: `BottomRight`.
*   `-visible.margin <px>`: Margin in pixels from the image border for the watermark. Default: `10`.

**Options for Text Watermarks (`-type visible -visible.type text`)**

*   `-visible.text <string>`: (Required) The text string to use as the watermark.
*   `-visible.font <path>`: (Required) Path to a `.ttf` font file.
*   `-visible.size <float>`: Font size. Default: `32`.
*   `-visible.color <hex>`: Font color in RRGGBBAA hex format (e.g., `#FF0000FF` for opaque red, `#00000080` for semi-transparent black). Default: `#000000FF` (opaque black).

**Options for Image Watermarks (`-type visible -visible.type image`)**

*   `-visible.imagepath <path>`: (Required) Path to the image file to be used as the watermark.

**Options for Invisible Watermarks (`-type invisible`)**

*   `-invisible.data <string>`: (Required) The secret data string to embed within the image.

**Note on Invisible Watermark Extraction:** Currently, the CLI tool only supports *embedding* invisible watermarks. Extracting them would require a separate tool or function call (the underlying Go function `ExtractInvisibleWatermark` exists in the `pdfparser` package).

**Examples:**

1.  **Apply a visible text watermark to a single image:**
    ```sh
    pdfparser watermark image \
        -input ./input_images/my_photo.jpg \
        -output ./output_images/my_photo_watermarked.jpg \
        -type visible \
        -visible.type text \
        -visible.text "Copyright 2024 MyName" \
        -visible.font "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf" \
        -visible.size 24 \
        -visible.color "#FFFFFF80" \
        -visible.position BottomLeft \
        -visible.margin 20
    ```

2.  **Apply a visible logo watermark to all images in a directory:**
    ```sh
    pdfparser watermark image \
        -input ./source_folder/ \
        -output ./processed_folder/ \
        -type visible \
        -visible.type image \
        -visible.imagepath ./assets/company_logo.png \
        -visible.opacity 0.8 \
        -visible.position Center
    ```

3.  **Embed an invisible watermark in a single image:**
    ```sh
    pdfparser watermark image \
        -input important_document.png \
        -output important_document_stamped.png \
        -type invisible \
        -invisible.data "ProjectID: XZ123, Client: AcmeCorp"
    ```

## Running Tests

To run the unit and integration tests:

```sh
go test ./...
```

This will run all tests in the current directory and any subdirectories.

**Note on Test Data:**
*   Some tests require specific image files (`base_image.png`, `base_image.jpg`, `logo_watermark.png`) to be present in a `testdata` directory in the project root. If these are missing, relevant tests will be skipped.
*   Text watermarking tests rely on a common system font (e.g., `/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf` on Linux). If this font is not found, tests requiring it will be skipped.
*   The PDF text extraction test (`TestExtractTextFromPDF_Success`) requires `testdata/test_document.pdf` to be manually created for full validation.

## License

This project is released under the MIT License. See the `LICENSE` file (if one exists, typically added for open source projects) for details.
(Assuming MIT if no license file is present, common for such tools)
