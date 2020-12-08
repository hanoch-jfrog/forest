package commands

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands"
	"github.com/jfrog/jfrog-cli-core/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/utils/config"
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

func getRtDetails(serverId string) (*config.ArtifactoryDetails, error) {
	details, err := commands.GetConfig(serverId, false)
	if err != nil {
		return nil, err
	}
	if details.Url == "" {
		return nil, fmt.Errorf("no server-id was found, or the server-id has no url")
	}
	details.Url = clientutils.AddTrailingSlashIfNeeded(details.Url)
	err = config.CreateInitialRefreshableTokensIfNeeded(details)
	if err != nil {
		return nil, err
	}
	return details, nil
}
