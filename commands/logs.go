package commands

import (
	"context"
	"fmt"
	"github.com/hanoch-jfrog/forest/client/livelog"
	"github.com/hanoch-jfrog/forest/client/livelog/strategy"
	"github.com/hanoch-jfrog/forest/util"
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func GetLogsCommand() components.Command {
	return components.Command{
		Name:        "logs",
		Description: "Fetch the log of a desired service",
		Aliases:     []string{"l"},
		Arguments:   getLogsArguments(),
		Flags:       getLogsFlags(),
		Action:      logsCmd,
	}
}

func getLogsArguments() []components.Argument {
	return []components.Argument{
		{Name: "server_id", Description: "JFrog CLI Artifactory server id"},
		{Name: "node_id", Description: "Selected node id"},
		{Name: "log_name", Description: "Selected log name"},
	}
}

func getLogsFlags() []components.Flag {
	return []components.Flag{
		components.BoolFlag{
			Name:         "i",
			Description:  "Activate interactive menu",
			DefaultValue: false,
		},
		components.BoolFlag{
			Name:         "f",
			Description:  "Do 'tail -f' on the log",
			DefaultValue: false,
		},
	}
}

func listenForTermination(cancelCtx context.CancelFunc) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGABRT)
	go func() {
		<-c
		cancelCtx()
		fmt.Println("\r- Terminating")
		os.Exit(0)
	}()
}

func logsCmd(c *components.Context) error {
	isStreaming := c.GetBoolFlagValue("f")
	isInteractive := c.GetBoolFlagValue("i")

	mainCtx, mainCtxCancel := context.WithCancel(context.Background())
	defer mainCtxCancel()
	listenForTermination(mainCtxCancel)

	if !isInteractive {
		if len(c.Arguments) != 3 {
			return fmt.Errorf("wrong number of arguments. Expected: 3, " + "Received: " + strconv.Itoa(len(c.Arguments)))
		}
		serverId := c.Arguments[0]
		nodeId := c.Arguments[1]
		logFileName := c.Arguments[2]
		return buildServiceFromArguments(mainCtx, serverId, nodeId, logFileName, isStreaming)
	}
	return interactiveMenu(mainCtx, isStreaming)
}

func buildServiceFromArguments(ctx context.Context, cliServerId, nodeId, logName string, isStreaming bool) error {
	err := validateArgument("server id", cliServerId,
		func() ([]string, error) {
			return fetchAllServerIds()
		})
	if err != nil {
		return err
	}

	serviceManager, err := newArtifactoryServiceManager(cliServerId)
	if err != nil {
		return err
	}
	artifactoryHttpStrategy := strategy.NewArtifactoryHttpStrategy(serviceManager)
	client := livelog.NewClient(artifactoryHttpStrategy)

	err = validateArgument("node id", nodeId,
		func() ([]string, error) {
			return client.GetServiceNodeIds(ctx)
		})
	if err != nil {
		return err
	}
	client.SetNodeId(nodeId)

	var logsRefreshRate time.Duration
	err = validateArgument("log name", logName,
		func() ([]string, error) {
			srvConfig, fetchErr := client.GetConfig(ctx)
			if fetchErr != nil {
				return nil, fetchErr
			}
			logsRefreshRate = util.MillisToDuration(srvConfig.RefreshRateMillis)
			return srvConfig.LogFileNames, nil
		})
	if err != nil {
		return err
	}
	client.SetLogFileName(logName)
	client.SetLogsRefreshRate(logsRefreshRate)
	return printLogs(ctx, client, isStreaming)
}

func validateArgument(argumentName string, wantedVal string, allValues func() ([]string, error)) error {
	values, err := allValues()
	if err != nil {
		return err
	}
	if len(values) == 0 {
		return fmt.Errorf("no %v found", argumentName)
	}
	if !util.InSlice(values, wantedVal) {
		return fmt.Errorf("%v not found [%v], consider using one of the following %v values [%v]", argumentName, wantedVal, argumentName, util.SliceToCsv(values))
	}
	return nil
}

func interactiveMenu(ctx context.Context, isStreaming bool) error {
	selectedCliServerId, err := selectCliServerId()
	if err != nil {
		return err
	}
	serviceManager, err := newArtifactoryServiceManager(selectedCliServerId)
	if err != nil {
		return err
	}
	artifactoryStrategy := strategy.NewArtifactoryHttpStrategy(serviceManager)
	client := livelog.NewClient(artifactoryStrategy)
	nodeId, err := selectNodeId(ctx, client)
	if err != nil {
		return err
	}
	client.SetNodeId(nodeId)
	logName, logsRefreshRate, err := selectLogNameAndFetchRefreshRate(ctx, client)
	if err != nil {
		return err
	}
	client.SetLogFileName(logName)
	client.SetLogsRefreshRate(logsRefreshRate)
	return printLogs(ctx, client, isStreaming)
}

func printLogs(ctx context.Context, client livelog.Client, isStreaming bool) error {
	if isStreaming {
		return client.TailLog(ctx, os.Stdout)
	}
	return client.CatLog(ctx, os.Stdout)
}
