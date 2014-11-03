package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"log"
	"os"
	"strings"
	"text/template"
)

const MergeRequestListTemplate string = `
# {{ .Title }}

{{ .Description}}
---
`

func getGitDir(given string) string {
	var err error
	if given == "" {
		given, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
	}

	return strings.TrimSuffix(given, "/") + "/.git"
}

/// Get gitlab url or fail!
func needGitlab(c *cli.Context) gitlab {
	r := needRemoteUrl(c)
	return newGitlab(r.base, c.String("token"))

}

/// Get remote url or fail!
func needRemoteUrl(c *cli.Context) gitRemote {
	remote := c.String("remote")
	dir := getGitDir(c.String("git-dir"))
	git := gitDir(dir)
	remoteUrl, err := git.getRemoteUrl(remote)
	if nil != err {
		log.Fatal(err)
	}

	return parseRemote(remoteUrl)
}

// Get token or fail!
func needToken(c *cli.Context) string {
	token := c.String("token")
	if token == "" {
		server := needGitlab(c)
		log.Fatal("Could not get api token, get one from:", server.getPrivateTokenUrl())
	}

	return token
}

func main() {

	app := cli.NewApp()
	app.Name = "lab"
	app.Usage = "Command-line client for Gitlab"

	flags := []cli.Flag{
		cli.StringFlag{
			Name: "git-dir",
		},
		cli.StringFlag{
			Name:  "remote",
			Value: "origin",
		},
		cli.StringFlag{
			Name:   "token",
			EnvVar: "LAB_PRIVATE_TOKEN",
		},
		cli.StringFlag{
			Name: "format, f",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "browse",
			Usage: "Open project homepage",
			Flags: flags,
			Action: func(c *cli.Context) {
				// TODO, cross-platform xdg-open: https://github.com/skratchdot/open-golang
				remote := c.String("remote")
				dir := getGitDir(c.String("git-dir"))

				format := c.String("format")
				if format == "" {
					format = MergeRequestListTemplate
				}

				if format == "help" {
					fmt.Println(MergeRequestListTemplate)
					return
				}

				git := gitDir(dir)
				remoteUrl, err := git.getRemoteUrl(remote)
				if nil != err {
					log.Fatal(err)
				}

				r := parseRemote(remoteUrl)

				server := newGitlab(r.base, c.String("token"))

				server.browseProject(r.path)
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
					Flags: flags,
					Action: func(c *cli.Context) {
						_ = needToken(c)
						server := needGitlab(c)
						remoteUrl := needRemoteUrl(c)

						format := c.String("format")
						if format == "" {
							format = MergeRequestListTemplate
						}

						if format == "help" {
							fmt.Println(MergeRequestListTemplate)
							return
						}

						mergeRequests, err := server.querymergeRequests(remoteUrl.path)
						if nil != err {
							log.Fatal(err)
						}

						tmpl := template.New("default-merge-request")
						tmpl, err = tmpl.Parse(format)
						if nil != err {
							log.Fatal(err)
						}

						for _, request := range mergeRequests {
							err = tmpl.Execute(os.Stdout, request)
							if err != nil {
								log.Fatal(err)
							}
						}
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
