package strategy

import "context"

type Http interface {
	// Returns the api path to use for the remote service's available nodes endpoint.
	NodesEndpoint() string
	// Performs a GET request to the remote service, using the passed endpoint.
	// if nodeId is not empty, it is appended as the X-JFrog-Node-Id header value.
	SendGet(ctx context.Context, endpoint, nodeId string) ([]byte, error)
}
