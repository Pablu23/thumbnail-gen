package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	thumbnailgen "github.com/pablu23/thumbnail-gen"
)

var (
	pathFlag          = flag.String("path", "", "path to video")
	intervalFlag      = flag.Int("interval", 20, "Interval in seconds between thumbnails")
	maxThumbnailsFlag = flag.Int("max", 0, "Max Thumbnails, default as much as possible with interval")
	blackFilterFlag   = flag.Bool("filter", true, "Try to filter out black frames, might be an expensive operation")
)

func main() {
	flag.Parse()

	if *pathFlag != "" {
		// thumbnails, num, err := thumbnailgen.GetThumbnail(*pathFlag, *intervalFlag, *maxThumbnailsFlag, *blackFilterFlag)
		thumbnails, num, err := thumbnailgen.GetThumbnailSegments(*pathFlag, 12, *blackFilterFlag)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Extracted %d thumbnails\n", num)

		name := filepath.Base(*pathFlag)
		for i, thumbnail := range thumbnails[:num] {
			err := os.WriteFile(fmt.Sprintf("%s-%d.png", name, i), thumbnail, 0600)
			if err != nil {
				panic(err)
			}
		}
	}
}
