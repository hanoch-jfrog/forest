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
	GetConfig(ctx context.Context) (*model.Config, error)

	// Sets the log file name to use when querying the remote service for log data.
	SetLogFileName(logFileName string)

	// Sets the refresh rate interval between each log request.
	SetLogsRefreshRate(logsRefreshRate time.Duration)

	// Writes a single log data snapshot from the remote service into the passed io.Writer.
	// The configured node id and log file name are used.
	// Any error during read or write is returned.
	CatLog(ctx context.Context, output io.Writer) error

	// Writes continuous log data snapshots from the remote service into the passed io.Writer,
	// on an interval set by the LogsRefreshRate, defaulting to 1 second.
	// The configured node id and log file name are used.
	// Any errors during read or write is returned.
	// NOTE: this call blocks until cancellation of the passed context.Context.
	TailLog(ctx context.Context, output io.Writer) error
}
