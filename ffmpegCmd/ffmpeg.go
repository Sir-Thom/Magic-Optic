package ffmpegCmd

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

var ffmpegCmd *exec.Cmd

// TODO make a ffmpeg builder that can be used to build custom  ffmpeg commands
// TODO make a ffmpeg  for rtsp streams
// TODO make a ffmpeg  for srt streams (optional)
// TODO make a way to change some of these settings from the web interface
// StartRtmpStream starts the RTMP stream using FFmpeg
func StartRtmpStream() (<-chan struct{}, error) {
	cmd := exec.Command("ffmpeg",
		"-f", "v4l2",
		"-i", "/dev/video0",
		"-c:v", "h264",
		"-preset", "ultrafast",
		"-tune", "zerolatency",
		"-b:v", "900k",
		"-c:a", "aac",
		"-strict", "experimental",
		"-f", "flv",
		"rtmp://192.168.0.221:1935/live")

	stopCh := make(chan struct{})

	cmd.Stderr = os.Stderr

	go func() {
		err := cmd.Run()
		if err != nil {
			log.Println("FFmpeg process exited", err)
		}
		// Introduce a delay before updating ffmpegCmd.Process
		time.Sleep(1 * time.Second)
		ffmpegCmd = cmd // Save the command so we can terminate it later
		close(stopCh)
	}()

	ffmpegCmd = cmd // save the command so we can terminate it later
	signalCh := make(chan os.Signal, 1)

	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	select {
	case <-stopCh:
		return stopCh, nil

	case <-signalCh:
		err := StopRtmpStream()
		if err != nil {
			return stopCh, err
		}
		close(signalCh)

	}
	<-stopCh
	return stopCh, nil
}

func StopRtmpStream() error {
	// If the FFmpeg process is not running, return immediately
	if ffmpegCmd == nil || ffmpegCmd.Process == nil || ffmpegCmd.ProcessState != nil && !ffmpegCmd.ProcessState.Exited() {
		return errors.New("FFmpeg process is not running")
	}

	// Try to gracefully terminate the FFmpeg process
	log.Println("Trying to gracefully terminate FFmpeg process...")
	err := ffmpegCmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		log.Println("Error sending SIGTERM signal:", err)
		return err
	}

	// If the FFmpeg process is still running, forcefully terminate it
	if !ffmpegCmd.ProcessState.Exited() {
		log.Println("Forcefully terminating FFmpeg process...")
		err := ffmpegCmd.Process.Kill()
		if err != nil {
			log.Println("Error forcefully terminating FFmpeg process:", err)
		}
	}

	return nil
}

func IsRtmpStreamRunning() bool {
	return ffmpegCmd != nil && ffmpegCmd.Process != nil
}
