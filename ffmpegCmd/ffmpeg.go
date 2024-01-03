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

// StartRtmpStream starts the RTMP stream using FFmpeg
func StartRtmpStream() error {
	cmd := exec.Command("ffmpeg",
		"-f", "v4l2",
		"-i", "/dev/video0",
		"-c:v", "h264",
		"-preset", "ultrafast",
		"-tune", "zerolatency",
		"-b:v", "900k",
		"-c:a", "aac",
		"-strict", "experimental",
		"-f", "fvlv",
		"rtmp://192.168.0.221:1935/live")

	stopCh := make(chan struct{})

	cmd.Stderr = os.Stderr

	go func() {
		err := cmd.Run()
		if err != nil {
			log.Printf("FFmpeg process terminated with error: %v\n", err)
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					// Handle the exit status if needed
					err.Error()

					log.Printf("Exit status: %d\n", status.ExitStatus())
				}
				close(stopCh) // Close the channel if there's an error
				return
			}
		}

		close(stopCh)
	}()

	ffmpegCmd = cmd // save the command so we can terminate it later
	signalCh := make(chan os.Signal, 1)

	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	select {
	case <-stopCh:
		return nil
	case <-signalCh:
		stopFFmpeg()
		return errors.New("FFmpeg process timed out")
	}
}

func StopRtmpStream() error {
	if ffmpegCmd != nil && ffmpegCmd.Process != nil {
		log.Println("Stopping FFmpeg process...")
		err := ffmpegCmd.Process.Signal(os.Interrupt)
		if err != nil {
			log.Println("Error stopping FFmpeg process:", err)
			return err
		}
		return nil
	}

	// If no FFmpeg process is running, return an error
	return errors.New("No active RTMP stream to stop")
}

// stopFFmpeg forcefully terminates the FFmpeg process
func stopFFmpeg() {
	// Wait for a few seconds to give FFmpeg a chance to gracefully terminate
	time.Sleep(5 * time.Second)

	// If the FFmpeg process is still running, forcefully terminate it
	if ffmpegCmd != nil && ffmpegCmd.Process != nil {
		log.Println("Forcefully terminating FFmpeg process...")
		err := ffmpegCmd.Process.Kill()
		if err != nil {
			log.Println("Error forcefully terminating FFmpeg process:", err)
		}
	}
}
