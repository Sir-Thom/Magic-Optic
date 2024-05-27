package api

import (
	"fmt"
	"os/exec"
	"strings"
)

func getAudioCodecs() ([]string, error) {
	cmd := exec.Command("ffmpeg", "-codecs")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute FFmpeg: %v", err)
	}

	var audioCodecs []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {

		if strings.Contains(line, "A.") {
			audioCodecs = append(audioCodecs, strings.Fields(line)[1])
		}
	}
	audioCodecs = append(audioCodecs[:0], audioCodecs[1:]...)
	return audioCodecs, nil
}
