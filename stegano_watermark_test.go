package pdfparser

import (
	"image"
	"image/color"
	"strings"
	"testing"
)

// TestEmbedSteganoWatermark tests the EmbedSteganoWatermark function.
func TestEmbedSteganoWatermark(t *testing.T) {
	// Create a simple dummy image for testing.
	baseImg := image.NewRGBA(image.Rect(0, 0, 100, 50))
	for x := 0; x < 100; x++ {
		for y := 0; y < 50; y++ {
			baseImg.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 100, A: 255})
		}
	}

	watermarkData := []byte("test watermark")
	optionalPassword := ""

	embeddedImg, err := EmbedSteganoWatermark(baseImg, watermarkData, optionalPassword)

	// Check for errors
	if err != nil {
		t.Fatalf("EmbedSteganoWatermark returned an error: %v", err)
	}

	// Check if a new image is returned
	if embeddedImg == nil {
		t.Fatal("EmbedSteganoWatermark returned a nil image.")
	}

	// Check if the returned image is different from the base image
	if embeddedImg == baseImg {
		t.Fatal("EmbedSteganoWatermark returned the same image instance, expected a new one.")
	}

	// Basic check: dimensions should be the same
	if embeddedImg.Bounds() != baseImg.Bounds() {
		t.Errorf("Embedded image bounds %v different from base image bounds %v", embeddedImg.Bounds(), baseImg.Bounds())
	}

	// More specific check (optional):
	// Check if at least one pixel is different (steganography should alter some pixels)
	// This is a simplistic check; a full steganography decode would be needed for true verification.
	changed := false
	for x := 0; x < baseImg.Bounds().Dx(); x++ {
		for y := 0; y < baseImg.Bounds().Dy(); y++ {
			if baseImg.At(x, y) != embeddedImg.At(x, y) {
				changed = true
				break
			}
		}
		if changed {
			break
		}
	}
	if !changed {
		t.Error("Embedded image is identical to the base image; expected some pixel changes.")
	}

	// Test with password
	optionalPasswordWithPass := "password123"
	embeddedImgWithPass, errWithPass := EmbedSteganoWatermark(baseImg, watermarkData, optionalPasswordWithPass)
	if errWithPass != nil {
		t.Fatalf("EmbedSteganoWatermark with password returned an error: %v", errWithPass)
	}
	if embeddedImgWithPass == nil {
		t.Fatal("EmbedSteganoWatermark with password returned a nil image.")
	}
	if embeddedImgWithPass == baseImg {
		t.Fatal("EmbedSteganoWatermark with password returned the same image instance, expected a new one.")
	}
	if embeddedImgWithPass.Bounds() != baseImg.Bounds() {
		t.Errorf("Embedded image with password bounds %v different from base image bounds %v", embeddedImgWithPass.Bounds(), baseImg.Bounds())
	}
	changedWithPass := false
	for x := 0; x < baseImg.Bounds().Dx(); x++ {
		for y := 0; y < baseImg.Bounds().Dy(); y++ {
			if baseImg.At(x, y) != embeddedImgWithPass.At(x, y) {
				changedWithPass = true
				break
			}
		}
		if changedWithPass {
			break
		}
	}
	if !changedWithPass {
		t.Error("Embedded image with password is identical to the base image; expected some pixel changes.")
	}

	// Test with data too large for the image
	// GetImageCapacity for a 100x50 image with MaxBitDepth (8 bits) should be 100*50*8 bits / 8 = 5000 bytes
	// However, Reed-Solomon codes add overhead. The stegano library itself will determine the true capacity.
	// For this test, we'll create data that is definitely too large.
	// A 100x50 image has 5000 pixels. Each can store up to 3 bits (RGB) if MaxBitDepth (1) is used for stegano.LSB.
	// Or if MaxBitDepth is 8, it's 1 bit per channel per pixel.
	// The library's GetImageCapacity should provide the actual usable bytes.
	// Let's try with a very large byte array.
	largeWatermarkData := make([]byte, 100*50*4) // Significantly larger than any reasonable capacity
	_, errLarge := EmbedSteganoWatermark(baseImg, largeWatermarkData, "")
	if errLarge == nil {
		t.Error("EmbedSteganoWatermark did not return an error for data exceeding image capacity.")
	} else {
		t.Logf("Correctly received error for large data: %v", errLarge)
	}

}

func TestExtractSteganoWatermark(t *testing.T) {
	// Create a simple dummy image for testing.
	baseImg := image.NewRGBA(image.Rect(0, 0, 200, 100)) // Increased size for more capacity
	for x := 0; x < 200; x++ {
		for y := 0; y < 100; y++ {
			baseImg.Set(x, y, color.RGBA{R: uint8(x % 255), G: uint8(y % 255), B: uint8((x + y) % 255), A: 255})
		}
	}

	originalWatermarkData := []byte("This is a secret message!")
	correctPassword := "supersecretpassword"

	// --- Test Case 1: Embed-Extract cycle (no password) ---
	embeddedImgNoPass, err := EmbedSteganoWatermark(baseImg, originalWatermarkData, "")
	if err != nil {
		t.Fatalf("Test Case 1: Embed (no password) failed: %v", err)
	}

	extractedDataNoPass, err := ExtractSteganoWatermark(embeddedImgNoPass, "")
	if err != nil {
		t.Fatalf("Test Case 1: Extract (no password) failed: %v", err)
	}
	if string(extractedDataNoPass) != string(originalWatermarkData) {
		t.Errorf("Test Case 1: Extracted data does not match original. Got: %s, Want: %s", string(extractedDataNoPass), string(originalWatermarkData))
	}
	t.Logf("Test Case 1: Successfully embedded and extracted data without password.")

	// --- Test Case 2: Embed-Extract cycle (with password) ---
	embeddedImgWithPass, err := EmbedSteganoWatermark(baseImg, originalWatermarkData, correctPassword)
	if err != nil {
		t.Fatalf("Test Case 2: Embed (with password) failed: %v", err)
	}

	extractedDataWithPass, err := ExtractSteganoWatermark(embeddedImgWithPass, correctPassword)
	if err != nil {
		t.Fatalf("Test Case 2: Extract (with correct password) failed: %v", err)
	}
	if string(extractedDataWithPass) != string(originalWatermarkData) {
		t.Errorf("Test Case 2: Extracted data does not match original. Got: %s, Want: %s", string(extractedDataWithPass), string(originalWatermarkData))
	}
	t.Logf("Test Case 2: Successfully embedded and extracted data with password.")

	// --- Test Case 3: Attempt extraction from an image without a watermark ---
	// Using the original baseImg which has no watermark
	_, err = ExtractSteganoWatermark(baseImg, "")
	if err == nil {
		t.Errorf("Test Case 3: Expected error when extracting from non-watermarked image, got nil")
	} else {
		t.Logf("Test Case 3: Correctly received error when extracting from non-watermarked image: %v", err)
		// Check for a more specific error if possible, e.g., contains "no watermark data found" or "failed to extract data"
		// This depends on the error messages from the stegano library and our function.
		// For now, any error is accepted.
	}

	// --- Test Case 4: Attempt extraction with a wrong password ---
	// embeddedImgWithPass was created in Test Case 2 with correctPassword
	_, err = ExtractSteganoWatermark(embeddedImgWithPass, "wrongpassword123")
	if err == nil {
		t.Errorf("Test Case 4: Expected error when extracting with wrong password, got nil")
	} else {
		// We expect an error related to decryption failure.
		// The error message from stegano.DecryptData might be "cipher: message authentication failed" or similar.
		t.Logf("Test Case 4: Correctly received error when extracting with wrong password: %v", err)
		if !(strings.Contains(err.Error(), "failed to decrypt watermark data") || strings.Contains(err.Error(), "cipher: message authentication failed")) {
			t.Logf("Test Case 4: Warning - error message was '%s', expected something about decryption failure.", err.Error())
		}
	}

	// --- Test Case 5: Attempt extraction with empty password when one is required ---
	// embeddedImgWithPass was created in Test Case 2 with correctPassword
	extractedDataWrongPass, err := ExtractSteganoWatermark(embeddedImgWithPass, "")
	if err == nil {
		t.Errorf("Test Case 5: Expected error when extracting with empty password (was embedded with password), got nil")
	} else {
		t.Logf("Test Case 5: Correctly received error when extracting with empty password (was embedded with password): %v", err)
		// The extracted data here will be the encrypted data. We should not be able to use it directly.
		if string(extractedDataWrongPass) == string(originalWatermarkData) {
			t.Errorf("Test Case 5: Extracted data with empty password matched original, but it should be encrypted.")
		}
	}
}
