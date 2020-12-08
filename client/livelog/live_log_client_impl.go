package livelog

import (
	"context"
	"fmt"
	"github.com/hanoch-jfrog/forest/client/livelog/model"
	"github.com/hanoch-jfrog/forest/client/livelog/strategy"
	"io"
	"time"
)

const (
	defaultRequestTimeout    = 15 * time.Second
	defaultLogRequestTimeout = time.Minute
	defaultLogsRefreshRate   = time.Second
)

type client struct {
	strategy        strategy.LiveLogStrategy
	nodeId          string
	logFileName     string
	logsRefreshRate time.Duration
}

func NewClient(strategy strategy.LiveLogStrategy) *client {
	return &client{
		strategy:        strategy,
		logsRefreshRate: defaultLogsRefreshRate,
	}
}

func (s *client) GetServiceNodes(ctx context.Context) (*model.ServiceNodes, error) {
	timeoutCtx, cancelTimeout := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancelTimeout()
	return s.strategy.GetServiceNodes(timeoutCtx)
}

func (s *client) GetConfig(ctx context.Context) (model.Config, error) {
	if s.nodeId == "" {
		return model.Config{}, fmt.Errorf("node id must be set")
	}

	timeoutCtx, cancelTimeout := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancelTimeout()
	return s.strategy.GetConfig(timeoutCtx, s.nodeId)
}

func (s *client) SetNodeId(nodeId string) {
	s.nodeId = nodeId
}

func (s *client) SetLogFileName(logFileName string) {
	s.logFileName = logFileName
}

func (s *client) SetLogsRefreshRate(logsRefreshRate time.Duration) {
	s.logsRefreshRate = logsRefreshRate
}

func (s *client) CatLog(ctx context.Context) io.Reader {
	logReader, _, err := s.doCatLog(ctx, 0)
	if err != nil {
		return newErrReader(err)
	}
	return logReader
}

func (s *client) TailLog(ctx context.Context) io.Reader {
	pageMarker := int64(0)
	readerChan := make(chan io.Reader)
	errChan := make(chan error)

	go func() {
		for {
			select {
			case <-ctx.Done():
				errChan <- io.EOF
				close(readerChan)
				close(errChan)
				return
			case <-time.After(s.logsRefreshRate):
				var logReader io.Reader
				var err error

				logReader, pageMarker, err = s.doCatLog(ctx, pageMarker)
				if err != nil {
					errChan <- err
					close(readerChan)
					close(errChan)
					return
				}
				readerChan <- logReader
			}
		}
	}()
	return newBlockingReader(readerChan, errChan)
}

func (s *client) doCatLog(ctx context.Context, lastPageMarker int64) (logReader io.Reader, newPageMarker int64, err error) {
	if s.nodeId == "" {
		return nil, 0, fmt.Errorf("node id must be set")
	}
	if s.logFileName == "" {
		return nil, 0, fmt.Errorf("log file name must be set")
	}

	timeoutCtx, cancelTimeout := context.WithTimeout(ctx, defaultLogRequestTimeout)
	defer cancelTimeout()
	return s.strategy.GetLiveLog(timeoutCtx, s.nodeId, s.logFileName, lastPageMarker)
}
