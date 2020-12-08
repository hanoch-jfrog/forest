package commands

import (
	"fmt"
	rtcommands "github.com/jfrog/jfrog-cli-core/artifactory/commands"
	"github.com/jfrog/jfrog-cli-core/artifactory/utils"
	configutil "github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
)

func newArtifactoryServiceManager(serverId string) (artifactory.ArtifactoryServicesManager, error) {
	rtDetails, err := getRtDetails(serverId)
	if err != nil {
		return nil, err
	}
	return utils.CreateServiceManager(rtDetails, false)
}

func getRtDetails(serverId string) (*configutil.ArtifactoryDetails, error) {
	details, err := rtcommands.GetConfig(serverId, false)
	if err != nil {
		return nil, err
	}
	if details.Url == "" {
		return nil, fmt.Errorf("no server-id was found, or the server-id has no url")
	}
	details.Url = clientutils.AddTrailingSlashIfNeeded(details.Url)
	err = configutil.CreateInitialRefreshableTokensIfNeeded(details)
	if err != nil {
		return nil, err
	}
	return details, nil
}

func fetchAllServerIds() ([]string, error) {
	configs, err := configutil.GetAllArtifactoryConfigs()
	if err != nil {
		return nil, err
	}
	if len(configs) == 0 {
		return nil, fmt.Errorf("no CLI server IDs found")
	}
	serverIds := make([]string, len(configs))
	for idx, conf := range configs {
		serverIds[idx] = conf.ServerId
	}
	return serverIds, nil
}
