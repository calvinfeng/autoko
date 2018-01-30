package main

import (
	"fmt"
	"image"
	"os"
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
		fmt.Printf("Rectangle %v\n", img.Bounds())
		CreateGaussianBlurImage("maps", name, img)
		// CreateEdgeDetectionImage("maps", name, img)
	}
}
