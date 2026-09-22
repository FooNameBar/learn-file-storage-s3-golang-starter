package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

func getVideoAspectRatio(filepath string) (string, error) {
	args := []string{"-v", "error", "-print_format", "json", "-show_streams", filepath}
	cmd := exec.Command("ffprobe", args...)

	stdOut := make([]byte, 0)
	// cmd.Stdout = bytes.NewBuffer(stdOut)

	// err := cmd.Run()
	stdOut, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("cmd run error: %v\n", err)
	}

	result := ffprobeStream{}
	err = json.Unmarshal(stdOut, &result)
	if err != nil {
		return "", fmt.Errorf("json unmarshaling error: %v\njson: %s\ncmd: %s\n", err, string(stdOut), cmd.String())
	}

	height, width := result.Streams[0].Height, result.Streams[0].Width
	// aspect :=  result.Streams[0].DisplayAspectRatio
	aspect := "other"
	if height > width {
		aspect = "portrait"
	} else if width > height {
		aspect = "landscape"
	}

	return aspect, nil
}

type ffprobeStream struct {
	Streams []struct {
		Height             int    `json:"height"`
		Width              int    `json:"width"`
		DisplayAspectRatio string `json:"display_aspect_ratio"`
	}
}
