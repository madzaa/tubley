package util

import (
	"bytes"
	"encoding/json"
	"os/exec"
)

type Ratio int

const (
	Landscape Ratio = iota
	Portrait
	Other
)

func (r Ratio) String() string {
	return [...]string{"landscape", "portrait", "other"}[r]
}

func GetVideoAspectRatio(filepath string) (string, error) {
	metadata := VideoMetadata{}
	getMetadata := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams",
		filepath)
	var b bytes.Buffer

	stdOut := getMetadata.Stdout
	_, err := stdOut.Write(b.AvailableBuffer())
	if err != nil {
		return "", err
	}
	err = getMetadata.Run()
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(b.Bytes(), &metadata)
	if err != nil {
		return "", err
	}

	for _, stream := range metadata.Streams {
		ratio := stream.Width / stream.Height
		switch ratio {
		case 0:
			return Portrait.String(), err
		case 1:
			return Landscape.String(), err
		default:
			return Other.String(), err
		}
	}
	return "", err
}
