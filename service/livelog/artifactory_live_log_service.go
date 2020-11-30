package livelog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/httpclient"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"io"
	"time"
)

const (
	artifactoryNodesEndpoint = "/api/system/nodes"
)

type artifactoryService struct {
	rt          artifactory.ArtifactoryServicesManager
	nodeId      string
	logFileName string
}

func NewArtifactoryService(rtServiceManager artifactory.ArtifactoryServicesManager) *artifactoryService {
	return &artifactoryService{
		rt: rtServiceManager,
	}
}

func (s *artifactoryService) GetServiceNodes(_ context.Context) ([]string, error) {
	client := s.rt.Client()
	httpClientDetails := (*client.ArtDetails).CreateHttpClientDetails()

	resBody, err := sendArtifactoryGet(client, httpClientDetails, artifactoryNodesEndpoint)
	if err != nil {
		return nil, err
	}

	logConfig := Config{}
	err = json.Unmarshal(resBody, &logConfig)
	return []string{string(resBody)}, err
}

func (s *artifactoryService) GetConfig(_ context.Context) (Config, error) {
	if s.nodeId == "" {
		return Config{}, fmt.Errorf("node id must be set")
	}

	client := s.rt.Client()
	httpClientDetails := (*client.ArtDetails).CreateHttpClientDetails()
	httpClientDetails.Headers[NodeIdHeader] = s.nodeId

	resBody, err := sendArtifactoryGet(client, httpClientDetails, configEndpoint)
	if err != nil {
		return Config{}, err
	}

	logConfig := Config{}
	err = json.Unmarshal(resBody, &logConfig)
	return logConfig, err
}

func (s *artifactoryService) SetNodeId(nodeId string) {
	s.nodeId = nodeId
}

func (s *artifactoryService) SetLogFileName(logFileName string) {
	s.logFileName = logFileName
}

func (s *artifactoryService) CatLog(ctx context.Context) io.Reader {
	logReader, _, err := s.doCatLog(ctx, 0)
	if err != nil {
		return newErrReadCloser(err)
	}
	return logReader
}

func (s *artifactoryService) TailLog(ctx context.Context, pollingInterval time.Duration) io.Reader {
	pageMarker := int64(0)
	buf := bytes.Buffer{}

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(pollingInterval):
			var logReader io.Reader
			var err error

			logReader, pageMarker, err = s.doCatLog(ctx, pageMarker)
			if err != nil {
				// TODO: log error (logger / stdout / stderr ?)
				if _, err := io.Copy(&buf, newErrReadCloser(err)); err != nil {
					// TODO: log error (logger / stdout / stderr ?)
					return
				}
			} else if _, err := io.Copy(&buf, logReader); err != nil {
				// TODO: log error (logger / stdout / stderr ?)
				return
			}
		}
	}()
	return &buf
}

func (s *artifactoryService) doCatLog(_ context.Context, lastPageMarker int64) (logReader io.Reader, newPageMarker int64, err error) {
	if s.nodeId == "" {
		return nil, 0, fmt.Errorf("node id must be set")
	}
	if s.logFileName == "" {
		return nil, 0, fmt.Errorf("log file name must be set")
	}

	client := s.rt.Client()
	httpClientDetails := (*client.ArtDetails).CreateHttpClientDetails()
	httpClientDetails.Headers[NodeIdHeader] = s.nodeId

	endpoint := fmt.Sprintf("%s?$file_size=%d&id=%s", dataEndpoint, lastPageMarker, s.logFileName)
	resBody, err := sendArtifactoryGet(client, httpClientDetails, endpoint)
	if err != nil {
		return nil, 0, err
	}

	logData := Data{}
	if err := json.Unmarshal(resBody, &logData); err != nil {
		return nil, 0, err
	}

	logDataBuf := bytes.NewBufferString(logData.Content)
	return logDataBuf, logData.PageMarker, nil
}

func sendArtifactoryGet(client *httpclient.ArtifactoryHttpClient,
	httpClientDetails httputils.HttpClientDetails, endpoint string) ([]byte, error) {

	baseUrl := (*client.ArtDetails).GetUrl()
	res, resBody, _, err := client.SendGet(baseUrl+endpoint, true, &httpClientDetails)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected response; status code: %d, message: %s", res.StatusCode, resBody)
	}
	return resBody, nil
}
