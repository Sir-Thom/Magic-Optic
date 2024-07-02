package main

import (
	"Magic-optic/api"
	"os/exec"
)

func main() {
	checkCommand("ffmpeg", "-version", "FFmpeg is not installed please install it")
	checkCommand("/usr/bin/v4l2-ctl", "--version", "v4l2-ctl is not installed please install it")

	api.Main()
}
func checkCommand(name string, args ...string) {
	cmd := exec.Command(name, args[:len(args)-1]...)
	if err := cmd.Run(); err != nil {
		panic(args[len(args)-1])
	}
}
