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

type RtmpStreamConfig struct {
	Device     string `json:"device"`
	DevicePath string `json:"devicePath"`
	VideoCodec string `json:"videoCodec"`
	Preset     string `json:"preset"`
	Tune       string `json:"tune"`
	Bitrate    string `json:"bitrate"`
	AudioCodec string `json:"audioCodec"`
	RtmpUrl    string `json:"rtmpUrl"`
}

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

func IsRtmpStreamRunning() bool {
	return ffmpegCmd != nil && ffmpegCmd.Process != nil
}
