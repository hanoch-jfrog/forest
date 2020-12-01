package livelog

import (
	"context"
	"io"
	"time"
)

const (
	configEndpoint = "/api/v1/system/logs/config"
	dataEndpoint   = "/api/v1/system/logs/data"
	NodeIdHeader   = "X-JFrog-Node-Id"
)

// Defines a LiveLog interface, configured for
type Service interface {
	// Queries and returns the available nodes from the remote service.
	GetServiceNodes(ctx context.Context) ([]string, error)

	// Sets the node id to use when querying the remote service for log data.
	SetNodeId(nodeId string)

	// Queries and returns the livelog configuration from the remote service.
	GetConfig(ctx context.Context) (Config, error)

	// Sets the log file name to use when querying the remote service for log data.
	SetLogFileName(logFileName string)

	// Returns an io.Reader, which can be used to read a single log data snapshot from the remote service.
	// The configured log file name is used, defaulting to console.log.
	// Any errors are transmitted on the returned reader.
	CatLog(ctx context.Context) io.Reader

	// Returns an io.Reader, which can be used to read consecutive log data snapshots from the remote service,
	// on an interval set by the passed pollingInterval.
	// The configured log file name is used, defaulting to console.log.
	// Any errors are transmitted on the returned reader.
	// Cancellation of the passed context.Context will terminate the underlying goroutine.
	TailLog(ctx context.Context, pollingInterval time.Duration) io.Reader
}

type Config struct {
	LogFileNames      []string `json:"logs,omitempty"`
	RefreshRateMillis int64    `json:"refresh_rate_millis,omitempty"`
}

type Data struct {
	LogFileModified int64  `json:"last_update_modified,omitempty"`
	Timestamp       int64  `json:"last_update_label,omitempty"`
	Content         string `json:"log_content,omitempty"`
	PageMarker      int64  `json:"file_size,omitempty"`
}
