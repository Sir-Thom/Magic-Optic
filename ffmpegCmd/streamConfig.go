package ffmpegCmd

import "os/exec"

type RtmpConfig struct {
	DevicePath string `json:"devicePath"`
	VideoCodec string `json:"videoCodec"`
	Preset     string `json:"preset"`
	Tune       string `json:"tune"`
	Bitrate    string `json:"bitrate"`
	AudioCodec string `json:"audioCodec"`
	StreamUrl  string `json:"streamUrl"`
}

func (config RtmpConfig) Command() *exec.Cmd {
	return exec.Command("ffmpeg",
		"-f", "v4l2",
		"-i", config.DevicePath,
		"-c:v", config.VideoCodec,
		"-preset", config.Preset,
		"-tune", config.Tune,
		"-b:v", config.Bitrate,
		"-c:a", config.AudioCodec,
		"-strict", "experimental",
		"-f", "flv",
		"-flvflags", "no_duration_filesize",
		"-max_muxing_queue_size", "9999",
		config.StreamUrl)
}

type RtspConfig struct {
	DevicePath string `json:"devicePath"`
	VideoCodec string `json:"videoCodec"`
	Preset     string `json:"preset"`
	Tune       string `json:"tune"`
	Bitrate    string `json:"bitrate"`
	AudioCodec string `json:"audioCodec"`
	StreamUrl  string `json:"streamUrl"`
}

func (config RtspConfig) Command() *exec.Cmd {
	return exec.Command("ffmpeg",
		"-f", "v4l2",
		"-i", config.DevicePath,
		"-c:v", config.VideoCodec,
		"-preset", config.Preset,
		"-tune", config.Tune,
		"-b:v", config.Bitrate,
		"-c:a", config.AudioCodec,
		"-strict", "experimental",
		"-f", "rtsp",
		"-flvflags", "no_duration_filesize",
		"-max_muxing_queue_size", "9999",
		config.StreamUrl)
}
