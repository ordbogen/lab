package main

import (
	"github.com/codegangsta/cli"
	"log"
	"os"
)

func main() {

	app := cli.NewApp()
	app.Name = "lab"
	app.Usage = "Interface with gitlab"

	app.Commands = []cli.Command{
		{
			Name:      "complete",
			ShortName: "c",
			Usage:     "complete a task on the list",
			Action: func(c *cli.Context) {
				log.Fatal("")
			},
		},
		{
			Name:      "merge-request",
			ShortName: "mr",
			Usage:     "do something with merge requests",
			Subcommands: []cli.Command{
				{
					Name:  "list",
					Usage: "list merge requests",
					Action: func(c *cli.Context) {
						log.Fatal("TODO: lab mr list")
					},
				},
				{
					Name:  "create",
					Usage: "create a merge request",
					Action: func(c *cli.Context) {
						log.Fatal(`Create merge request:
1. lab mr create -> <current-branch>..<lab default branch>
2. lab mr create <target branch> ->  <current-branch>..<target branch>
3. lab mr create <source branch>..<target branch> -> <source branch>..<target branch>
					`)
					},
				},
				{
					// Accept using the current branch or a given mr id
					Name:  "accept",
					Usage: "accept merge request by the current branch",
					Action: func(c *cli.Context) {
						log.Fatal("TODO: lab mr accept [<id>]")
					},
				},
			},
		},
	}

	app.Run(os.Args)
}
