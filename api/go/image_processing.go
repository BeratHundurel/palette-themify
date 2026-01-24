package main

import (
	"image"
	"image/color"
	"runtime"
	"sync"
)

func processImageWithShepardsMethod(
	img image.Image,
	paletteRGBAs []color.RGBA,
	luminosity float64,
	nearest int,
	power float64,
	maxDistanceSq float64,
) *image.RGBA {
	bounds := img.Bounds()
	height := bounds.Dy()
	out := image.NewRGBA(bounds)

	numWorkers := max(min(runtime.GOMAXPROCS(0), height), 1)
	rowsPerWorker := (height + numWorkers - 1) / numWorkers

	var wg sync.WaitGroup
	for workerID := range numWorkers {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			startY := bounds.Min.Y + id*rowsPerWorker
			endY := min(startY+rowsPerWorker, bounds.Max.Y)

			for y := startY; y < endY; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					originalRGBA := toRGBA(img.At(x, y))

					if originalRGBA.A == 0 {
						out.Set(x, y, color.Transparent)
						continue
					}

					if maxDistanceSq > 0 {
						if nearestDistanceSquared(originalRGBA, paletteRGBAs) > maxDistanceSq {
							out.Set(x, y, originalRGBA)
							continue
						}
					}

					adjusted := applyLuminosity(originalRGBA, luminosity)
					finalColor := shepardsMethodColor(adjusted, paletteRGBAs, nearest, power)
					out.Set(x, y, finalColor)
				}
			}
		}(workerID)
	}

	wg.Wait()
	return out
}
