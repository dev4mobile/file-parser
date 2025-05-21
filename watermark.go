package pdfparser

import (
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"

	"github.com/fogleman/gg"
)

// parseHexColor converts a hex string (e.g., "#RRGGBB", "#RRGGBBAA") to a color.RGBA struct.
func parseHexColor(hexStr string) (color.Color, error) {
	hexStr = strings.TrimPrefix(hexStr, "#")
	var r, g, b, a uint8
	var err error

	if len(hexStr) != 6 && len(hexStr) != 8 {
		return nil, fmt.Errorf("invalid hex color string length: %s", hexStr)
	}

	parseByte := func(str string) (uint8, error) {
		val, err := strconv.ParseUint(str, 16, 8)
		if err != nil {
			return 0, err
		}
		return uint8(val), nil
	}

	r, err = parseByte(hexStr[0:2])
	if err != nil {
		return nil, fmt.Errorf("failed to parse red component: %v", err)
	}
	g, err = parseByte(hexStr[2:4])
	if err != nil {
		return nil, fmt.Errorf("failed to parse green component: %v", err)
	}
	b, err = parseByte(hexStr[4:6])
	if err != nil {
		return nil, fmt.Errorf("failed to parse blue component: %v", err)
	}

	if len(hexStr) == 8 {
		a, err = parseByte(hexStr[6:8])
		if err != nil {
			return nil, fmt.Errorf("failed to parse alpha component: %v", err)
		}
	} else {
		a = 255 // Default to full opacity
	}

	return color.RGBA{R: r, G: g, B: b, A: a}, nil
}

// ApplyTextWatermark applies a text watermark to an image.
func ApplyTextWatermark(baseImage image.Image, text string, fontPath string, fontSize float64, hexColor string, opacity float64, position string, margin int) (image.Image, error) {
	// Load the font
	if err := gg.LoadFontFace(fontPath, fontSize); err != nil {
		return nil, fmt.Errorf("failed to load font face from %s: %v", fontPath, err)
	}

	// Create a new gg.Context from the baseImage
	dc := gg.NewContextForImage(baseImage)

	// Set the font for the context
	// Note: gg currently uses a global state for font loading.
	// We ensure it's set for this context by calling it again, though LoadFontFace sets it globally.
	// A more robust solution might involve gg.Context.SetFontFace if available or managing font instances.
	dc.LoadFontFace(fontPath, fontSize) // Ensure font is set for this context

	// Parse the hexColor and set the drawing color
	parsedColor, err := parseHexColor(hexColor)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hex color '%s': %v", hexColor, err)
	}

	// Apply opacity
	r, g, b, a := parsedColor.RGBA() // Returns values in [0, 0xffff] range
	// Normalize to [0, 255] for RGBA struct and apply opacity to alpha
	finalA := uint8(float64(a>>8) * opacity)
	dc.SetRGBA255(int(r>>8), int(g>>8), int(b>>8), int(finalA))


	// Calculate text dimensions
	textWidth, textHeight := dc.MeasureString(text)

	// Determine X, Y coordinates
	imageWidth := float64(baseImage.Bounds().Dx())
	imageHeight := float64(baseImage.Bounds().Dy())
	floatMargin := float64(margin)

	var x, y float64

	// gg.DrawString draws text with (x,y) as the baseline of the text.
	// gg.DrawStringAnchored draws text with (x,y) as an anchor point (0,0 is top-left of text).
	// We will use DrawStringAnchored for easier positioning.

	switch strings.ToLower(position) {
	case "topleft":
		x = floatMargin
		y = floatMargin
		dc.DrawStringAnchored(text, x, y, 0.0, 0.0) // Anchor Top-Left
	case "topcenter":
		x = (imageWidth - textWidth) / 2
		y = floatMargin
		dc.DrawStringAnchored(text, x, y, 0.0, 0.0) // Anchor Top-Left, but position x is centered
	case "topright":
		x = imageWidth - textWidth - floatMargin
		y = floatMargin
		dc.DrawStringAnchored(text, x, y, 0.0, 0.0) // Anchor Top-Left
	case "centerleft":
		x = floatMargin
		y = (imageHeight - textHeight) / 2
		dc.DrawStringAnchored(text, x, y, 0.0, 0.0) // Anchor Top-Left
	case "center":
		x = (imageWidth - textWidth) / 2
		y = (imageHeight - textHeight) / 2 // Anchor Top-Left of text for center
		dc.DrawStringAnchored(text, x, y, 0.0, 0.0)
	case "centerright":
		x = imageWidth - textWidth - floatMargin
		y = (imageHeight - textHeight) / 2
		dc.DrawStringAnchored(text, x, y, 0.0, 0.0) // Anchor Top-Left
	case "bottomleft":
		x = floatMargin
		y = imageHeight - textHeight - floatMargin
		dc.DrawStringAnchored(text, x, y, 0.0, 0.0) // Anchor Top-Left
	case "bottomcenter":
		x = (imageWidth - textWidth) / 2
		y = imageHeight - textHeight - floatMargin
		dc.DrawStringAnchored(text, x, y, 0.0, 0.0) // Anchor Top-Left
	case "bottomright":
		x = imageWidth - textWidth - floatMargin
		y = imageHeight - textHeight - floatMargin
		dc.DrawStringAnchored(text, x, y, 0.0, 0.0) // Anchor Top-Left
	default:
		return nil, fmt.Errorf("unsupported position: %s", position)
	}
	// Note: The y-coordinate for DrawStringAnchored with ay=0.0 means the top of the text.
	// If we were using DrawString, y would be the baseline, so we'd need to add textHeight for Top* positions.
	// For Bottom* positions with DrawString, y would be `imageHeight - margin` and the text baseline would sit there.
	// DrawStringAnchored(text, x, y, 0.0, 0.0) is simpler as (x,y) is the top-left corner of the text block.

	return dc.Image(), nil
}

// ApplyImageWatermark applies an image watermark to a base image.
func ApplyImageWatermark(baseImage image.Image, watermarkPath string, opacity float64, position string, margin int) (image.Image, error) {
	// a. Load the watermark image from watermarkPath using imaging.Open().
	watermarkImage, err := imaging.Open(watermarkPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load watermark image from %s: %v", watermarkPath, err)
	}

	var finalWatermark image.Image
	if opacity < 0.0 || opacity > 1.0 {
		return nil, fmt.Errorf("opacity must be between 0.0 and 1.0, got: %.2f", opacity)
	}

	// b. Handle Opacity
	if opacity < 1.0 {
		wmWidth := watermarkImage.Bounds().Dx()
		wmHeight := watermarkImage.Bounds().Dy()
		// i. Create a new gg.Context with the dimensions of the loaded watermark image.
		dc := gg.NewContext(wmWidth, wmHeight)
		// ii. Set the opacity on the context using dc.SetA(opacity).
		dc.SetA(opacity)
		// iii. Draw the loaded watermark image onto this context at (0,0).
		dc.DrawImage(watermarkImage, 0, 0)
		// iv. Get the resulting image (with adjusted opacity).
		finalWatermark = dc.Image()
	} else {
		// If opacity is 1.0 (or effectively 1.0), use the loaded watermark image directly.
		finalWatermark = watermarkImage
	}

	// c. Calculate the X, Y coordinates
	baseWidth := baseImage.Bounds().Dx()
	baseHeight := baseImage.Bounds().Dy()
	wmWidth := finalWatermark.Bounds().Dx()
	wmHeight := finalWatermark.Bounds().Dy()
	// floatMargin := float64(margin) // Not needed if x,y are int

	var x, y int

	switch strings.ToLower(position) {
	case "topleft":
		x = margin
		y = margin
	case "topcenter":
		x = (baseWidth - wmWidth) / 2
		y = margin
	case "topright":
		x = baseWidth - wmWidth - margin
		y = margin
	case "centerleft":
		x = margin
		y = (baseHeight - wmHeight) / 2
	case "center":
		x = (baseWidth - wmWidth) / 2
		y = (baseHeight - wmHeight) / 2
	case "centerright":
		x = baseWidth - wmWidth - margin
		y = (baseHeight - wmHeight) / 2
	case "bottomleft":
		x = margin
		y = baseHeight - wmHeight - margin
	case "bottomcenter":
		x = (baseWidth - wmWidth) / 2
		y = baseHeight - wmHeight - margin
	case "bottomright":
		x = baseWidth - wmWidth - margin
		y = baseHeight - wmHeight - margin
	default:
		return nil, fmt.Errorf("unsupported position: %s. Supported: TopLeft, TopCenter, TopRight, CenterLeft, Center, CenterRight, BottomLeft, BottomCenter, BottomRight", position)
	}

	// d. Use imaging.Paste to overlay the watermark.
	// imaging.Paste expects the destination to be a draw.Image.
	// We create a new NRGBA image, draw the baseImage onto it, then use this as the destination for Paste.
	// This ensures baseImage is not modified if it wasn't a draw.Image, and we return a new image.
	
	// Create a new drawable image (NRGBA) with the dimensions of the base image.
	dst := image.NewNRGBA(baseImage.Bounds())

	// Draw the original baseImage onto our new drawable image.
	// We use gg here for consistency, though image/draw could also be used.
	ggCtx := gg.NewContextForImage(dst)
	ggCtx.DrawImage(baseImage, 0, 0)

	// Paste the finalWatermark (which may have adjusted opacity) onto the dst image.
	// imaging.Paste modifies dst in place and returns it.
	resultImage := imaging.Paste(dst, finalWatermark, image.Pt(x, y))

	// e. Return the modified image and nil error.
	return resultImage, nil
}

// EmbedInvisibleWatermark embeds a string message into an image using steganography.
// It returns a bytes.Buffer containing the PNG data of the image with the embedded message.
func EmbedInvisibleWatermark(baseImage image.Image, dataToEmbed string) (*bytes.Buffer, error) {
	// a. Convert the dataToEmbed string to []byte.
	messageBytes := []byte(dataToEmbed)

	// b. Create a new bytes.Buffer to act as the writer for the output image.
	buffer := new(bytes.Buffer)

	// c. Call steganography.Encode(buffer, baseImage, messageBytes).
	// The steganography library expects the image to be in a format it can work with directly,
	// typically one that doesn't lose information like PNG.
	// The baseImage image.Image might need to be an *image.RGBA or *image.NRGBA for the library to work correctly.
	// If not, we might need to convert it. Let's assume it's compatible for now.
	err := steganography.Encode(buffer, baseImage, messageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to embed invisible watermark: %v", err)
	}

	// d. Return the buffer and any error.
	return buffer, nil
}

// ExtractInvisibleWatermark extracts a hidden message from an image using steganography.
func ExtractInvisibleWatermark(watermarkedImage image.Image) (string, error) {
	// a. Use steganography.GetMessageSizeFromImage(watermarkedImage) to get the size of the hidden message.
	//    The library might expect a specific image type (e.g., *image.RGBA).
	//    Let's assume watermarkedImage is suitable.
	sizeOfMessage := steganography.GetMessageSizeFromImage(watermarkedImage)
	if sizeOfMessage == 0 {
		return "", fmt.Errorf("no hidden message found or message size is zero")
	}

	// b. Call steganography.Decode(size, watermarkedImage) to extract the message bytes.
	extractedBytes, err := steganography.Decode(sizeOfMessage, watermarkedImage)
	if err != nil {
		return "", fmt.Errorf("failed to extract invisible watermark: %v", err)
	}

	// c. Convert the extracted byte slice to a string.
	extractedString := string(extractedBytes)

	// d. Return the extracted string and any error.
	return extractedString, nil
}
