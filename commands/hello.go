package commands

import (
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-cli-plugin-template/client/livelog"
)

func GetHelloCommand() components.Command {
	return components.Command{
		Name:        "hello",
		Description: "Says Hello.",
		Aliases:     []string{"hi"},
		Flags:       getHelloFlags(),
		Action: func(c *components.Context) error {
			return livelog.CommandExample(c)
		},
	}
}

func getHelloFlags() []components.Flag {
	return []components.Flag{
		components.StringFlag{
			Name:        "server-id",
			Description: "Artifactory server ID configured using the config command.",
		},
	}
}
