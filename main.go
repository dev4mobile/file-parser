package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"os"
	"path/filepath"
	"pdfparser" // Assuming your package pdfparser contains ImageJob, ProcessImagesConcurrently, etc.
	"runtime"
	"strings"

	"github.com/disintegration/imaging" // For potentially creating dummy images if needed by other parts, or if used directly.
)

func main() {
	if len(os.Args) < 2 {
		printGeneralUsage()
		return
	}

	switch os.Args[1] {
	case "pdf":
		handlePdfCommand()
	case "watermark":
		if len(os.Args) < 3 {
			fmt.Println("Expected 'image' subcommand after 'watermark'")
			printGeneralUsage()
			return
		}
		if os.Args[2] == "image" {
			handleWatermarkImageCommand()
		} else {
			fmt.Printf("Unknown subcommand for 'watermark': %s\n", os.Args[2])
			printGeneralUsage()
		}
	case "-h", "--help":
		printGeneralUsage()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printGeneralUsage()
	}
}

func printGeneralUsage() {
	fmt.Println("\nGo PDF Parser & Watermarker Tool")
	fmt.Println("---------------------------------")
	fmt.Println("Usage: pdfparser <command> [subcommand] [options]")
	fmt.Println("\nAvailable Commands:")
	fmt.Println("  pdf <path_to_pdf_file.pdf>")
	fmt.Println("    Extracts text from the specified PDF file.")
	fmt.Println("\n  watermark image -input <path> -output <path> -type <type> [options...]")
	fmt.Println("    Applies watermarks to images. Use 'watermark image -h' for detailed options.")
	fmt.Println("\n  -h, --help")
	fmt.Println("    Show this general usage message.")
}

func handlePdfCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: pdfparser pdf <path_to_pdf_file.pdf>")
		return
	}
	pdfFilePath := os.Args[2]
	if !strings.HasSuffix(strings.ToLower(pdfFilePath), ".pdf") {
		fmt.Println("Error: Please provide a valid .pdf file.")
		return
	}

	fmt.Printf("\nAttempting to extract text from PDF: %s\n", pdfFilePath)
	extractedText, err := pdfparser.ExtractTextFromPDF(pdfFilePath)
	if err != nil {
		fmt.Printf("Error extracting text from PDF '%s': %v\n", pdfFilePath, err)
	} else {
		fmt.Println("\nExtracted text:")
		fmt.Println(extractedText)
	}
}

func handleWatermarkImageCommand() {
	imageWatermarkCmd := flag.NewFlagSet("image", flag.ExitOnError)

	// Define flags
	input := imageWatermarkCmd.String("input", "", "Path to input image file or directory (required)")
	output := imageWatermarkCmd.String("output", "", "Path to output file or directory (required)")
	typeFlag := imageWatermarkCmd.String("type", "", "Type of watermark: 'visible' or 'invisible' (required)")
	concurrency := imageWatermarkCmd.Int("concurrency", runtime.NumCPU(), "Number of concurrent workers")

	// Visible Watermark Flags
	visibleType := imageWatermarkCmd.String("visible.type", "", "Type of visible watermark: 'text' or 'image'")
	visibleOpacity := imageWatermarkCmd.Float64("visible.opacity", 0.5, "Opacity (0.0-1.0)")
	visiblePosition := imageWatermarkCmd.String("visible.position", "BottomRight", "Position (e.g., TopLeft, Center)")
	visibleMargin := imageWatermarkCmd.Int("visible.margin", 10, "Margin in pixels")
	visibleText := imageWatermarkCmd.String("visible.text", "", "Watermark text (if -visible.type=text)")
	visibleFont := imageWatermarkCmd.String("visible.font", "", "Path to font file (if -visible.type=text)")
	visibleFontSize := imageWatermarkCmd.Float64("visible.size", 32, "Font size (if -visible.type=text)")
	visibleColor := imageWatermarkCmd.String("visible.color", "#000000FF", "Font color hex (RRGGBBAA if -visible.type=text)")
	visibleImagePath := imageWatermarkCmd.String("visible.imagepath", "", "Path to watermark image (if -visible.type=image)")

	// Invisible Watermark Flags
	invisibleData := imageWatermarkCmd.String("invisible.data", "", "Data to embed (if -type=invisible)")

	// Stegano Watermark Flags
	steganoData := imageWatermarkCmd.String("stegano.data", "", "Data to embed (if -type=stegano)")
	steganoPassword := imageWatermarkCmd.String("stegano.password", "", "Optional password for stegano watermark")

	// Custom usage message for the subcommand
	imageWatermarkCmd.Usage = func() {
		fmt.Fprintf(imageWatermarkCmd.Output(), "Usage: pdfparser watermark image -input <path> -output <path> -type <type> [options...]\n\n")
		fmt.Fprintln(imageWatermarkCmd.Output(), "Required flags:")
		fmt.Fprintln(imageWatermarkCmd.Output(), "  -input <path>      Input image file or directory.")
		fmt.Fprintln(imageWatermarkCmd.Output(), "  -output <path>     Output file or directory.")
		fmt.Fprintln(imageWatermarkCmd.Output(), "  -type <type>       Watermark type: 'visible', 'invisible', or 'stegano'.")
		fmt.Fprintln(imageWatermarkCmd.Output(), "\nOptions for all watermarks:")
		fmt.Fprintln(imageWatermarkCmd.Output(), "  -concurrency <num> Number of workers (default: number of CPUs).")
		fmt.Fprintln(imageWatermarkCmd.Output(), "\nOptions for -type=visible:")
		fmt.Fprintln(imageWatermarkCmd.Output(), "  -visible.type <type>      Type of visible watermark: 'text' or 'image' (required if -type=visible).")
		fmt.Fprintln(imageWatermarkCmd.Output(), "  -visible.opacity <float>  Opacity (0.0-1.0, default 0.5).")
		fmt.Fprintln(imageWatermarkCmd.Output(), "  -visible.position <pos>   Position (default BottomRight).")
		fmt.Fprintln(imageWatermarkCmd.Output(), "  -visible.margin <px>      Margin in pixels (default 10).")
		fmt.Fprintln(imageWatermarkCmd.Output(), "  For -visible.type=text:")
		fmt.Fprintln(imageWatermarkCmd.Output(), "    -visible.text <string>    Watermark text (required).")
		fmt.Fprintln(imageWatermarkCmd.Output(), "    -visible.font <path>      Path to font file (required).")
		fmt.Fprintln(imageWatermarkCmd.Output(), "    -visible.size <float>     Font size (default 32).")
		fmt.Fprintln(imageWatermarkCmd.Output(), "    -visible.color <hex>      Font color RRGGBBAA (default #000000FF).")
		fmt.Fprintln(imageWatermarkCmd.Output(), "  For -visible.type=image:")
		fmt.Fprintln(imageWatermarkCmd.Output(), "    -visible.imagepath <path> Path to watermark image (required).")
		fmt.Fprintln(imageWatermarkCmd.Output(), "\nOptions for -type=invisible:")
		fmt.Fprintln(imageWatermarkCmd.Output(), "  -invisible.data <string>  Data to embed (required if -type=invisible).")
		fmt.Fprintln(imageWatermarkCmd.Output(), "\nOptions for -type=stegano:")
		fmt.Fprintln(imageWatermarkCmd.Output(), "  Embeds data into the image using LSB (Least Significant Bit) steganography.")
		fmt.Fprintln(imageWatermarkCmd.Output(), "  -stegano.data <string>    The string data to embed into the image. (Required)")
		fmt.Fprintln(imageWatermarkCmd.Output(), "  -stegano.password <string> Optional password. If provided, the data will be encrypted before embedding.")
		fmt.Fprintln(imageWatermarkCmd.Output(), "                           The same password is required for extraction.")
		fmt.Fprintln(imageWatermarkCmd.Output(), "  Recommendation: Use PNG format for the output image (-output <name>.png).")
		fmt.Fprintln(imageWatermarkCmd.Output(), "                  LSB steganography is sensitive to lossy compression (like JPEG).")
		fmt.Fprintln(imageWatermarkCmd.Output(), "  Robustness Note: This method includes error correction (Reed-Solomon codes).")
		fmt.Fprintln(imageWatermarkCmd.Output(), "                   However, LSB-based steganography is generally NOT resilient to")
		fmt.Fprintln(imageWatermarkCmd.Output(), "                   significant image manipulations (e.g., aggressive JPEG compression,")
		fmt.Fprintln(imageWatermarkCmd.Output(), "                   resizing, substantial cropping). Manage expectations accordingly.")
		fmt.Fprintln(imageWatermarkCmd.Output(), "\nExample (visible text watermark on a single file):")
		fmt.Fprintln(imageWatermarkCmd.Output(), "  pdfparser watermark image -input myimage.png -output watermarked.png -type visible -visible.type text -visible.text \"Confidential\" -visible.font /path/to/font.ttf")
		fmt.Fprintln(imageWatermarkCmd.Output(), "\nExample (invisible watermark on all images in a directory):")
		fmt.Fprintln(imageWatermarkCmd.Output(), "  pdfparser watermark image -input ./img_folder -output ./out_folder -type invisible -invisible.data \"secret code\"")
		fmt.Fprintln(imageWatermarkCmd.Output(), "\nExample (steganography watermark on a single file, recommended PNG output):")
		fmt.Fprintln(imageWatermarkCmd.Output(), "  pdfparser watermark image -input myimage.png -output stegano_watermarked.png -type stegano -stegano.data \"my secret data\" -stegano.password \"secure\"")

	}

	if len(os.Args) < 4 { // Not enough args for "watermark image"
		imageWatermarkCmd.Usage()
		return
	}
	imageWatermarkCmd.Parse(os.Args[3:])

	// Validate Inputs
	if *input == "" || *output == "" || *typeFlag == "" {
		fmt.Println("Error: -input, -output, and -type flags are required.")
		imageWatermarkCmd.Usage()
		return
	}

	jobWatermarkType := "" // To be one of "visibleText", "visibleImage", "invisible"

	switch *typeFlag {
	case "visible":
		if *visibleType == "" {
			fmt.Println("Error: -visible.type is required when -type=visible.")
			imageWatermarkCmd.Usage()
			return
		}
		if *visibleType == "text" {
			if *visibleText == "" || *visibleFont == "" {
				fmt.Println("Error: -visible.text and -visible.font are required for text watermarks.")
				imageWatermarkCmd.Usage()
				return
			}
			jobWatermarkType = "visibleText"
		} else if *visibleType == "image" {
			if *visibleImagePath == "" {
				fmt.Println("Error: -visible.imagepath is required for image watermarks.")
				imageWatermarkCmd.Usage()
				return
			}
			jobWatermarkType = "visibleImage"
		} else {
			fmt.Printf("Error: Invalid -visible.type: %s. Must be 'text' or 'image'.\n", *visibleType)
			imageWatermarkCmd.Usage()
			return
		}
	case "invisible":
		if *invisibleData == "" {
			fmt.Println("Error: -invisible.data is required for invisible watermarks.")
			imageWatermarkCmd.Usage()
			return
		}
		jobWatermarkType = "invisible"
	case "stegano":
		if *steganoData == "" {
			fmt.Println("Error: -stegano.data is required for stegano watermarks.")
			imageWatermarkCmd.Usage()
			return
		}
		jobWatermarkType = "steganoBlind"
	default:
		fmt.Printf("Error: Invalid -type: %s. Must be 'visible', 'invisible', or 'stegano'.\n", *typeFlag)
		imageWatermarkCmd.Usage()
		return
	}

	// Collect Input Files
	var inputFiles []string
	inputInfo, err := os.Stat(*input)
	if err != nil {
		fmt.Printf("Error accessing input path '%s': %v\n", *input, err)
		return
	}

	isInputFile := !inputInfo.IsDir()
	outputIsDir := false // Assume output is a file path unless input is a dir or output path ends with /

	if !isInputFile { // Input is a directory
		outputIsDir = true // If input is dir, output must be dir
		err := filepath.WalkDir(*input, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				ext := strings.ToLower(filepath.Ext(path))
				if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
					inputFiles = append(inputFiles, path)
				}
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Error walking input directory '%s': %v\n", *input, err)
			return
		}
		if len(inputFiles) == 0 {
			fmt.Printf("No compatible image files (.png, .jpg, .jpeg) found in directory '%s'\n", *input)
			return
		}
	} else { // Input is a file
		inputFiles = append(inputFiles, *input)
		// Check if output path "looks" like a directory (e.g. ends with / or already exists as dir)
		// This is a simplification. Robust CLI would require explicit -output-dir flag.
		if strings.HasSuffix(*output, string(filepath.Separator)) {
			outputIsDir = true
		} else {
			outputStat, statErr := os.Stat(*output)
			if statErr == nil && outputStat.IsDir() {
				outputIsDir = true
			}
			// If output doesn't exist and doesn't end with separator, assume it's a file path
		}
	}

	outputBase := *output
	if outputIsDir {
		if err := os.MkdirAll(*output, 0755); err != nil {
			fmt.Printf("Error creating output directory '%s': %v\n", *output, err)
			return
		}
		outputBase = *output // Base for joining file names
	} else if len(inputFiles) > 1 {
		fmt.Println("Error: Output must be a directory when processing multiple input files.")
		return
	}

	// Prepare ImageJobs
	var jobs []pdfparser.ImageJob
	for _, inputFile := range inputFiles {
		job := pdfparser.ImageJob{
			InputPath:          inputFile,
			WatermarkType:      jobWatermarkType,
			Opacity:            *visibleOpacity,
			Position:           *visiblePosition,
			Margin:             *visibleMargin,
			Text:               *visibleText,
			FontPath:           *visibleFont,
			FontSize:           *visibleFontSize,
			HexColor:           *visibleColor,
			WatermarkImagePath: *visibleImagePath,
			InvisibleData:      *invisibleData,
			SteganoData:        *steganoData,
			SteganoPassword:    *steganoPassword,
		}
		if outputIsDir {
			job.OutputPath = filepath.Join(outputBase, filepath.Base(inputFile))
		} else {
			job.OutputPath = outputBase // Direct file path
		}
		jobs = append(jobs, job)
	}

	if len(jobs) == 0 {
		fmt.Println("No image processing jobs to run.")
		return
	}

	fmt.Printf("Preparing to process %d image(s) with %d worker(s).\n", len(jobs), *concurrency)

	// Execute Processing
	// Dummy functions for testing CLI structure, replace with actual calls
	// For now, just print the jobs
	// fmt.Printf("Jobs to process: %+v\n", jobs)

	processingErrors := pdfparser.ProcessImagesConcurrently(jobs, *concurrency)

	errorCount := 0
	// The current ProcessImagesConcurrently returns []error, but workers log directly.
	// For a more robust CLI, workers should return errors to be aggregated here.
	// For now, we'll rely on worker logs for individual errors.
	// This simulates counting errors if they were returned.
	for _, job := range jobs {
		// Attempt to stat output file as a proxy for success (very basic)
		if _, statErr := os.Stat(job.OutputPath); statErr != nil {
			// This doesn't mean the job failed, worker might have logged error
			// For now, assume if output file doesn't exist, it's an error for summary
			// errorCount++ // Commenting out as it's not reliable based on current worker design
		}
	}
	// Count errors based on the returned slice from ProcessImagesConcurrently
	// (which is currently empty but designed for future population)
	errorCount = len(processingErrors)

	fmt.Printf("\nProcessing summary:\n")
	fmt.Printf("  Total images scheduled: %d\n", len(jobs))
	// The current errorCount is based on the empty `processingErrors` slice.
	// Workers log errors directly. If `ProcessImagesConcurrently` is updated
	// to collect errors from workers, this count will be meaningful.
	if errorCount > 0 {
		fmt.Printf("  Number of jobs with errors (reported by main): %d\n", errorCount)
		fmt.Println("  Please check worker logs for details on individual job errors.")
	} else {
		fmt.Println("  All jobs dispatched. Check worker logs for individual job status and errors.")
	}
	fmt.Printf("  Output directory/file: %s\n", *output)
}

// This is a placeholder for creating dummy images, if needed for testing.
// It's not directly part of the CLI logic itself but was in the previous main.go for testing.
// Keeping it commented out for now.
func _createDummyImageIfNotExists(path string, width, height int, c color.Color) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("Creating dummy image %s (%dx%d)...\n", path, width, height)
		dummyImg := image.NewNRGBA(image.Rect(0, 0, width, height))
		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				dummyImg.Set(x, y, c)
			}
		}
		if errSave := imaging.Save(dummyImg, path); errSave != nil {
			fmt.Printf("Failed to create dummy image %s: %v\n", path, errSave)
		}
	}
}
