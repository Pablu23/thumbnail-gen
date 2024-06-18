package thumbnailgen

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

type TimeFilter struct {
	Start float64
	End   float64
}

type Options struct {
	Interval      int
	Segments      int
	UseSegments   bool
	Format        string
	Scale         string
	EnableFilter  bool
	MaxThumbnails int
}

func NewDefaultOptions() Options {
	return Options{
		Interval:      5,
		Segments:      0,
		UseSegments:   false,
		Format:        "webp",
		Scale:         "-1:-1",
		EnableFilter:  true,
		MaxThumbnails: 0,
	}
}

func (options *Options) Apply(opts ...func(*Options)) *Options {
	for _, opt := range opts {
		opt(options)
	}

	return options
}

func (opts *Options) GetThumbnail(path string) ([][]byte, int, error) {
	var filters []TimeFilter
	if opts.EnableFilter {
		f, err := GetFilter(path)
		if err != nil {
			return nil, 0, err
		}
		filters = f
	}

	length, err := GetVideoLength(path)
	if err != nil {
		return nil, 0, err
	}

	var interval int 
	if opts.UseSegments {
		interval = int(length / float64(opts.Segments))
	} else {
    interval = opts.Interval
	}

	return GetThumbnailUnderlying(path, opts.MaxThumbnails, filters, length, interval, opts.EnableFilter, opts.Format, opts.Scale)
}

func GetThumbnailUnderlying(path string, maxThumbnails int, filters []TimeFilter, length float64, intervalSeconds int, enableFilter bool, format string, scale string) ([][]byte, int, error) {
	buf := bytes.NewBuffer(nil)
	framesExtracted := 0

	var out [][]byte
	if maxThumbnails > 0 {
		out = make([][]byte, maxThumbnails)
	} else {
		out = make([][]byte, 0)
	}

	var time float64 = 1
	for {
		if (maxThumbnails > 0 && framesExtracted >= maxThumbnails) || time >= length {
			break
		}

		if enableFilter {
			if ok, next := FrameLiesWithinFilter(time, filters); ok {
				time = next + 1
				continue
			}
		}

		err := GetImage(buf, path, int(time), format, scale)
		if err != nil {
			return nil, 0, err
		}

		b := bytes.Clone(buf.Bytes())[0:buf.Len()]

		if maxThumbnails > 0 {
			out[framesExtracted] = b
		} else {
			out = append(out, b)
		}

		time += float64(intervalSeconds)
		framesExtracted += 1
		buf.Reset()
	}
	return out, framesExtracted, nil
}

// Return true, if time is inside filter, also returns next time, when not in filter anymore.
// If False it returns input time
func FrameLiesWithinFilter(time float64, filters []TimeFilter) (bool, float64) {
	for _, filter := range filters {
		if time >= filter.Start && time <= filter.End {
			return true, filter.End
		}

	}
	return false, time
}

func GetImage(buf *bytes.Buffer, path string, timestamp int, format string, scale string) error {
	var t time.Time
	t = t.Add(time.Duration(timestamp) * time.Second)
	cmd := exec.Command("ffmpeg", "-ss", t.Format("15:04:05"), "-i", path, "-vframes", "1", "-c:v", format, "-filter:v", fmt.Sprintf("scale=%s", scale), "-f", "image2pipe", "-")
	cmd.Stdout = buf
	err := cmd.Run()
	return err
}

func GetVideoLength(path string) (float64, error) {
	cmd := exec.Command("ffprobe", "-v", "quiet", "-select_streams", "v:0", "-show_entries", "stream=duration", "-of", "default=noprint_wrappers=1:nokey=1", path)
	buf := bytes.NewBuffer(nil)
	cmd.Stdout = buf
	err := cmd.Run()
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(strings.ReplaceAll(buf.String(), "\n", ""), 64)
}

func GetFilter(path string) ([]TimeFilter, error) {
	buf := bytes.NewBuffer(nil)
	cmd := exec.Command("ffprobe", "-f", "lavfi", "-i", fmt.Sprintf("movie=%s,blackdetect[out0]", path), "-show_entries", "tags=lavfi.black_start,lavfi.black_end", "-of", "default=nw=1", "-v", "quiet")
	cmd.Stdout = buf
	err := cmd.Run()
	// TODO: Replace with strings.Builder
	filterStr := buf.String()
	filters := strings.Split(filterStr, "\n")
	filters = slices.Compact(filters)
	i := 0
	reg := regexp.MustCompile(`([0-9]*\.?[0-9]+)`)
	blackFilters := make([]TimeFilter, len(filters)/2)
	for {
		if i >= len(filters) {
			break
		}
		start := reg.FindString(filters[i])
		var end string
		if i+1 >= len(filters) || filters[i+1] == "" {
			end = start
		} else {
			end = reg.FindString(filters[i+1])
		}

		if start == "" || end == "" {
			break
		}

		s, err := strconv.ParseFloat(start, 64)
		if err != nil {
			return nil, err
		}

		e, err := strconv.ParseFloat(end, 64)
		if err != nil {
			return nil, err
		}

		blackFilters[i/2] = TimeFilter{
			Start: s,
			End:   e,
		}
		i += 2
	}

	return blackFilters, err
}

func GetFramerate(path string) (float64, error) {
	cmd := exec.Command("ffprobe", "-v", "0", "-of", "csv=p=0", "-select_streams", "v:0", "-show_entries", "stream=r_frame_rate", path)
	buf := bytes.NewBuffer(nil)
	cmd.Stdout = buf
	err := cmd.Run()
	if err != nil {
		return 0, err
	}
	s := buf.String()
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.TrimSpace(s)
	ss := strings.Split(s, "/")
	f1, err := strconv.Atoi(ss[0])
	if err != nil {
		return 0, err
	}

	f2, err := strconv.Atoi(ss[1])
	if err != nil {
		return 0, err
	}
	return float64(f1) / float64(f2), nil
}
