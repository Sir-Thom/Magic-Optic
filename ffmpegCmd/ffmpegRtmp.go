package ffmpegCmd

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"syscall"
)

var ffmpegCmd *exec.Cmd

func StartRtmpStream(config RtmpConfig) (<-chan struct{}, error) {
	var currentConfig = config
	cmd := exec.Command("ffmpeg",
		"-f", "v4l2",
		"-i", currentConfig.DevicePath,
		"-c:v", currentConfig.VideoCodec,
		"-preset", currentConfig.Preset,
		"-tune", currentConfig.Tune,
		"-b:v", currentConfig.Bitrate,
		"-c:a", currentConfig.AudioCodec,
		"-strict", "experimental",
		"-f", "flv",
		currentConfig.StreamUrl)
	currentConfig.StreamType = "rtmp"
	stopCh := make(chan struct{})

	cmd.Stderr = os.Stderr

	go func() {
		defer close(stopCh)

		err := cmd.Run()
		if err != nil {
			log.Println("FFmpeg process exited", err)
		}
	}()

	ffmpegCmd = cmd
	return stopCh, nil
}

func StopRtmpStream() error {
	if ffmpegCmd == nil || ffmpegCmd.Process == nil {
		return errors.New("FFmpeg process is not running")
	}

	if ffmpegCmd.ProcessState != nil && ffmpegCmd.ProcessState.Exited() {
		return errors.New("FFmpeg process has already exited")
	}

	log.Println("Trying to gracefully terminate FFmpeg process...")
	err := ffmpegCmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		log.Println("Error sending SIGTERM signal:", err)
		return err
	}

	err = ffmpegCmd.Wait()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// Process exited with a non-zero status
			status := exitErr.Sys().(syscall.WaitStatus)
			log.Printf("FFmpeg process exited with non-zero status: %d\n", status.ExitStatus())
		} else {
			log.Println("Error waiting for FFmpeg process to exit:", err)
			return err
		}
	}

	return nil
}
func IsStreamRunning() bool {
	return ffmpegCmd != nil && ffmpegCmd.Process != nil
}
