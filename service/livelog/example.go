package livelog

import (
	"context"
	"fmt"
	"github.com/jfrog/jfrog-cli-core/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"io"
	"os"
	"time"
)

// demonstrates how to call this service from the command scope
func commandExample(c *components.Context) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	// ask which remote service to use: (e.g. artifactory / distribution / etc.)
	fmt.Println("Which service? [artifactory]")
	rtDetails, err := getRtDetails(c)
	if err != nil {
		panic("Aghhhh")
	}
	rtServiceManager, err := utils.CreateServiceManager(rtDetails, false)
	liveLogService := NewArtifactoryService(rtServiceManager)

	// get nodes for parameters
	nodes, err := liveLogService.GetServiceNodes(ctx)
	if err != nil {
		panic("Aghhhh")
	}
	if len(nodes) == 0 {
		fmt.Println("No available nodes for the remote service")
		return
	}

	// ask which node to use:
	fmt.Println(fmt.Sprintf("Which node? %v", nodes))
	liveLogService.SetLogFileName(nodes[0])

	// get config for parameters
	conf, err := liveLogService.GetConfig(ctx)
	if err != nil {
		panic("Aghhhh")
	}
	if len(conf.LogFileNames) == 0 {
		fmt.Println("No available log files for the remote service node")
		return
	}

	// ask which log file to use:
	fmt.Println(fmt.Sprintf("Which log file? %v", conf.LogFileNames))
	liveLogService.SetLogFileName(conf.LogFileNames[0])

	// cat
	timeoutCtx1, cancel1 := context.WithTimeout(context.Background(), time.Minute)
	defer cancel1()

	catReader := liveLogService.CatLog(timeoutCtx1)
	_, err = io.Copy(os.Stdout, catReader)
	if err != nil {
		panic("Aghhhh")
	}

	// tail
	timeoutCtx2, cancel2 := context.WithTimeout(context.Background(), time.Minute)
	defer cancel2()

	tailReader := liveLogService.CatLog(timeoutCtx2)
	_, err = io.Copy(os.Stdout, tailReader)
	if err != nil {
		panic("Aghhhh")
	}
}
