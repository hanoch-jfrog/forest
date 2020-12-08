package strategy

import (
	"context"
	"fmt"
	"github.com/hanoch-jfrog/forest/client/livelog/constants"
	"github.com/jfrog/jfrog-client-go/artifactory"
)

const (
	artifactoryNodesEndpoint = "api/system/nodes"
)

func NewArtifactoryHttpStrategy(rt artifactory.ArtifactoryServicesManager) *artifactoryHttpStrategy {
	return &artifactoryHttpStrategy{
		rt: rt,
	}
}

type artifactoryHttpStrategy struct {
	rt artifactory.ArtifactoryServicesManager
}

func (s *artifactoryHttpStrategy) NodesEndpoint() string {
	return artifactoryNodesEndpoint
}

func (s *artifactoryHttpStrategy) SendGet(_ context.Context, endpoint, nodeId string) ([]byte, error) {
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
