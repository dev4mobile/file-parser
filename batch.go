package pdfparser

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"sync"

	"github.com/disintegration/imaging"
)

type ImageJob struct {
	InputPath     string
	OutputPath    string
	WatermarkType string // "visibleText", "visibleImage", "invisible"

	// Options for visible text watermark
	Text     string
	FontPath string
	FontSize float64
	HexColor string

	// Options for visible image watermark
	WatermarkImagePath string // Path to the watermark image file

	// Common for visible watermarks
	Opacity  float64
	Position string
	Margin   int

	// Options for invisible watermark
	InvisibleData string

	// Options for steganography blind watermark
	SteganoData     string
	SteganoPassword string
}

// worker processes image jobs from the jobs channel.
func worker(id int, jobs <-chan ImageJob, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		fmt.Printf("Worker %d: Starting job for %s\n", id, job.InputPath)

		baseImage, err := imaging.Open(job.InputPath)
		if err != nil {
			fmt.Printf("Worker %d: Error opening image %s: %v\n", id, job.InputPath, err)
			// Consider collecting this error if ProcessImagesConcurrently needs to return them
			continue
		}

		var outputImage image.Image
		var outputBuffer *bytes.Buffer // For invisible watermark

		switch job.WatermarkType {
		case "visibleText":
			// Ensure FontPath is valid or handle error appropriately
			if job.FontPath == "" {
				// This is a temporary measure for testing concurrency as per subtask note.
				// In a real scenario, font handling would be more robust.
				fmt.Printf("Worker %d: Warning - FontPath is empty for %s. Text watermarking might fail or use default.\n", id, job.InputPath)
				// Potentially use a system default if possible or skip/error out.
				// For now, let ApplyTextWatermark handle it (it might error if font is crucial and not found by gg).
			}
			outputImage, err = ApplyTextWatermark(baseImage, job.Text, job.FontPath, job.FontSize, job.HexColor, job.Opacity, job.Position, job.Margin)
			if err != nil {
				fmt.Printf("Worker %d: Error applying text watermark to %s: %v\n", id, job.InputPath, err)
				continue
			}
		case "visibleImage":
			outputImage, err = ApplyImageWatermark(baseImage, job.WatermarkImagePath, job.Opacity, job.Position, job.Margin)
			if err != nil {
				fmt.Printf("Worker %d: Error applying image watermark to %s: %v\n", id, job.InputPath, err)
				continue
			}
		case "invisible":
			outputBuffer, err = EmbedInvisibleWatermark(baseImage, job.InvisibleData)
			if err != nil {
				fmt.Printf("Worker %d: Error embedding invisible watermark in %s: %v\n", id, job.InputPath, err)
				continue
			}
		case "steganoBlind":
			if job.SteganoData == "" {
				fmt.Printf("Worker %d: Error for %s - SteganoData cannot be empty for steganoBlind type\n", id, job.InputPath)
				continue
			}
			watermarkBytes := []byte(job.SteganoData)

			outputImage, err = EmbedSteganoWatermark(baseImage, watermarkBytes, job.SteganoPassword)
			if err != nil {
				fmt.Printf("Worker %d: Error applying stegano blind watermark to %s: %v\n", id, job.InputPath, err)
				continue
			}
			// outputImage is now populated and will be saved by existing logic.
		default:
			fmt.Printf("Worker %d: Unknown watermark type '%s' for job %s\n", id, job.WatermarkType, job.InputPath)
			continue
		}

		// Save the processed image
		if job.WatermarkType == "invisible" {
			if outputBuffer != nil {
				err = os.WriteFile(job.OutputPath, outputBuffer.Bytes(), 0644)
				if err != nil {
					fmt.Printf("Worker %d: Error writing invisible watermarked image %s: %v\n", id, job.OutputPath, err)
					continue
				}
			}
		} else {
			if outputImage != nil {
				err = imaging.Save(outputImage, job.OutputPath)
				if err != nil {
					fmt.Printf("Worker %d: Error saving watermarked image %s: %v\n", id, job.OutputPath, err)
					continue
				}
			}
		}
		fmt.Printf("Worker %d: Successfully processed %s to %s\n", id, job.InputPath, job.OutputPath)
	}
	fmt.Printf("Worker %d: Finished\n", id)
}

// ProcessImagesConcurrently processes a list of image jobs using a worker pool.
func ProcessImagesConcurrently(jobsToProcess []ImageJob, numWorkers int) []error {
	jobsChan := make(chan ImageJob, len(jobsToProcess))
	var wg sync.WaitGroup
	var processingErrors []error // Currently, workers log errors; this slice is for future enhancement.

	fmt.Printf("Starting %d workers for %d jobs.\n", numWorkers, len(jobsToProcess))

	// Start workers
	for i := 1; i <= numWorkers; i++ {
		go worker(i, jobsChan, &wg)
	}

	// Send jobs to workers
	for _, job := range jobsToProcess {
		wg.Add(1)
		jobsChan <- job
	}
	close(jobsChan) // Close channel to signal workers that no more jobs will be sent

	wg.Wait() // Wait for all workers to complete

	fmt.Println("All workers have finished.")
	return processingErrors // Returning an empty slice for now as per current worker implementation
}
