package livelog

import (
	"context"
	"fmt"
	"github.com/jfrog/jfrog-cli-core/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io"
	"os"
	"time"
)

// demonstrates how to call this service from the command scope
func CommandExample(c *components.Context) error {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	// ask which remote service to use: (e.g. artifactory / distribution / etc.)
	log.Output("Which service? [artifactory]")
	rtDetails, err := getRtDetails(c)
	if err != nil {
		return fmt.Errorf("failed getting artifactory details")
	}
	rtServiceManager, err := utils.CreateServiceManager(rtDetails, false)
	liveLogService := NewArtifactoryClient(rtServiceManager)

	// get nodes for parameters
	nodes, err := liveLogService.GetServiceNodes(ctx)
	if err != nil {
		return fmt.Errorf("failed getting available nodes from artifactory: %v", err)
	}
	if len(nodes) == 0 {
		return fmt.Errorf("no available nodes for the remote service")
	}

	// ask which node to use:
	log.Output(fmt.Sprintf("Which node? %v", nodes))
	liveLogService.SetLogFileName(nodes[0])

	// get config for parameters
	conf, err := liveLogService.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed getting livelog configuration from artifactory: %v", err)
	}
	if len(conf.LogFileNames) == 0 {
		return fmt.Errorf("no available log files for the remote service node")
	}

	// ask which log file to use:
	log.Output(fmt.Sprintf("Which log file? %v", conf.LogFileNames))
	liveLogService.SetLogFileName(conf.LogFileNames[0])

	// cat
	timeoutCtx1, cancel1 := context.WithTimeout(context.Background(), time.Minute)
	defer cancel1()

	catReader := liveLogService.CatLog(timeoutCtx1)
	_, err = io.Copy(os.Stdout, catReader)
	if err != nil {
		return fmt.Errorf("failed reading logs from artifactory: %v", err)
	}

	// tail
	timeoutCtx2, cancel2 := context.WithTimeout(context.Background(), time.Minute)
	defer cancel2()

	tailReader := liveLogService.CatLog(timeoutCtx2)
	_, err = io.Copy(os.Stdout, tailReader)
	if err != nil {
		return fmt.Errorf("failed reading logs from artifactory: %v", err)
	}
	return nil
}
