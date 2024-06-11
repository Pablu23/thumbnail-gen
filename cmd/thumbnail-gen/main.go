package main

import (
	"flag"
	"fmt"

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
    _, err := thumbnailgen.GetFramerate(*pathFlag)
		if err != nil {
			fmt.Printf("Framerate: %s\n", err)
		}

		_, err = thumbnailgen.GetVideoLength(*pathFlag)
		if err != nil {
			fmt.Printf("Video Length: %s\n", err)
		}

    _, err = thumbnailgen.GetFilter(*pathFlag)
    if err != nil {
			fmt.Printf("Filter: %s\n", err)
		}
		// thumbnails, err := thumbnailgen.GetThumbnail(*pathFlag, *intervalFlag, *maxThumbnailsFlag)
		// if err != nil {
		// 	panic(err)
		// }
		//
		// name := filepath.Base(*pathFlag)
		// for i, thumbnail := range thumbnails {
		// 	err := os.WriteFile(fmt.Sprintf("%s-%d.png", name, i), thumbnail, 0600)
		// 	if err != nil {
		// 		panic(err)
		// 	}
		// }
	}
}
