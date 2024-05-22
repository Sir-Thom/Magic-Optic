package ffmpegCmd

type HlsConfig struct {
	Device     string `json:"device"`
	DevicePath string `json:"devicePath"`
	VideoCodec string `json:"videoCodec"`
	Preset     string `json:"preset"`
	Tune       string `json:"tune"`
	Bitrate    string `json:"bitrate"`
	AudioCodec string `json:"audioCodec"`
	StreamUrl  string `json:"streamUrl"`
	StreamType string `json:"streamType"`
}

type RtmpConfig struct {
	Device     string `json:"device"`
	DevicePath string `json:"devicePath"`
	VideoCodec string `json:"videoCodec"`
	Preset     string `json:"preset"`
	Tune       string `json:"tune"`
	Bitrate    string `json:"bitrate"`
	AudioCodec string `json:"audioCodec"`
	StreamUrl  string `json:"streamUrl"`
	StreamType string `json:"streamType"`
}
