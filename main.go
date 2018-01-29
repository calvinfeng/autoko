package main

import (
	"fmt"
	"image"
	"os"
	"time"
)

func main() {
	reader, err := os.Open("maps/caterpillar.png")
	if err != nil {
		fmt.Println("Error", err)
	}

	defer reader.Close()

	img, name, decodeErr := image.Decode(reader)
	if decodeErr != nil {
		fmt.Println("Decoding has error", decodeErr)
	} else {
		fmt.Printf("Successfully decoded %s\n", name)
		fmt.Printf("Rectangle %v\n", img.Bounds())

		start := time.Now()
		CreateGaussianBlurImage("maps", name, img)
		// CreateEdgeDetectionImage("maps", name, img)
		end := time.Now()

		fmt.Printf("Algorithm took %v time\n", end.Sub(start))
	}
}
