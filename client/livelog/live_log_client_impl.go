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

func (s *client) GetServiceNodeIds(ctx context.Context) ([]string, error) {
	timeoutCtx, cancelTimeout := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancelTimeout()
	endpoint := s.httpStrategy.NodesEndpoint()
	resBody, err := s.httpStrategy.SendGet(timeoutCtx, endpoint, "")
	if err != nil {
		return nil, err
	}

	serviceNodes := model.ServiceNodes{}
	if err = json.Unmarshal(resBody, &serviceNodes); err != nil {
		return nil, err
	}
	if len(serviceNodes.Nodes) == 0 {
		return nil, fmt.Errorf("no node ids found")
	}
	nodeIds := make([]string, len(serviceNodes.Nodes))
	for idx, serviceNode := range serviceNodes.Nodes {
		nodeIds[idx] = serviceNode.NodeId
	}
	return nodeIds, err
}

func (s *client) GetConfig(ctx context.Context) (*model.Config, error) {
	if s.nodeId == "" {
		return nil, fmt.Errorf("node id must be set")
	}

	timeoutCtx, cancelTimeout := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancelTimeout()
	resBody, err := s.httpStrategy.SendGet(timeoutCtx, constants.ConfigEndpoint, s.nodeId)
	if err != nil {
		return nil, err
	}

	logConfig := model.Config{}
	err = json.Unmarshal(resBody, &logConfig)
	if err != nil {
		return nil, err
	}
	if len(logConfig.LogFileNames) == 0 {
		return nil, fmt.Errorf("no log file names were found")
	}
	return &logConfig, nil
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
	if err != nil {
		return err
	}
	_, err = io.Copy(output, logReader)
	return err
}

func (s *client) TailLog(ctx context.Context, output io.Writer) error {
	pageMarker := int64(0)
	curLogRefreshRate := time.Duration(0)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(curLogRefreshRate):
			if curLogRefreshRate == 0 {
				curLogRefreshRate = s.logsRefreshRate
			}
			var logReader io.Reader
			var err error

			logReader, pageMarker, err = s.doCatLog(ctx, pageMarker)
			if err != nil {
				return err
			}
			_, err = io.Copy(output, logReader)
			if err != nil {
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
