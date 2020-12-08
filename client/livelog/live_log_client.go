package livelog

import (
	"context"
	"github.com/hanoch-jfrog/forest/client/livelog/model"
	"io"
	"time"
)

type Client interface {
	// Queries and returns the available nodes from the remote service.
	GetServiceNodes(ctx context.Context) (*model.ServiceNodes, error)

	// Sets the node id to use when querying the remote service for log data.
	SetNodeId(nodeId string)

	// Queries and returns the livelog configuration from the remote service, based on the set node id.
	GetConfig(ctx context.Context) (model.Config, error)

	// Sets the log file name to use when querying the remote service for log data.
	SetLogFileName(logFileName string)

	// Sets the refresh rate interval between each log request.
	SetLogsRefreshRate(logsRefreshRate time.Duration)

	// Returns an io.Reader, which can be used to read a single log data snapshot from the remote service.
	// The configured node id and log file name are used.
	// Any errors are transmitted on the returned reader.
	CatLog(ctx context.Context) io.Reader

	// Returns an io.Reader, which can be used to read continuous log data snapshots from the remote service,
	// on an interval set by the LogsRefreshRate, defaulting to 1 second.
	// The configured node id and log file name are used.
	// Any errors are transmitted on the returned reader.
	// Cancellation of the passed context.Context will terminate the underlying goroutine.
	TailLog(ctx context.Context) io.Reader
}
