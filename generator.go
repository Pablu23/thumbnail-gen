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

func GetThumbnail(path string, intervalSeconds int, maxThumbnails int, enableFilter bool) ([][]byte, error) {
	var filters []TimeFilter
	if enableFilter {
		f, err := GetFilter(path)
		if err != nil {
			return nil, err
		}
		filters = f
	}

	fps, err := GetFramerate(path)
	if err != nil {
		return nil, err
	}

	length, err := GetVideoLength(path)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	currFrame := 0
	framesExtracted := 0

	var out [][]byte
	if maxThumbnails > 0 {
		out = make([][]byte, maxThumbnails)
	} else {
		out = make([][]byte, 0)
	}

	for {
		time := float64(currFrame) / fps
		if (maxThumbnails > 0 && framesExtracted >= maxThumbnails) || time >= length {
			break
		}

		currFrame += int(fps) * intervalSeconds
		if enableFilter && FrameLiesWithinFilter(time, filters) {
			continue
		}

		err := GetImage(buf, path, int(time), "png")
		if err != nil {
			return nil, err
		}

		b := bytes.Clone(buf.Bytes())[0:buf.Len()]

		if maxThumbnails > 0 {
			out[framesExtracted] = b
		} else {
			out = append(out, b)
		}

		framesExtracted += 1
		buf.Reset()
	}
	return out, nil
}

func FrameLiesWithinFilter(time float64, filters []TimeFilter) bool {
	for _, filter := range filters {
		if time >= filter.Start && time <= filter.End {
			return true
		}

	}
	return false
}

func GetImage(buf *bytes.Buffer, path string, timestamp int, format string) error {
	var t time.Time
	t = t.Add(time.Duration(timestamp) * time.Second)
	cmd := exec.Command("ffmpeg", "-ss", t.Format("15:04:05"), "-i", path, "-vframes", "1", "-c:v", format, "-f", "image2pipe", "-")
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
