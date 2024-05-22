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

func StartHlsStream(config HlsConfig) (<-chan struct{}, error) {

	cmd := exec.Command("ffmpeg",
		"-f", "v4l2",
		"-i", config.DevicePath,
		"-c:v", config.VideoCodec,
		"-preset", config.Preset,
		"-tune", config.Tune,
		"-b:v", config.Bitrate,
		"-c:a", config.AudioCodec,
		"-strict", "experimental",
		"-f", "hls",
		"-hls_time", "10", // Segment duration
		"-hls_playlist_type", "event", // HLS playlist type
		config.StreamUrl)

	stopCh := make(chan struct{})

	cmd.Stderr = os.Stderr

	go func() {
		err := cmd.Run()
		if err != nil {
			log.Println("FFmpeg process exited", err)
		}
		time.Sleep(1 * time.Second)
		ffmpegCmd = cmd
		close(stopCh)
	}()

	ffmpegCmd = cmd
	signalCh := make(chan os.Signal, 1)

	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	select {
	case <-stopCh:
		return stopCh, nil

	case <-signalCh:
		err := StopStream()
		if err != nil {
			return stopCh, err
		}
		close(signalCh)

	}
	<-stopCh
	return stopCh, nil
}
func StopStream() error {
	if ffmpegCmd == nil || ffmpegCmd.Process == nil || ffmpegCmd.ProcessState != nil && !ffmpegCmd.ProcessState.Exited() {
		return errors.New("FFmpeg process is not running")
	}

	log.Println("Trying to gracefully terminate FFmpeg process...")
	err := ffmpegCmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		log.Println("Error sending SIGTERM signal:", err)
		return err
	}

	if !ffmpegCmd.ProcessState.Exited() {
		log.Println("Forcefully terminating FFmpeg process...")
		err := ffmpegCmd.Process.Kill()
		if err != nil {
			log.Println("Error forcefully terminating FFmpeg process:", err)
		}
	}

	return nil
}

func IsStreamRunning() bool {
	return ffmpegCmd != nil && ffmpegCmd.Process != nil
}
