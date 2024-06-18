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
	formatFlag        = flag.String("format", "png", "Output format")
	segmentsFlag      = flag.Bool("segments", false, "If active uses interval for how many segments are supposed to be done")
	widthFlag         = flag.Int("width", -1, "Set the width of the picture to scale to, -1 = max")
)

func main() {
	flag.Parse()

	if *pathFlag != "" {
		// thumbnails, num, err := thumbnailgen.GetThumbnail(*pathFlag, *intervalFlag, *maxThumbnailsFlag, *blackFilterFlag)
		opts := thumbnailgen.NewDefaultOptions()
		thumbnails, num, err := opts.Apply(func(o *thumbnailgen.Options) {
			o.EnableFilter = *blackFilterFlag
			o.Interval = *intervalFlag
			o.Format = *formatFlag
			o.UseSegments = *segmentsFlag
			o.Segments = *intervalFlag
			o.Scale = fmt.Sprintf("%d:-1", *widthFlag)
		}).GetThumbnail(*pathFlag)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Extracted %d thumbnails\n", num)

		name := filepath.Base(*pathFlag)
		for i, thumbnail := range thumbnails[:num] {
			err := os.WriteFile(fmt.Sprintf("%s-%d.%s", name, i, *formatFlag), thumbnail, 0600)
			if err != nil {
				panic(err)
			}
		}
	}
}
