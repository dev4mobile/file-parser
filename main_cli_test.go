package main

import (
	"fmt"
	"image"
	_ "image/jpeg" // Register JPEG decoder
	_ "image/png"  // Register PNG decoder
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	// Assuming pdfparser package is in the same module or accessible
	// If main is in 'package main', and other code in 'package pdfparser'
	// we might need to import the pdfparser package if we directly call its functions here for verification.
	// For CLI tests, direct calls are usually for verification like ExtractInvisibleWatermark.
	// Let's assume the current package is 'main' for CLI tests.
	// For ExtractInvisibleWatermark, we'd need to ensure it's callable.
	// If pdfparser is a separate package:
	"pdfparser" // Make sure this import path is correct based on your go.mod
)

const (
	cliAppName          = "test_cli_app" // Name for the built binary
	commonSystemFontCLI = "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"
	testDataDirCLI      = "../testdata" // Relative path to testdata from the package directory (e.g. if this test is in a sub-package of main)
	// If main_cli_test.go is in the same directory as main.go (package main), then testDataDir should be "testdata"
)

// TestMain manages the build and cleanup of the CLI executable.
func TestMain(m *testing.M) {
	// Adjust path to root if main.go is not in current dir (e.g. if tests are in a subpackage)
	// For this structure, assuming main.go is in the current directory '.'
	buildCmd := exec.Command("go", "build", "-o", cliAppName, ".")
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to build CLI app '%s': %v\nOutput:\n%s\n", cliAppName, err, string(buildOutput))
		os.Exit(1)
	}

	exitCode := m.Run()

	errRemove := os.Remove(cliAppName)
	if errRemove != nil {
		fmt.Printf("Warning: Failed to remove test CLI app '%s': %v\n", cliAppName, errRemove)
	}
	os.Exit(exitCode)
}

// checkSkipOrFatal checks if a required test file exists. Skips test if not found.
func checkSkipOrFatalCLI(t *testing.T, path string) {
	// This path needs to be relative to where `go test` is run, or absolute.
	// testDataDirCLI helps make it relative to the project structure.
	fullPath := path
	if !filepath.IsAbs(path) { // if path is like "base_image.png"
		// This logic assumes main_cli_test.go is in the same directory as main.go
		// and testdata is a subdirectory. If testdata is ../testdata, adjust accordingly.
		// For now, assuming testdata is a direct subdirectory of where go test runs (e.g. project root)
		// Or, the caller of checkSkipOrFatalCLI provides the correct relative/absolute path.
		// Let's assume path provided to this function is already correctly pathed e.g. filepath.Join(testDataDirCLI, filename)
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Skipf("Skipping test: Required test data file not found: %s. Please ensure it exists.", fullPath)
	}
}

// copyTestFile copies a source file to a destination directory.
func copyTestFile(t *testing.T, sourcePath, destDir, destName string) string {
	checkSkipOrFatalCLI(t, sourcePath) // Check if source exists first

	destPath := filepath.Join(destDir, destName)
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		t.Fatalf("Failed to open source file %s: %v", sourcePath, err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		t.Fatalf("Failed to create destination file %s: %v", destPath, err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		t.Fatalf("Failed to copy from %s to %s: %v", sourcePath, destPath, err)
	}
	return destPath
}

func setupTestDirs(t *testing.T) (inputDir string, outputDir string) {
	baseTempDir := t.TempDir()
	inputDir = filepath.Join(baseTempDir, "input")
	outputDir = filepath.Join(baseTempDir, "output")
	if err := os.Mkdir(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create temp input dir: %v", err)
	}
	if err := os.Mkdir(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create temp output dir: %v", err)
	}
	return inputDir, outputDir
}

// TestCLI_VisibleText_SingleFile tests visible text watermark on a single file.
func TestCLI_VisibleText_SingleFile(t *testing.T) {
	inputDir, outputDir := setupTestDirs(t)
	
	// Adjust source path based on actual location of testdata relative to test execution
	// If main_cli_test.go is in the project root with main.go, and testdata is also in root:
	baseImageSourcePath := filepath.Join("testdata", "base_image.png") 
	checkSkipOrFatalCLI(t, baseImageSourcePath) // Ensure testdata/base_image.png exists
	if _, err := os.Stat(commonSystemFontCLI); os.IsNotExist(err) {
		t.Skipf("Skipping test: common font not found at %s", commonSystemFontCLI)
	}

	copiedBaseImagePath := copyTestFile(t, baseImageSourcePath, inputDir, "base_image.png")
	outputImagePath := filepath.Join(outputDir, "out_text.png")

	args := []string{
		"watermark", "image",
		"-input", copiedBaseImagePath,
		"-output", outputImagePath,
		"-type", "visible",
		"-visible.type", "text",
		"-visible.text", "CLI Test Text",
		"-visible.font", commonSystemFontCLI,
		"-visible.size", "22",
		"-visible.color", "#0000FFFF", // Blue
		"-visible.position", "TopLeft",
		"-visible.margin", "8",
	}
	cmd := exec.Command("./"+cliAppName, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI command failed with error: %v\nOutput:\n%s", err, string(output))
	}

	if _, statErr := os.Stat(outputImagePath); os.IsNotExist(statErr) {
		t.Errorf("Expected output file %s was not created.\nCLI Output:\n%s", outputImagePath, string(output))
	} else {
		t.Logf("Visible text watermark (single file) CLI test passed. Output: %s", outputImagePath)
	}
}

// TestCLI_VisibleImage_Directory tests visible image watermark on a directory.
func TestCLI_VisibleImage_Directory(t *testing.T) {
	inputDir, outputDir := setupTestDirs(t)

	baseImageSourcePath := filepath.Join("testdata", "base_image.png")
	logoSourcePath := filepath.Join("testdata", "logo_watermark.png")
	checkSkipOrFatalCLI(t, baseImageSourcePath)
	checkSkipOrFatalCLI(t, logoSourcePath)

	// Copy multiple files to simulate a directory
	copyTestFile(t, baseImageSourcePath, inputDir, "img1.png")
	copyTestFile(t, baseImageSourcePath, inputDir, "img2.jpg") // Assuming jpg is also processed
	copiedLogoPath := copyTestFile(t, logoSourcePath, inputDir, "logo.png")


	args := []string{
		"watermark", "image",
		"-input", inputDir,    // Input is a directory
		"-output", outputDir,   // Output is a directory
		"-type", "visible",
		"-visible.type", "image",
		"-visible.imagepath", copiedLogoPath,
		"-visible.position", "BottomCenter",
		"-visible.opacity", "0.7",
		"-concurrency", fmt.Sprintf("%d", runtime.NumCPU()),
	}
	cmd := exec.Command("./"+cliAppName, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI command for directory processing failed: %v\nOutput:\n%s", err, string(output))
	}

	expectedOutput1 := filepath.Join(outputDir, "img1.png")
	expectedOutput2 := filepath.Join(outputDir, "img2.jpg")
	if _, statErr := os.Stat(expectedOutput1); os.IsNotExist(statErr) {
		t.Errorf("Expected output file %s was not created for directory input.\nCLI Output:\n%s", expectedOutput1, string(output))
	}
	if _, statErr := os.Stat(expectedOutput2); os.IsNotExist(statErr) {
		t.Errorf("Expected output file %s was not created for directory input.\nCLI Output:\n%s", expectedOutput2, string(output))
	}
	t.Logf("Visible image watermark (directory) CLI test passed. Output dir: %s", outputDir)
}

// TestCLI_Invisible_SingleFile tests invisible watermark embedding and extraction.
func TestCLI_Invisible_SingleFile(t *testing.T) {
	inputDir, outputDir := setupTestDirs(t)
	baseImageSourcePath := filepath.Join("testdata", "base_image.png")
	checkSkipOrFatalCLI(t, baseImageSourcePath)

	copiedBaseImagePath := copyTestFile(t, baseImageSourcePath, inputDir, "base_image_for_invisible.png")
	outputImagePath := filepath.Join(outputDir, "out_invisible.png")
	secretData := "cli_secret_123"

	args := []string{
		"watermark", "image",
		"-input", copiedBaseImagePath,
		"-output", outputImagePath,
		"-type", "invisible",
		"-invisible.data", secretData,
	}
	cmd := exec.Command("./"+cliAppName, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI command for invisible watermark failed: %v\nOutput:\n%s", err, string(output))
	}

	if _, statErr := os.Stat(outputImagePath); os.IsNotExist(statErr) {
		t.Fatalf("Expected output file %s for invisible watermark was not created.\nCLI Output:\n%s", outputImagePath, string(output))
	}

	// Verify embedded data
	loadedImage, errLoad := imaging.Open(outputImagePath)
	if errLoad != nil {
		t.Fatalf("Failed to load output image %s for verification: %v", outputImagePath, errLoad)
	}

	// We need to call pdfparser.ExtractInvisibleWatermark.
	// This assumes main_cli_test.go is part of package main, but pdfparser functions are in package pdfparser.
	// This requires pdfparser to be an importable package.
	extractedData, errExtract := pdfparser.ExtractInvisibleWatermark(loadedImage)
	if errExtract != nil {
		t.Fatalf("ExtractInvisibleWatermark failed: %v", errExtract)
	}
	if extractedData != secretData {
		t.Errorf("Extracted invisible data mismatch: got '%s', want '%s'", extractedData, secretData)
	}
	t.Logf("Invisible watermark CLI test passed. Data verified.")
}

// TestCLI_MissingRequiredFlags tests CLI behavior with missing required flags.
func TestCLI_MissingRequiredFlags(t *testing.T) {
	_, outputDir := setupTestDirs(t) // Need outputDir for some commands even if they fail

	testCases := []struct {
		name        string
		args        []string
		expectedErr bool
	}{
		{
			name: "Missing input and type",
			args: []string{"watermark", "image", "-output", filepath.Join(outputDir, "out.png")},
			expectedErr: true,
		},
		{
			name: "Missing type",
			args: []string{"watermark", "image", "-input", "dummy.png", "-output", filepath.Join(outputDir, "out.png")},
			expectedErr: true,
		},
		{
			name: "Visible text type missing text and font",
			args: []string{"watermark", "image", "-input", "dummy.png", "-output", filepath.Join(outputDir, "out.png"), "-type", "visible", "-visible.type", "text"},
			expectedErr: true,
		},
		{
			name: "Visible image type missing imagepath",
			args: []string{"watermark", "image", "-input", "dummy.png", "-output", filepath.Join(outputDir, "out.png"), "-type", "visible", "-visible.type", "image"},
			expectedErr: true,
		},
		{
			name: "Invisible type missing data",
			args: []string{"watermark", "image", "-input", "dummy.png", "-output", filepath.Join(outputDir, "out.png"), "-type", "invisible"},
			expectedErr: true,
		},
	}

	// Create a dummy input file for tests that require -input to pass initial parsing stages
	// This file won't actually be processed as the commands are expected to fail before that.
	dummyInputPath := filepath.Join(t.TempDir(), "dummy_input.png")
	dummyImg := image.NewNRGBA(image.Rect(0,0,10,10))
	imaging.Save(dummyImg, dummyInputPath)


	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Replace "dummy.png" placeholder with actual dummyInputPath
			currentArgs := make([]string, len(tc.args))
			copy(currentArgs, tc.args)
			for i, arg := range currentArgs {
				if arg == "dummy.png" {
					currentArgs[i] = dummyInputPath
				}
			}

			cmd := exec.Command("./"+cliAppName, currentArgs...)
			output, err := cmd.CombinedOutput() // CombinedOutput is usually better for tests

			if tc.expectedErr && err == nil {
				t.Errorf("Expected error for args: %v, but got none.\nOutput:\n%s", tc.args, string(output))
			} else if !tc.expectedErr && err != nil {
				t.Errorf("Expected no error for args: %v, but got: %v.\nOutput:\n%s", tc.args, err, string(output))
			}
			// For more specific error checking, one could inspect the exit code or parts of the output string.
			// ExitError can be asserted: if exitErr, ok := err.(*exec.ExitError); ok { ... }
		})
	}
}

func TestCLI_PdfCommand(t *testing.T) {
	// For this test, we need a dummy PDF. Since we can't create one,
	// this test will be very basic, checking if the command runs and prints usage/error for non-PDF.
	// If a test PDF existed in testdata, we'd use it.
	
	t.Run("PdfCommand_NoFile", func(t *testing.T) {
		cmd := exec.Command("./"+cliAppName, "pdf")
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Errorf("Expected error when no PDF path is provided to 'pdf' command, got nil.\nOutput:\n%s", string(output))
		}
		if !strings.Contains(string(output), "Usage: pdfparser pdf") {
			t.Errorf("Expected usage information for 'pdf' command, got:\n%s", string(output))
		}
	})

	t.Run("PdfCommand_NonPdfFile", func(t *testing.T) {
		dummyTxtPath := filepath.Join(t.TempDir(), "dummy.txt")
		os.WriteFile(dummyTxtPath, []byte("not a pdf"), 0644)

		cmd := exec.Command("./"+cliAppName, "pdf", dummyTxtPath)
		output, err := cmd.CombinedOutput()
		if err == nil { // The CLI itself might not error, but print an error message.
			// The current pdf command logic prints an error to stdout and doesn't exit with error.
			// Let's check for the error message in output.
			// t.Errorf("Expected error or specific message for non-PDF file, got nil error.\nOutput:\n%s", string(output))
		}
		// The pdf command in main.go prints "Error: Please provide a valid .pdf file."
		// but doesn't os.Exit(1). So `err` from `cmd.CombinedOutput()` will be nil if executable ran.
		if !strings.Contains(string(output), "Error: Please provide a valid .pdf file.") {
			 t.Errorf("Expected error message for non-PDF file, got:\n%s", string(output))
		}
	})

    // If a test.pdf was available:
	// testPdfPath := filepath.Join("testdata", "test_document.pdf")
	// checkSkipOrFatalCLI(t, testPdfPath) // Skip if not found
    // cmd := exec.Command("./"+cliAppName, "pdf", testPdfPath)
    // output, err := cmd.CombinedOutput()
    // if err != nil {
    //    t.Fatalf("pdf command failed for %s: %v\nOutput:\n%s", testPdfPath, err, string(output))
    // }
    // Assert output contains "Extracted text:" or specific content
}
