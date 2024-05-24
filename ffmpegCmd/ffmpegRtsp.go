package ffmpegCmd

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"syscall"
)

type RtspStreamConfig struct {
	device     string
	devicePath string
	videoCodec string
	preset     string
	tune       string
	bitrate    string
	audioCodec string
	rtmpUrl    string
}

// TODO make a ffmpeg builder that can be used to build custom  ffmpeg commands
// TODO make a ffmpeg  for rtsp streams
// TODO make a ffmpeg  for srt streams (optional)
// TODO make a way to change some of these settings from the web interface
// Function to start RTSP stream
func StartRtspStream(config RtspConfig) (<-chan struct{}, error) {
	cmd := exec.Command("ffmpeg",
		"-f", "v4l2",
		"-i", config.DevicePath,
		"-c:v", config.VideoCodec,
		"-preset", config.Preset,
		"-tune", config.Tune,
		"-b:v", config.Bitrate,
		"-c:a", config.AudioCodec,
		"-strict", "experimental",
		"-f", "rtsp",
		config.StreamUrl)

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

// Function to stop RTSP stream
func StopRtspStream() error {
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

// Function to check if RTSP stream is running
func IsRtspStreamRunning() bool {
	return ffmpegCmd != nil && ffmpegCmd.Process != nil
}
