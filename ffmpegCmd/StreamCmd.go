package ffmpegCmd

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/google/uuid"
)

type StreamConfig interface {
	Command() *exec.Cmd
}

type Stream struct {
	id     string
	cmd    *exec.Cmd
	stopCh chan struct{}
}

type StreamManager struct {
	streams map[string]*Stream
	mu      sync.Mutex
}

func NewStreamManager() *StreamManager {
	return &StreamManager{
		streams: make(map[string]*Stream),
	}
}

func (sm *StreamManager) StartStream(config StreamConfig) (string, <-chan struct{}, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	id := uuid.New().String()
	if _, exists := sm.streams[id]; exists {
		return "", nil, errors.New("stream with this ID already exists")
	}

	cmd := config.Command()
	stopCh := make(chan struct{})

	cmd.Stderr = os.Stderr

	go func() {
		defer close(stopCh)
		err := cmd.Run()
		if err != nil {
			log.Printf("FFmpeg process exited: %v\n", err)
		}
	}()

	sm.streams[id] = &Stream{
		id:     id,
		cmd:    cmd,
		stopCh: stopCh,
	}
	return id, stopCh, nil
}

func (sm *StreamManager) StopStream(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	stream, exists := sm.streams[id]

	if stream.cmd.ProcessState != nil && stream.cmd.ProcessState.Exited() {
		delete(sm.streams, id)
		return errors.New("FFmpeg process has already exited")
	}

	if !exists {
		return errors.New("stream not found")
	}

	log.Println("Trying to gracefully terminate FFmpeg process...")
	err := stream.cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		log.Printf("Error sending SIGTERM signal: %v\n", err)
		return err
	}

	err = stream.cmd.Wait()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			status := exitErr.Sys().(syscall.WaitStatus)
			log.Printf("FFmpeg process exited with non-zero status: %d\n", status.ExitStatus())
		} else {
			log.Printf("Error waiting for FFmpeg process to exit: %v\n", err)
			return err
		}
	}

	delete(sm.streams, id)
	return nil
}

func (sm *StreamManager) IsStreamRunning(id string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	stream, exists := sm.streams[id]
	if !exists {
		return false
	}
	return stream.cmd.Process != nil
}

func (sm *StreamManager) CheckAllStream() map[string]*Stream {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	return sm.streams

}
