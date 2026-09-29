package main

import (
	"fmt"
	"os/exec"
)

func processVideoForFastStart(filePath string) (string, error) {
	outputFile := filePath + ".processing"

	args := []string{"-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputFile}
	cmd := exec.Command("ffmpeg", args...)

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("cmd run error: %v\n", err)
	}

	return outputFile, nil
}
