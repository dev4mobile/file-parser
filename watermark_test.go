package pdfparser

import (
	"bytes"
	"image"
	"image/color"
	_ "image/jpeg" // Register JPEG decoder
	_ "image/png"  // Register PNG decoder
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/disintegration/imaging" // For saving test images if needed, or loading
)

// commonFontPath is a system font path to be used in tests.
const commonFontPath = "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"

// checkTestFileExists skips the test if the required file does not exist.
func checkTestFileExists(t *testing.T, path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("Skipping test: Required test data file not found: %s. Please create it.", path)
	}
}

// createDummyImage creates a simple image for testing if real images are missing.
// This is useful for basic tests but not for those requiring specific content.
func createDummyImage(t *testing.T, path string, width, height int, c color.Color) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, c)
		}
	}
	// Try to save it if we are in a context where that helps (e.g. manual run)
	// For automated tests, the checkTestFileExists should ideally run first.
	// os.MkdirAll(filepath.Dir(path), 0755)
	// imaging.Save(img, path) // This might fail in some CI environments
	return img
}

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		name     string
		hexStr   string
		expected color.RGBA
		wantErr  bool
	}{
		{"Valid 6-digit", "#FF0000", color.RGBA{R: 255, G: 0, B: 0, A: 255}, false},
		{"Valid 8-digit", "#00FF0080", color.RGBA{R: 0, G: 255, B: 0, A: 128}, false},
		{"Valid 6-digit no hash", "0000FF", color.RGBA{R: 0, G: 0, B: 255, A: 255}, false},
		{"Valid 8-digit no hash", "ABCDEF12", color.RGBA{R: 0xAB, G: 0xCD, B: 0xEF, A: 0x12}, false},
		{"Invalid length short", "#123", color.RGBA{}, true},
		{"Invalid length long", "#1234567", color.RGBA{}, true},
		{"Invalid char", "#XXYYZZ", color.RGBA{}, true},
		{"Empty string", "", color.RGBA{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHexColor(tt.hexStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseHexColor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("parseHexColor() got = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestApplyTextWatermark(t *testing.T) {
	baseImagePath := filepath.Join("testdata", "base_image.png")
	checkTestFileExists(t, baseImagePath) // Skips if base_image.png is not present

	baseImage, err := imaging.Open(baseImagePath)
	if err != nil {
		t.Fatalf("Failed to load base image %s: %v. Ensure it's a valid PNG.", baseImagePath, err)
	}

	// Check if the common font exists, skip tests if not
	if _, err := os.Stat(commonFontPath); os.IsNotExist(err) {
		t.Skipf("Skipping text watermark tests: common font not found at %s", commonFontPath)
	}

	t.Run("ValidTextWatermark", func(t *testing.T) {
		img, err := ApplyTextWatermark(baseImage, "Test", commonFontPath, 20, "#FF0000FF", 0.8, "Center", 10)
		if err != nil {
			t.Fatalf("ApplyTextWatermark() with valid params returned error: %v", err)
		}
		if img == nil {
			t.Fatal("ApplyTextWatermark() with valid params returned nil image")
		}
		if img.Bounds() != baseImage.Bounds() {
			t.Errorf("Output image bounds %v different from base image bounds %v", img.Bounds(), baseImage.Bounds())
		}
	})

	t.Run("InvalidFontPath", func(t *testing.T) {
		_, err := ApplyTextWatermark(baseImage, "Test", "/non/existent/font.ttf", 20, "#FF0000FF", 0.8, "Center", 10)
		if err == nil {
			t.Error("ApplyTextWatermark() with invalid font path expected error, got nil")
		}
	})

	positions := []string{"TopLeft", "BottomRight", "Center"}
	t.Run("DifferentPositions", func(t *testing.T) {
		for _, pos := range positions {
			t.Run(pos, func(t *testing.T) {
				_, err := ApplyTextWatermark(baseImage, "PosTest", commonFontPath, 15, "#0000FFFF", 1.0, pos, 5)
				if err != nil {
					t.Errorf("ApplyTextWatermark() for position %s returned error: %v", pos, err)
				}
			})
		}
	})

	opacities := []struct { name string; val float64 }{ {"Opaque", 1.0}, {"SemiTransparent", 0.5}, {"Transparent", 0.0} }
	t.Run("DifferentOpacities", func(t *testing.T) {
		for _, op := range opacities {
			t.Run(op.name, func(t *testing.T) {
				_, err := ApplyTextWatermark(baseImage, "OpacityTest", commonFontPath, 15, "#00FF00FF", op.val, "Center", 5)
				if err != nil {
					t.Errorf("ApplyTextWatermark() for opacity %.1f returned error: %v", op.val, err)
				}
			})
		}
	})
}


func TestApplyImageWatermark(t *testing.T) {
	baseImagePath := filepath.Join("testdata", "base_image.png")
	logoWatermarkPath := filepath.Join("testdata", "logo_watermark.png")

	checkTestFileExists(t, baseImagePath)
	checkTestFileExists(t, logoWatermarkPath)

	baseImage, err := imaging.Open(baseImagePath)
	if err != nil {
		t.Fatalf("Failed to load base image %s: %v. Ensure it's a valid PNG.", baseImagePath, err)
	}
	// Watermark image is loaded by the function itself, so no need to load it here.

	t.Run("ValidImageWatermark", func(t *testing.T) {
		img, err := ApplyImageWatermark(baseImage, logoWatermarkPath, 0.7, "Center", 10)
		if err != nil {
			t.Fatalf("ApplyImageWatermark() with valid params returned error: %v", err)
		}
		if img == nil {
			t.Fatal("ApplyImageWatermark() with valid params returned nil image")
		}
		if img.Bounds() != baseImage.Bounds() {
			t.Errorf("Output image bounds %v different from base image bounds %v", img.Bounds(), baseImage.Bounds())
		}
	})

	t.Run("InvalidWatermarkPath", func(t *testing.T) {
		_, err := ApplyImageWatermark(baseImage, "/non/existent/watermark.png", 0.7, "Center", 10)
		if err == nil {
			t.Error("ApplyImageWatermark() with invalid watermark path expected error, got nil")
		}
	})
	
	positions := []string{"TopLeft", "BottomRight", "Center"}
	t.Run("DifferentPositions", func(t *testing.T) {
		for _, pos := range positions {
			t.Run(pos, func(t *testing.T) {
				_, err := ApplyImageWatermark(baseImage, logoWatermarkPath, 1.0, pos, 5)
				if err != nil {
					t.Errorf("ApplyImageWatermark() for position %s returned error: %v", pos, err)
				}
			})
		}
	})

	opacities := []struct { name string; val float64 }{ {"Opaque", 1.0}, {"SemiTransparent", 0.5}, {"Transparent", 0.0} }
	t.Run("DifferentOpacities", func(t *testing.T) {
		for _, op := range opacities {
			t.Run(op.name, func(t *testing.T) {
				_, err := ApplyImageWatermark(baseImage, logoWatermarkPath, op.val, "Center", 5)
				if err != nil {
					// Allow error for opacity 0.0 if the library treats it as fully transparent and potentially problematic for some operations
					// However, the current implementation should handle it gracefully by drawing a fully transparent image.
					t.Errorf("ApplyImageWatermark() for opacity %.1f returned error: %v", op.val, err)
				}
			})
		}
	})
}

func TestInvisibleWatermarkingCycle(t *testing.T) {
	testCases := []struct {
		name        string
		imagePath   string
		imageFormat string // "png" or "jpeg"
	}{
		{"PNG_Cycle", filepath.Join("testdata", "base_image.png"), "png"},
		{"JPG_Cycle", filepath.Join("testdata", "base_image.jpg"), "jpeg"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			checkTestFileExists(t, tc.imagePath)
			baseImage, err := imaging.Open(tc.imagePath)
			if err != nil {
				t.Fatalf("Failed to load base image %s: %v.", tc.imagePath, err)
			}

			secretMessage := "Hello, Steganography! This is a test for " + strings.ToUpper(tc.imageFormat) + " images."

			// Embed
			buffer, errEmbed := EmbedInvisibleWatermark(baseImage, secretMessage)
			if errEmbed != nil {
				t.Fatalf("EmbedInvisibleWatermark() returned error: %v", errEmbed)
			}
			if buffer == nil || buffer.Len() == 0 {
				t.Fatal("EmbedInvisibleWatermark() returned nil or empty buffer")
			}

			// Decode the image from this buffer
			imgData, _, errDecode := image.Decode(bytes.NewReader(buffer.Bytes()))
			if errDecode != nil {
				t.Fatalf("image.Decode() from buffer returned error: %v", errDecode)
			}

			// Extract
			// The steganography library might be sensitive to image format changes
			// or specific image types (e.g., expects *image.RGBA).
			// image.Decode might return image.Image; ensure it's compatible with ExtractInvisibleWatermark.
			// For JPGs, LSB steganography is lossy and often unreliable.
			// The `auyer/steganography` library uses LSB and is primarily designed for lossless formats like PNG.
			// It might "work" on JPEGs by chance if the JPEG compression didn't alter the LSBs too much,
			// but this is not guaranteed.
			if tc.imageFormat == "jpeg" {
				// t.Logf("Note: Steganography on JPEGs is inherently unreliable due to lossy compression. This test may fail.")
                // Let's try, but be aware it might fail not due to code error but format limitations.
			}

			extractedMessage, errExtract := ExtractInvisibleWatermark(imgData)
			if errExtract != nil {
				if tc.imageFormat == "jpeg" && strings.Contains(errExtract.Error(), "message size is zero") {
					t.Logf("Known issue: ExtractInvisibleWatermark on JPG failed as expected due to lossy compression or format incompatibility with LSB: %v", errExtract)
					t.Skip("Skipping JPG steganography message content assertion due to inherent unreliability with LSB on lossy formats.")
				}
				t.Fatalf("ExtractInvisibleWatermark() returned error: %v", errExtract)
			}

			if extractedMessage != secretMessage {
				t.Errorf("Extracted message does not match original.\nOriginal: '%s'\nExtracted: '%s'", secretMessage, extractedMessage)
			}
		})
	}
}

// Helper function to create a dummy font file for local testing if needed.
// Not used in automated tests as we rely on commonFontPath or skip.
func _createDummyFontFile(t *testing.T, path string) {
	// This is very simplified and NOT a valid TTF.
	// For real tests, a minimal valid TTF would be needed.
	// Or, as done in tests, use a known system font.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(path), 0755)
		err := os.WriteFile(path, []byte("dummyfont"), 0644)
		if err != nil {
			t.Logf("Could not create dummy font file for local testing: %v", err)
		}
	}
}

// Helper function to create a dummy image file for local testing.
// Not used in automated tests as we rely on checkTestFileExists or skip.
func _createDummyImageFile(t *testing.T, path string, width, height int, c color.Color) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		img := image.NewNRGBA(image.Rect(0, 0, width, height))
		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				img.Set(x, y, c)
			}
		}
		os.MkdirAll(filepath.Dir(path), 0755)
		err := imaging.Save(img, path)
		if err != nil {
			t.Logf("Could not create dummy image file %s for local testing: %v", path, err)
		}
	}
}

func TestMain(m *testing.M) {
	// // Setup: Create dummy files if you want tests to run without manual setup.
	// // However, for CI, it's better to have these files checked in or ensure
	// // the environment provides them. For this task, we use t.Skip if files are missing.
	//
	// // _createDummyFontFile(nil, filepath.Join("testdata", "sample.ttf")) // Not a valid TTF
	// _createDummyImageFile(nil, filepath.Join("testdata", "base_image.png"), 100, 100, color.NRGBA{R:200, G:200, B:220, A:255})
	// _createDummyImageFile(nil, filepath.Join("testdata", "base_image.jpg"), 100, 100, color.NRGBA{R:220, G:200, B:200, A:255})
	// _createDummyImageFile(nil, filepath.Join("testdata", "logo_watermark.png"), 20, 20, color.NRGBA{R:0, G:0, B:255, A:128}) // Blue, semi-transparent

	exitCode := m.Run()
	// Teardown: Clean up dummy files if created.
	// os.Remove(filepath.Join("testdata", "sample.ttf"))
	// os.Remove(filepath.Join("testdata", "base_image.png"))
	// os.Remove(filepath.Join("testdata", "base_image.jpg"))
	// os.Remove(filepath.Join("testdata", "logo_watermark.png"))
	os.Exit(exitCode)
}
