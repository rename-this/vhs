package session

import "time"

// Config is a general config.
type Config struct {
	PrometheusAddr string

	Debug             bool
	DebugPackets      bool
	DebugHTTPMessages bool

	ProfilePathCPU    string
	ProfilePathMemory string
	ProfileHTTPAddr   string

	Plugin string
}

// FlowConfig is a Flow config.
type FlowConfig struct {
	SourceDuration     time.Duration
	InputDrainDuration time.Duration

	Addr            string
	AddrSink        string
	CaptureResponse bool
	Middleware      string
	HTTPTimeout     time.Duration

	TCPTimeout time.Duration

	BufferOutput bool

	GCSBucketName string
	GCSObjectName string

	InputFile string

	S3CompatEndpoint   string
	S3CompatAccessKey  string
	S3CompatSecretKey  string
	S3CompatToken      string
	S3CompatSecure     bool
	S3CompatBucketName string
	S3CompatObjectName string
}
