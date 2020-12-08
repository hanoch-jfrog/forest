package commands

import (
	"context"
	"fmt"
	"github.com/hanoch-jfrog/forest/client/livelog"
	"github.com/hanoch-jfrog/forest/client/livelog/model"
	"github.com/hanoch-jfrog/forest/util"
	"github.com/manifoldco/promptui"
	"time"
)

func selectLogNameAndFetchRefreshRate(ctx context.Context, client livelog.Client) (selectedLogName string, logsRefreshRate time.Duration, err error) {
	var srvConfig *model.Config
	srvConfig, err = client.GetConfig(ctx)
	if err != nil {
		return
	}
	logsRefreshRate = util.MillisToDuration(srvConfig.RefreshRateMillis)
	selectedLogName, err = runInteractiveMenu("Select log name", "Available log names", srvConfig.LogFileNames)
	return
}

func selectNodeId(ctx context.Context, client livelog.Client) (string, error) {
	nodeIds, err := client.GetServiceNodeIds(ctx)
	if err != nil {
		return "", err
	}
	return runInteractiveMenu("Select node id", "Available nodes", nodeIds)
}

func selectCliServerId() (string, error) {
	serverIds, err := fetchAllServerIds()
	if err != nil {
		return "", err
	}
	return runInteractiveMenu("Select JFrog CLI server id", "Available server IDs", serverIds)
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
