package pdfparser

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/disintegration/imaging"
)

const (
	robustnessWatermarkMessage = "RobustnessCheckMessage123!@#"
	robustnessTestPassword     = "testpassword"
	baseImageRobustnessPath    = "testdata/base_image_robustness.png" // Created in previous steps
	robustnessOutputDir        = "testdata/robustness_output"
)

// Helper function to load an image
func loadTestImage(t *testing.T, path string) image.Image {
	t.Helper()
	img, err := imaging.Open(path)
	if err != nil {
		t.Fatalf("Failed to load image %s: %v", path, err)
	}
	return img
}

// Helper function to save an image (uses imaging.Save for format flexibility)
func saveTestImage(t *testing.T, img image.Image, name string) string {
	t.Helper()
	if _, err := os.Stat(robustnessOutputDir); os.IsNotExist(err) {
		os.MkdirAll(robustnessOutputDir, 0755)
	}
	outputPath := filepath.Join(robustnessOutputDir, name)
	err := imaging.Save(img, outputPath)
	if err != nil {
		t.Fatalf("Failed to save image %s: %v", outputPath, err)
	}
	t.Logf("Saved image: %s", outputPath)
	return outputPath
}

// TestRobustnessBaseline embeds the watermark and tries to extract it.
// The output watermarked_original.png is used by subsequent tests.
func TestRobustnessBaseline(t *testing.T) {
	baseImg := loadTestImage(t, baseImageRobustnessPath)
	watermarkedOriginalPath := filepath.Join(robustnessOutputDir, "watermarked_original.png")

	t.Log("TestRobustnessBaseline: Embedding watermark...")
	embeddedImg, err := EmbedSteganoWatermark(baseImg, []byte(robustnessWatermarkMessage), robustnessTestPassword)
	if err != nil {
		t.Fatalf("Baseline: Embed failed: %v", err)
	}
	saveTestImage(t, embeddedImg, "watermarked_original.png") // Save for other tests

	t.Log("TestRobustnessBaseline: Extracting watermark from original watermarked image...")
	loadedWatermarkedImg := loadTestImage(t, watermarkedOriginalPath)
	extractedData, err := ExtractSteganoWatermark(loadedWatermarkedImg, robustnessTestPassword)
	if err != nil {
		t.Errorf("Baseline: Extract failed: %v", err)
		return
	}
	if string(extractedData) == robustnessWatermarkMessage {
		t.Logf("Baseline: Extraction successful, message MATCHES.")
	} else {
		t.Errorf("Baseline: Extraction successful, but message MISMATCH. Got: '%s', Want: '%s'", string(extractedData), robustnessWatermarkMessage)
	}
}

func TestRobustnessJPEGCompression(t *testing.T) {
	watermarkedOriginalPath := filepath.Join(robustnessOutputDir, "watermarked_original.png")
	if _, err := os.Stat(watermarkedOriginalPath); os.IsNotExist(err) {
		t.Skip("Skipping JPEG test as watermarked_original.png not found. Run TestRobustnessBaseline first.")
		return
	}
	baseWatermarkedImg := loadTestImage(t, watermarkedOriginalPath)

	jpegQualities := []int{75, 50}

	for _, quality := range jpegQualities {
		t.Run(fmt.Sprintf("JPEG_Q%d", quality), func(t *testing.T) {
			t.Logf("Testing JPEG Compression Q%d: Applying JPEG compression...", quality)

			var buf bytes.Buffer
			err := jpeg.Encode(&buf, baseWatermarkedImg, &jpeg.Options{Quality: quality})
			if err != nil {
				t.Fatalf("JPEG Q%d: Failed to encode to JPEG: %v", quality, err)
			}

			jpegImg, _, err := image.Decode(&buf)
			if err != nil {
				t.Fatalf("JPEG Q%d: Failed to decode JPEG from buffer: %v", quality, err)
			}
			saveTestImage(t, jpegImg, fmt.Sprintf("watermarked_jpeg_q%d.jpg", quality))

			t.Logf("Testing JPEG Compression Q%d: Attempting to extract watermark...", quality)
			extractedData, err := ExtractSteganoWatermark(jpegImg, robustnessTestPassword)
			if err != nil {
				t.Logf("JPEG Q%d: Extract FAILED: %v", quality, err)
			} else {
				if string(extractedData) == robustnessWatermarkMessage {
					t.Logf("JPEG Q%d: Extract successful, message MATCHES.", quality)
				} else {
					t.Logf("JPEG Q%d: Extract successful, but message MISMATCH. Got: '%s', Want: '%s'", quality, string(extractedData), robustnessWatermarkMessage)
				}
			}
		})
	}
}

func TestRobustnessResizing(t *testing.T) {
	watermarkedOriginalPath := filepath.Join(robustnessOutputDir, "watermarked_original.png")
	if _, err := os.Stat(watermarkedOriginalPath); os.IsNotExist(err) {
		t.Skip("Skipping Resizing test as watermarked_original.png not found. Run TestRobustnessBaseline first.")
		return
	}
	baseWatermarkedImg := loadTestImage(t, watermarkedOriginalPath)
	originalWidth := baseWatermarkedImg.Bounds().Dx()
	originalHeight := baseWatermarkedImg.Bounds().Dy()

	// Resize to 50%
	t.Log("Testing Resizing 50%: Resizing image...")
	resized50Img := imaging.Resize(baseWatermarkedImg, originalWidth/2, originalHeight/2, imaging.Lanczos)
	saveTestImage(t, resized50Img, "watermarked_resized_50pct.png")

	t.Log("Testing Resizing 50%: Attempting to extract watermark...")
	extractedData50, err50 := ExtractSteganoWatermark(resized50Img, robustnessTestPassword)
	if err50 != nil {
		t.Logf("Resizing 50%: Extract FAILED: %v", err50)
	} else {
		if string(extractedData50) == robustnessWatermarkMessage {
			t.Logf("Resizing 50%: Extract successful, message MATCHES.")
		} else {
			t.Logf("Resizing 50%: Extract successful, but message MISMATCH. Got: '%s', Want: '%s'", string(extractedData50), robustnessWatermarkMessage)
		}
	}

	// (Optional) Resize back
	t.Log("Testing Resizing Back: Resizing image back to original dimensions...")
	resizedBackImg := imaging.Resize(resized50Img, originalWidth, originalHeight, imaging.Lanczos)
	saveTestImage(t, resizedBackImg, "watermarked_resized_back.png")

	t.Log("Testing Resizing Back: Attempting to extract watermark...")
	extractedDataBack, errBack := ExtractSteganoWatermark(resizedBackImg, robustnessTestPassword)
	if errBack != nil {
		t.Logf("Resizing Back: Extract FAILED: %v", errBack)
	} else {
		if string(extractedDataBack) == robustnessWatermarkMessage {
			t.Logf("Resizing Back: Extract successful, message MATCHES.")
		} else {
			t.Logf("Resizing Back: Extract successful, but message MISMATCH. Got: '%s', Want: '%s'", string(extractedDataBack), robustnessWatermarkMessage)
		}
	}
}

func TestRobustnessCropping(t *testing.T) {
	watermarkedOriginalPath := filepath.Join(robustnessOutputDir, "watermarked_original.png")
	if _, err := os.Stat(watermarkedOriginalPath); os.IsNotExist(err) {
		t.Skip("Skipping Cropping test as watermarked_original.png not found. Run TestRobustnessBaseline first.")
		return
	}
	baseWatermarkedImg := loadTestImage(t, watermarkedOriginalPath)
	width := baseWatermarkedImg.Bounds().Dx()
	height := baseWatermarkedImg.Bounds().Dy()

	// Crop 5% from borders - imaging.CropCenter might be easiest if it fits the need.
	// CropCenter crops to a new width and height from the center.
	// To remove 5% from each border, new width = width * 0.9, new height = height * 0.9
	newWidth := int(float64(width) * 0.90)
	newHeight := int(float64(height) * 0.90)

	t.Logf("Testing Cropping: Cropping image from %dx%d to %dx%d...", width, height, newWidth, newHeight)
	croppedImg := imaging.CropCenter(baseWatermarkedImg, newWidth, newHeight)
	saveTestImage(t, croppedImg, "watermarked_cropped.png")

	t.Log("Testing Cropping: Attempting to extract watermark...")
	extractedData, err := ExtractSteganoWatermark(croppedImg, robustnessTestPassword)
	if err != nil {
		t.Logf("Cropping: Extract FAILED: %v", err)
	} else {
		if string(extractedData) == robustnessWatermarkMessage {
			t.Logf("Cropping: Extract successful, message MATCHES.")
		} else {
			t.Logf("Cropping: Extract successful, but message MISMATCH. Got: '%s', Want: '%s'", string(extractedData), robustnessWatermarkMessage)
		}
	}
}

func TestRobustnessBrightness(t *testing.T) {
	watermarkedOriginalPath := filepath.Join(robustnessOutputDir, "watermarked_original.png")
	if _, err := os.Stat(watermarkedOriginalPath); os.IsNotExist(err) {
		t.Skip("Skipping Brightness test as watermarked_original.png not found. Run TestRobustnessBaseline first.")
		return
	}
	baseWatermarkedImg := loadTestImage(t, watermarkedOriginalPath)

	brightnessPercentage := 10.0 // Adjust brightness by +10%
	t.Logf("Testing Brightness: Adjusting brightness by +%.1f%%...", brightnessPercentage)
	brightenedImg := imaging.AdjustBrightness(baseWatermarkedImg, brightnessPercentage)
	saveTestImage(t, brightenedImg, "watermarked_brightness.png")

	t.Log("Testing Brightness: Attempting to extract watermark...")
	extractedData, err := ExtractSteganoWatermark(brightenedImg, robustnessTestPassword)
	if err != nil {
		t.Logf("Brightness: Extract FAILED: %v", err)
	} else {
		if string(extractedData) == robustnessWatermarkMessage {
			t.Logf("Brightness: Extract successful, message MATCHES.")
		} else {
			t.Logf("Brightness: Extract successful, but message MISMATCH. Got: '%s', Want: '%s'", string(extractedData), robustnessWatermarkMessage)
		}
	}
}
