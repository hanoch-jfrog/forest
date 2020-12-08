package livelog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hanoch-jfrog/forest/client/livelog/constants"
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
	httpStrategy    strategy.Http
	nodeId          string
	logFileName     string
	logsRefreshRate time.Duration
}

func NewClient(strategy strategy.Http) *client {
	return &client{
		httpStrategy:    strategy,
		logsRefreshRate: defaultLogsRefreshRate,
	}
}

func (s *client) GetServiceNodes(ctx context.Context) (*model.ServiceNodes, error) {
	timeoutCtx, cancelTimeout := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancelTimeout()
	endpoint := s.httpStrategy.NodesEndpoint()
	resBody, err := s.httpStrategy.SendGet(timeoutCtx, endpoint, "")
	if err != nil {
		return nil, err
	}

	serviceNodes := model.ServiceNodes{}
	err = json.Unmarshal(resBody, &serviceNodes)
	return &serviceNodes, err
}

func (s *client) GetConfig(ctx context.Context) (model.Config, error) {
	if s.nodeId == "" {
		return model.Config{}, fmt.Errorf("node id must be set")
	}

	timeoutCtx, cancelTimeout := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancelTimeout()
	resBody, err := s.httpStrategy.SendGet(timeoutCtx, constants.ConfigEndpoint, s.nodeId)
	if err != nil {
		return model.Config{}, err
	}

	logConfig := model.Config{}
	err = json.Unmarshal(resBody, &logConfig)
	return logConfig, err
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

func (s *client) CatLog(ctx context.Context, output io.Writer) error {
	logReader, _, err := s.doCatLog(ctx, 0)
	_, err = io.Copy(output, logReader)
	return err
}

func (s *client) TailLog(ctx context.Context, output io.Writer) error {
	pageMarker := int64(0)

	catLogFunc := func() error {
		var logReader io.Reader
		var internalErr error

		logReader, pageMarker, internalErr = s.doCatLog(ctx, pageMarker)
		if internalErr != nil {
			return internalErr
		}
		_, internalErr = io.Copy(output, logReader)
		if internalErr != nil {
			return internalErr
		}
		return nil
	}
	if err := catLogFunc(); err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(s.logsRefreshRate):
			if err := catLogFunc(); err != nil {
				return err
			}
		}
	}
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
	endpoint := fmt.Sprintf("%s?$file_size=%d&id=%s", constants.DataEndpoint, lastPageMarker, s.logFileName)
	resBody, err := s.httpStrategy.SendGet(timeoutCtx, endpoint, s.nodeId)
	if err != nil {
		return nil, 0, err
	}

	logData := model.Data{}
	if err := json.Unmarshal(resBody, &logData); err != nil {
		return nil, 0, err
	}

	logDataBuf := bytes.NewBufferString(logData.Content)
	return logDataBuf, logData.PageMarker, nil
}
