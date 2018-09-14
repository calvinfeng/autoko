package main

import (
	"autokeepout"
	"fmt"
	"image"
	"os"
	"time"
)

func main() {
	mapName := "caterpillar"

	reader, err := os.Open(fmt.Sprintf("maps/%s.png", mapName))
	if err != nil {
		fmt.Println("Error", err)
	}

	defer reader.Close()

	img, _, decodeErr := image.Decode(reader)
	if decodeErr != nil {
		fmt.Println("Decoding has error", decodeErr)
	} else {
		fmt.Printf("Successfully decoded %s\n", mapName)
		fmt.Printf("Loaded image has the following dimension: %v\n", img.Bounds())

		start := time.Now()
		// CreateGaussianBlurImage("maps", name, img)
		// CreateEdgeDetectionImage("maps", name, img)
		// CreateFloodFillImage("maps", name, img)
		autokeepout.CreateAutoKeepoutImage("maps", mapName, img)
		end := time.Now()

		fmt.Printf("Algorithm took %v to complete \n", end.Sub(start))
	}
}
