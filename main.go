package main

import (
	"fmt"
	"image"
	"os"
	"time"

	"github.com/calvinfeng/autoko/annotate"
)

func main() {
	mapName := "microsoft"

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
		// annotate.CreateFloodFillImage("maps", mapName, img)
		// annotate.CreateGaussianBlurImage("maps", mapName, img)
		// annotate.CreateEdgeDetectionImage("maps", mapName, img)
		// annotate.CreateClusteringImage("maps", mapName, img)
		annotate.CreateConvexHullImage("maps", mapName, img)
		end := time.Now()

		fmt.Printf("Algorithm took %v to complete \n", end.Sub(start))
	}
}
