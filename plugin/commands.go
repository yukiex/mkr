package plugin

import (
	"fmt"

	"gopkg.in/urfave/cli.v1"
)

var CommandPlugin = cli.Command{
	Name:        "plugin",
	Usage:       "Manage mackerel plugin",
	Description: `Manage mackerel plugin`,
	Subcommands: []cli.Command{
		{
			Name:        "install",
			Usage:       "install mackerel plugin",
			Description: `WIP`,
			Action:      doPluginInstall,
		},
	},
	Hidden: true,
}

func doPluginInstall(c *cli.Context) error {
	fmt.Println("do plugin install")
	return nil
}
