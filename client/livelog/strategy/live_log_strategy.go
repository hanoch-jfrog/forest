package strategy

import (
	"context"
	"github.com/hanoch-jfrog/forest/client/livelog/model"
	"io"
)

type LiveLogStrategy interface {
	// Queries and returns the available nodes from the remote service.
	GetServiceNodes(ctx context.Context) (*model.ServiceNodes, error)

	// Queries and returns the livelog configuration from remote service specific node.
	GetConfig(ctx context.Context, nodeId string) (model.Config, error)

	// Returns an io.Reader, which can be used to read a single log file data snapshot,
	// from the remote service specific node.
	// Any errors are transmitted on the returned reader.
	GetLiveLog(ctx context.Context, nodeId, logFileName string, lastPageMarker int64) (logReader io.Reader, newPageMarker int64, err error)
}
