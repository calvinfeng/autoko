package main

import (
	"fmt"
	"image"
	"os"
	"time"
)

func main() {
	reader, err := os.Open("maps/fetch_office.png")
	if err != nil {
		fmt.Println("Error", err)
	}

	defer reader.Close()

	img, name, decodeErr := image.Decode(reader)
	if decodeErr != nil {
		fmt.Println("Decoding has error", decodeErr)
	} else {
		fmt.Printf("Successfully decoded %s\n", name)
		fmt.Printf("Loaded image has the following dimension: %v\n", img.Bounds())

		start := time.Now()
		// CreateGaussianBlurImage("maps", name, img)
		// CreateEdgeDetectionImage("maps", name, img)
		// CreateFloodFillImage("maps", name, img)
		// CreateAutoKeepoutImage("maps", name, img)
		points := []*Point{
			{false, 0, 0},
			{false, 3, 1},
			{false, 1, 1},
			{false, 1, 3},
			{false, 3, 4},
			{false, 0, 4},
			{false, 4, 3},
			{false, -1, 2},
		}

		LabelHullVertices(points)
		for _, point := range points {
			if point.IsHullVertex {
				fmt.Println(point)
			}
		}

		end := time.Now()

		fmt.Printf("Algorithm took %v to complete \n", end.Sub(start))
	}
}
