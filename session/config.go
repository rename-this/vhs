package session

import "time"

// Config is a VHS config.
type Config struct {
	FlowDuration       time.Duration
	InputDrainDuration time.Duration
	ShutdownDuration   time.Duration

	Addr            string
	CaptureResponse bool
	Middleware      string
	TCPTimeout      time.Duration
	HTTPTimeout     time.Duration

	PrometheusAddr string

	GCSBucketName string
	GCSObjectName string

	BufferOutput bool
}
