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
)

func main() {
	flag.Parse()

	if *pathFlag != "" {
		thumbnails, err := thumbnailgen.GetThumbnail(*pathFlag, *intervalFlag, *maxThumbnailsFlag)
		if err != nil {
			panic(err)
		}

		name := filepath.Base(*pathFlag)
		for i, thumbnail := range thumbnails {
			err := os.WriteFile(fmt.Sprintf("%s-%d.png", name, i), thumbnail, 0600)
			if err != nil {
				panic(err)
			}
		}
	}
}
