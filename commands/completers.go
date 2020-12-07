package commands

import (
	"context"
	"fmt"
	"github.com/hanoch-jfrog/forest/client/livelog"
	"github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/manifoldco/promptui"
	"time"
)

const requestTimeout = 15 * time.Second

func selectLogNameAndFetchRefreshRate(ctx context.Context) (selectedLogName string, logsRefreshRate time.Duration, err error) {
	var srvConfig livelog.Config
	srvConfig, err = fetchServerConfig(ctx)
	if err != nil {
		return
	}
	logsRefreshRate = livelog.MillisToDuration(srvConfig.RefreshRateMillis)
	selectedLogName, err = runInteractiveMenu("Please select log name", "Available log names", srvConfig.LogFileNames)
	return
}

func fetchServerConfig(ctx context.Context) (srvConfig livelog.Config, err error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	srvConfig, err = artifactoryServiceClient.GetConfig(timeoutCtx)
	if err != nil {
		return
	}
	if len(srvConfig.LogFileNames) == 0 {
		err = fmt.Errorf("no log file names were found")
		return
	}
	return
}

func selectNodeId(ctx context.Context) (string, error) {
	nodeIds, err := fetchAllNodeIds(ctx)
	if err != nil {
		return "", err
	}
	return runInteractiveMenu("Please select node number", "Available nodes", nodeIds)
}

func fetchAllNodeIds(ctx context.Context) ([]string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	serviceNodes, err := artifactoryServiceClient.GetServiceNodes(timeoutCtx)
	if err != nil {
		return nil, err
	}
	if len(serviceNodes.Nodes) == 0 {
		return nil, fmt.Errorf("no Node Ids found")
	}
	nodeIds := make([]string, len(serviceNodes.Nodes))
	for idx, serviceNode := range serviceNodes.Nodes {
		nodeIds[idx] = serviceNode.NodeID
	}
	return nodeIds, nil
}

func selectCliServerId() (string, error) {
	serverIds, err := fetchAllServerIds()
	if err != nil {
		return "", err
	}

	return runInteractiveMenu("Please select JFrog CLI server id", "Available server IDs", serverIds)
}

func fetchAllServerIds() ([]string, error) {
	configs, err := config.GetAllArtifactoryConfigs()
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

func runInteractiveMenu(selectionHeader string, selectionLabel string, values []string) (string, error) {
	if selectionHeader != "" {
		fmt.Println(selectionHeader)
	}

	selectMenu := promptui.Select{
		Label: selectionLabel,
		Items: values,
	}
	_, res, err := selectMenu.Run()
	return res, err
}
