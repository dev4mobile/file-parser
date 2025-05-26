package pdfparser

import (
	_ "image/jpeg" // Register JPEG decoder for loading test images if any are jpeg
	_ "image/png"  // Register PNG decoder
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/disintegration/imaging"
)

// commonFontPathBatch is a system font path to be used in batch tests.
const commonFontPathBatch = "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf" // Duplicated from watermark_test for now

// checkTestFileExistsBatch skips the test if the required file does not exist.
func checkTestFileExistsBatch(t *testing.T, path string) { // Duplicated from watermark_test for now
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("Skipping test: Required test data file not found: %s. Please create it or ensure it's in testdata/.", path)
	}
}

func TestProcessImagesConcurrently(t *testing.T) {
	// 2. Test File Handling: Check for required files and skip if not found.
	baseImagePath := filepath.Join("..", "testdata", "base_image.png") // Assuming testdata is in parent dir relative to package dir
	logoWatermarkPath := filepath.Join("..", "testdata", "logo_watermark.png")

	checkTestFileExistsBatch(t, baseImagePath)
	checkTestFileExistsBatch(t, logoWatermarkPath)

	if _, err := os.Stat(commonFontPathBatch); os.IsNotExist(err) {
		t.Skipf("Skipping some tests: common font not found at %s", commonFontPathBatch)
	}

	// a. Create a temporary output directory
	tempDir := t.TempDir()
	t.Logf("Temporary output directory: %s", tempDir)

	// b. Define a slice of ImageJob structs
	jobs := []ImageJob{
		{ // i. Job 1: Visible text watermark
			InputPath:     baseImagePath,
			OutputPath:    filepath.Join(tempDir, "text_out.png"),
			WatermarkType: "visibleText",
			Text:          "Batch Test Text",
			FontPath:      commonFontPathBatch,
			FontSize:      24,
			HexColor:      "#00FF00", // Green
			Opacity:       0.75,
			Position:      "Center",
			Margin:        10,
		},
		{ // ii. Job 2: Visible image watermark
			InputPath:          baseImagePath,
			OutputPath:         filepath.Join(tempDir, "image_out.png"),
			WatermarkType:      "visibleImage",
			WatermarkImagePath: logoWatermarkPath,
			Opacity:            0.6,
			Position:           "BottomLeft",
			Margin:             5,
		},
		{ // iii. Job 3: Invisible watermark
			InputPath:     baseImagePath,
			OutputPath:    filepath.Join(tempDir, "invisible_out.png"),
			WatermarkType: "invisible",
			InvisibleData: "test_secret_data",
		},
	}

	// c. Call ProcessImagesConcurrently
	// Using runtime.NumCPU() as per subtask, can be a fixed number like 2 for more predictable tests.
	numWorkers := runtime.NumCPU()
	if numWorkers < 1 {
		numWorkers = 1
	}

	t.Logf("Starting ProcessImagesConcurrently with %d workers for %d jobs.", numWorkers, len(jobs))
	processingErrors := ProcessImagesConcurrently(jobs, numWorkers)

	// d. Assert that the returned []error slice is empty
	if len(processingErrors) > 0 {
		t.Errorf("ProcessImagesConcurrently returned %d errors:", len(processingErrors))
		for i, err := range processingErrors {
			t.Errorf("  Error %d: %v", i+1, err)
		}
	}
	// Note: The current implementation of ProcessImagesConcurrently and worker logs errors
	// directly and doesn't populate the returned error slice. This assertion might not catch
	// errors logged by workers. For a more robust test, workers should return errors up.
	// For now, we proceed to check file creation as the primary success indicator.

	// e. For each job, verify that the corresponding output file was created
	for _, job := range jobs {
		if _, err := os.Stat(job.OutputPath); os.IsNotExist(err) {
			t.Errorf("Expected output file %s was not created for job input %s", job.OutputPath, job.InputPath)
		} else {
			t.Logf("Verified output file exists: %s", job.OutputPath)
		}
	}

	// f. Verification for Invisible Watermark
	invisibleOutputJob := jobs[2] // Assuming the 3rd job is the invisible watermark one
	if _, err := os.Stat(invisibleOutputJob.OutputPath); err == nil {
		// 1. Load the tempDir/invisible_out.png image.
		loadedInvisibleImage, errLoad := imaging.Open(invisibleOutputJob.OutputPath)
		if errLoad != nil {
			t.Errorf("Failed to load output image %s for invisible watermark verification: %v", invisibleOutputJob.OutputPath, errLoad)
		} else {
			// 2. Call ExtractInvisibleWatermark on it.
			extractedData, errExtract := ExtractInvisibleWatermark(loadedInvisibleImage)
			if errExtract != nil {
				t.Errorf("ExtractInvisibleWatermark failed for %s: %v", invisibleOutputJob.OutputPath, errExtract)
			} else {
				// 3. Assert that the extracted data matches "test_secret_data".
				if extractedData != "test_secret_data" {
					t.Errorf("Extracted invisible data mismatch: got '%s', want '%s'", extractedData, "test_secret_data")
				} else {
					t.Logf("Successfully verified invisible watermark data for %s", invisibleOutputJob.OutputPath)
				}
			}
		}
	} else {
		t.Errorf("Skipping invisible watermark verification because output file %s was not created.", invisibleOutputJob.OutputPath)
	}
}
