package main

import (
	"fmt"
	"log"
	"os"

	"github.com/vzvu3k6k/hyperpaper"
)

func handleErr(err error) {
	log.Fatal(err)
}

func main() {
	boundingBoxesFile, err := os.Open("C3-5.tsv")
	if err != nil {
		handleErr(err)
	}
	defer boundingBoxesFile.Close()

	boundingBoxes, err := hyperpaper.LoadBoundingBoxes(boundingBoxesFile)
	if err != nil {
		handleErr(err)
	}

	layoutFile, err := os.Open("C5-5-1.sorted.xml")
	if err != nil {
		handleErr(err)
	}
	defer layoutFile.Close()

	visibleRects, err := hyperpaper.BuildVisibleRects(layoutFile)
	if err != nil {
		handleErr(err)
	}

	for _, b := range boundingBoxes {
		if b.Page != 1 {
			continue
		}
		for _, r := range visibleRects {
			if hyperpaper.IsOverlapping(b.Rect, r) {
				fmt.Printf("visible: %s\n", b.Text)
				break
			}
		}
	}
}
