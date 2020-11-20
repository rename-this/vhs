package session

import "time"

// Config is a general config.
type Config struct {
	PrometheusAddr string

	GCSBucketName string
	GCSObjectName string


	Debug             bool
	DebugPackets      bool
	DebugHTTPMessages bool

	ProfilePathCPU    string
	ProfilePathMemory string
	ProfileHTTPAddr   string
}

// FlowConfig is a Flow config.
type FlowConfig struct {
	FlowDuration       time.Duration
	InputDrainDuration time.Duration
	ShutdownDuration   time.Duration

	Addr            string
	CaptureResponse bool
	Middleware      string
	TCPTimeout      time.Duration
	HTTPTimeout     time.Duration

	BufferOutput bool
}
