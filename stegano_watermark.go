package pdfparser

import (
	"fmt"
	"image"

	"github.com/scott-mescudi/stegano"
)

// EmbedSteganoWatermark embeds watermarkData into baseImage using steganography.
// It returns a new image with the watermark embedded.
// If optionalPassword is provided, the watermarkData is encrypted before embedding.
func EmbedSteganoWatermark(baseImage image.Image, watermarkData []byte, optionalPassword string) (image.Image, error) {
	embedder := stegano.NewEmbedHandler()

	// Check capacity
	// Note: stegano.GetImageCapacity returns capacity in bytes.
	// The Encode method in stegano compresses data before checking capacity.
	// EmbedDataIntoImage does not compress. So, we check capacity against raw data size.
	// Reed-Solomon codes will add overhead, which is part of what GetImageCapacity accounts for.
	capacity, err := stegano.GetImageCapacity(baseImage, stegano.MaxBitDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to get image capacity: %w", err)
	}

	dataToEmbed := watermarkData
	// Encrypt data if password is provided
	if optionalPassword != "" {
		encryptedData, err := stegano.EncryptData(watermarkData, optionalPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt watermark data: %w", err)
		}
		dataToEmbed = encryptedData
	}

	// Check if data (potentially encrypted) is too large for the image
	// This check should ideally account for any overhead stegano itself adds,
	// beyond just the raw data size. GetImageCapacity is supposed to give the *usable* capacity.
	if uint64(len(dataToEmbed)) > capacity {
		return nil, fmt.Errorf("watermark data size (%d bytes) exceeds image capacity (%d bytes)", len(dataToEmbed), capacity)
	}

	// Embed data into image
	// According to documentation, EmbedDataIntoImage uses LSB and does not compress.
	embeddedImage, err := embedder.EmbedDataIntoImage(baseImage, dataToEmbed, stegano.LSB)
	if err != nil {
		return nil, fmt.Errorf("failed to embed data into image: %w", err)
	}

	return embeddedImage, nil
}

// ExtractSteganoWatermark extracts watermarkData from a watermarkedImage.
// If optionalPassword was used during embedding, it must be provided for decryption.
func ExtractSteganoWatermark(watermarkedImage image.Image, optionalPassword string) ([]byte, error) {
	extractor := stegano.NewExtractHandler()

	// Extract data from image
	// Assuming LSB was used for embedding, as it's common and matches EmbedDataIntoImage.
	extractedData, err := extractor.ExtractDataFromImage(watermarkedImage, stegano.LSB)
	if err != nil {
		// This error might occur if no watermark is found or if the image format is unsupported.
		return nil, fmt.Errorf("failed to extract data from image: %w", err)
	}

	if len(extractedData) == 0 {
		// Consider if empty data is an error or a valid "no watermark found" scenario.
		// The stegano library might return an error above if no EOD marker is found.
		// If it returns empty data without error, this means no data was embedded.
		return nil, fmt.Errorf("no watermark data found in image")
	}

	// Decrypt data if password is provided
	if optionalPassword != "" {
		decryptedData, err := stegano.DecryptData(extractedData, optionalPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt watermark data: %w", err)
		}
		return decryptedData, nil
	}

	return extractedData, nil
}
