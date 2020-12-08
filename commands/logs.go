package commands

import (
	"context"
	"fmt"
	"github.com/hanoch-jfrog/forest/client/livelog"
	"github.com/hanoch-jfrog/forest/client/livelog/strategy"
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var livelogClient livelog.Client

func GetLogsCommand() components.Command {
	return components.Command{
		Name:        "logs",
		Description: "Fetch the log of a desired service",
		Aliases:     []string{"l"},
		Arguments:   getLogsArguments(),
		Flags:       getLogsFlags(),
		EnvVars:     getLogsEnvVar(),
		Action: func(c *components.Context) error {
			return logsCmd(c)
		},
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

func getLogsEnvVar() []components.EnvVar {
	return []components.EnvVar{}
}

func SetupCloseHandler(cancelCtx context.CancelFunc) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancelCtx()
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		os.Exit(0)
	}()
}

func logsCmd(c *components.Context) error {
	if len(c.Arguments) != 3 && len(c.Arguments) != 0 {
		return fmt.Errorf("wrong number of arguments. Expected: 3 or 0, " + "Received: " + strconv.Itoa(len(c.Arguments)))
	}

	mainCtx, mainCtxCancel := context.WithCancel(context.Background())
	defer mainCtxCancel()
	SetupCloseHandler(mainCtxCancel)

	logTail := c.GetBoolFlagValue("f")
	enableInteractiveMenu := c.GetBoolFlagValue("i")

	var err error
	if enableInteractiveMenu {
		livelogClient, err = interactiveMenu(mainCtx)
		if err != nil {
			return err
		}
	} else {
		if len(c.Arguments) != 3 {
			return fmt.Errorf("wrong number of arguments. Expected: 3, " + "Received: " + strconv.Itoa(len(c.Arguments)))
		}

		serverId := c.Arguments[0]
		nodeId := c.Arguments[1]
		logName := c.Arguments[2]

		livelogClient, err = buildServiceFromArguments(mainCtx, serverId, nodeId, logName)
		if err != nil {
			return err
		}
	}
	return printLogs(mainCtx, logTail)
}

func buildServiceFromArguments(ctx context.Context, cliServerId, nodeId, logName string) (livelog.Client, error) {
	err := validateArgument("server id", cliServerId,
		func() ([]string, error) {
			return fetchAllServerIds()
		})
	if err != nil {
		return nil, err
	}

	serviceManager, err := newArtifactoryServiceManager(cliServerId)
	if err != nil {
		return nil, err
	}
	artiStrategy := strategy.NewArtifactoryLiveLogStrategy(serviceManager)
	livelogClient = livelog.NewClient(artiStrategy)

	err = validateArgument("node id", nodeId,
		func() ([]string, error) {
			return fetchAllNodeIds(ctx)
		})
	if err != nil {
		return nil, err
	}
	livelogClient.SetNodeId(nodeId)

	var logsRefreshRate time.Duration
	err = validateArgument("log name", logName,
		func() ([]string, error) {
			srvConfig, fetchErr := fetchServerConfig(ctx)
			if fetchErr != nil {
				return nil, fetchErr
			}
			logsRefreshRate = MillisToDuration(srvConfig.RefreshRateMillis)
			return srvConfig.LogFileNames, nil
		})
	if err != nil {
		return nil, err
	}
	livelogClient.SetLogFileName(logName)
	livelogClient.SetLogsRefreshRate(logsRefreshRate)
	return livelogClient, nil
}

func validateArgument(argumentName string, wantedVal string, allValues func() ([]string, error)) error {
	values, err := allValues()
	if err != nil {
		return err
	}
	if len(values) == 0 {
		return fmt.Errorf("no %v found", argumentName)
	}
	if !InSlice(values, wantedVal) {
		return fmt.Errorf("%v not found [%v], consider using one of the following %v [%v]", argumentName, wantedVal, argumentName, SliceToCsv(values))
	}
	return nil
}

func interactiveMenu(ctx context.Context) (livelog.Client, error) {
	selectedCliServerId, err := selectCliServerId()
	if err != nil {
		return nil, err
	}
	serviceManager, err := newArtifactoryServiceManager(selectedCliServerId)
	if err != nil {
		return nil, err
	}
	artiStrategy := strategy.NewArtifactoryLiveLogStrategy(serviceManager)
	livelogClient = livelog.NewClient(artiStrategy)

	nodeId, err := selectNodeId(ctx)
	if err != nil {
		return nil, err
	}
	livelogClient.SetNodeId(nodeId)

	logName, logsRefreshRate, err := selectLogNameAndFetchRefreshRate(ctx)
	if err != nil {
		return nil, err
	}
	livelogClient.SetLogFileName(logName)
	livelogClient.SetLogsRefreshRate(logsRefreshRate)

	return livelogClient, nil
}

func printLogs(ctx context.Context, tail bool) error {
	var err error
	if tail {
		err = livelogClient.TailLog(ctx, os.Stdout)
	} else {
		err = livelogClient.CatLog(ctx, os.Stdout)
	}
	return err
}
