package strategy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hanoch-jfrog/forest/client/livelog/constants"
	"github.com/hanoch-jfrog/forest/client/livelog/model"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"io"
)

const (
	artifactoryNodesEndpoint = "api/system/nodes"
)

func NewArtifactoryLiveLogStrategy(rt artifactory.ArtifactoryServicesManager) *artifactoryLiveLogStrategy {
	return &artifactoryLiveLogStrategy{
		rt: rt,
	}
}

type artifactoryLiveLogStrategy struct {
	rt artifactory.ArtifactoryServicesManager
}

func (s *artifactoryLiveLogStrategy) GetServiceNodes(_ context.Context) (*model.ServiceNodes, error) {
	resBody, err := s.sendArtifactoryGet(artifactoryNodesEndpoint, "")
	if err != nil {
		return nil, err
	}

	serviceNodes := model.ServiceNodes{}
	err = json.Unmarshal(resBody, &serviceNodes)
	return &serviceNodes, err
}

func (s *artifactoryLiveLogStrategy) GetConfig(_ context.Context, nodeId string) (model.Config, error) {
	if nodeId == "" {
		return model.Config{}, fmt.Errorf("node id must be set")
	}

	resBody, err := s.sendArtifactoryGet(constants.ConfigEndpoint, nodeId)
	if err != nil {
		return model.Config{}, err
	}

	logConfig := model.Config{}
	err = json.Unmarshal(resBody, &logConfig)
	return logConfig, err
}

func (s *artifactoryLiveLogStrategy) GetLiveLog(_ context.Context, nodeId, logFileName string, lastPageMarker int64) (logReader io.Reader, newPageMarker int64, err error) {
	endpoint := fmt.Sprintf("%s?$file_size=%d&id=%s", constants.DataEndpoint, lastPageMarker, logFileName)
	resBody, err := s.sendArtifactoryGet(endpoint, nodeId)
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

func (s *artifactoryLiveLogStrategy) sendArtifactoryGet(endpoint, nodeId string) ([]byte, error) {
	client := s.rt.Client()
	httpClientDetails := (*client.ArtDetails).CreateHttpClientDetails()
	if nodeId != "" {
		httpClientDetails.Headers[constants.NodeIdHeader] = nodeId
	}

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
