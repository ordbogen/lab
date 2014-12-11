package main

import (
	"code.google.com/p/gopass"
	"encoding/xml"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

type config struct {
	PrivateToken string `toml:"private_token"`
}

func promptForMergeRequest(c *cli.Context) *mergeRequest {
	remoteUrl := needRemoteUrl(c)
	state := c.String("state")
	server := needGitlab(c)
	token := needToken(c)
	server.token = token

	format := c.String("format")
	if format == "" {
		format = MergeRequestCheckoutListTemplate
	}
	tmpl, err := newColorTemplate("default-merge-request-list-template", format)
	if nil != err {
		log.Fatal(err)
	}

	mergeRequests, err := server.queryMergeRequests(remoteUrl.path, state)
	if nil != err {
		log.Fatal(err)
	}
	for i, request := range mergeRequests {
		fmt.Fprintf(os.Stderr, color.RedString("%%d: "), i)
		err = tmpl.Execute(os.Stderr, request)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Prompt for id
	var mergeRequest mergeRequest
	for {
		fmt.Fprintf(os.Stderr, "Select a merge request: ")
		var id int
		_, err = fmt.Scanf("%d", &id)
		if nil != err {
			continue
		}
		if id > len(mergeRequests)-1 {
			continue
		}

		mergeRequest = mergeRequests[id]
		break
	}

	return &mergeRequest
}

/// Browse a url, x or text
func browse(url string) {
	log.Printf("Opening \"%s\"...\n", url)
	err := browsePlatform(url)
	if nil != err {
		log.Fatal(err)
	}
}

/// Get gitlab merge requests or fail!
func needMergeRequests(c *cli.Context) ([]mergeRequest, error) {
	token := needToken(c)
	server := needGitlab(c)
	server.token = token

	remoteUrl := needRemoteUrl(c)
	state := c.String("state")

	return server.queryMergeRequests(remoteUrl.path, state)
}

/// Get gitlab url or fail!
func needGitlab(c *cli.Context) gitlab {
	r := needRemoteUrl(c)
	for _, host := range []string{"github.com", "code.google.com", "bitbucket.org"} {
		if strings.HasSuffix(r.base, host) {
			log.Fatalf("Gitlab server on: \"%s\"? I don't think so\n", r.base)
		}
	}
	return newGitlab(r.base)
}

func needGitDir(c *cli.Context) gitDir {
	given := c.String("git-dir")
	var err error
	if given == "" {
		given, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
	}

	return gitDir(strings.TrimSuffix(given, "/") + "/.git")
}

/// Get remote url or fail!
func needRemoteUrl(c *cli.Context) gitRemote {
	remote := c.String("remote")
	git := needGitDir(c)
	remoteUrl, err := git.getRemoteUrl(remote)
	if nil != err {
		log.Fatal(err)
	}

	return parseRemote(remoteUrl)
}

// Get token or fail!
func needToken(c *cli.Context) string {
	token := c.String("token")
	gitDir := needGitDir(c)
	wd, err := gitDir.Getwd()
	if nil != err {
		log.Fatal(err)
	}

	projectLabFile := filepath.Join(wd, ".lab")
	if token == "" {
		// Try getting token from $PROJECT/.lab
		var config config
		_, err := toml.DecodeFile(projectLabFile, &config)
		if err != nil {
			// ...
			if os.IsNotExist(err) {
				// ~/.labrc does not exist, move on
			} else {
				log.Fatalf("%T\n", err)
			}
		}

		token = config.PrivateToken
		if token == "" {
			// Prompt for private token
			fmt.Fprintln(os.Stderr, "Login to get private token for gitlab")
			var login string
			var password string
			for {
				fmt.Fprint(os.Stderr, "Login: ")
				_, err := fmt.Scanf("%s", &login)
				err = nil
				if err != nil || login == "" {
					continue
				}
				password, err = gopass.GetPass("Password: ")
				err = nil
				if err != nil || password == "" {
					continue
				}

				if login != "" && password != "" {
					break
				}
			}

			server := needGitlab(c)
			session, err := server.getSession(login, password)
			if err != nil {
				log.Fatal(err)
			}
			token = session.PrivateToken

			// Write to $PROJECT/.lab
			f, err := os.OpenFile(projectLabFile, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0600)
			if nil != err {
				log.Fatal(err)
			}
			defer f.Close()

			config.PrivateToken = token
			enc := toml.NewEncoder(f)
			err = enc.Encode(config)
			if nil != err {
				log.Fatal(err)
			}
			fmt.Fprintf(os.Stderr, "Saved private token to %s\n", projectLabFile)
		}
	}

	// Use token from arguments or environment
	if token == "" {
		server := needGitlab(c)
		log.Fatal(
			"Could not get api token, get one from: \"",
			server.getPrivateTokenUrl(),
			"\n\nexport as LAB_PRIVATE_TOKEN or use as flag: --token <token>",
		)
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

	mergeRequestFlags := append(flags, cli.StringFlag{
		Name:  "state",
		Value: "opened",
	})

	app.Commands = []cli.Command{
		{
			Name:  "browse",
			Usage: "Open project homepage",
			Flags: flags,
			Action: func(c *cli.Context) {
				server := needGitlab(c)
				remote := needRemoteUrl(c)
				addr := server.getProjectUrl(remote.path)
				browse(addr)
			},
		},
		{
			Name:  "feed",
			Usage: "Get your GitLab feed",
			Flags: flags,
			Action: func(c *cli.Context) {
				server := needGitlab(c)
				token := needToken(c)
				server.token = token

				feedUrl := server.getFeedUrl()

				contents, err := server.buildFeed("GET", feedUrl, nil)
				if err != nil {
					log.Fatal("%s", err)
				}

				var activity activityFeed

				err = xml.Unmarshal(contents, &activity)

				if err != nil {
					log.Fatal("%s", err)
				}

				commits := activity.Entries

				// templating - feed title

				formatTitle := c.String("format")
				if formatTitle == "" {
					formatTitle = FeedTitleTemplate
				}

				titleTmpl, err := newTemplate("title-feed", formatTitle, doColors(os.Stdout))
				if nil != err {
					log.Fatal(err)
				}

				err = titleTmpl.Execute(os.Stdout, activity)
				if err != nil {
					log.Fatal(err)
				}

				// templating - feed entry

				format := c.String("format")
				if format == "" {
					format = FeedTemplate
				}

				tmpl, err := newTemplate("default-feed", format, doColors(os.Stdout))
				if nil != err {
					log.Fatal(err)
				}

				for _, commit := range commits {
					err = tmpl.Execute(os.Stdout, commit)
					if err != nil {
						log.Fatal(err)
					}
				}

				return
			},
		},
		{
			Name:      "merge-request",
			ShortName: "mr",
			Subcommands: []cli.Command{
				{
					Name:  "create",
					Usage: "Create merge request, default target branch: master.",
					Flags: flags,
					Action: func(c *cli.Context) {
						server := needGitlab(c)
						token := needToken(c)
						server.token = token
						remoteUrl := needRemoteUrl(c)
						gitDir := needGitDir(c)

						currentBranch, err := gitDir.getCurrentBranch()
						if nil != err {
							log.Fatal(err)
						}
						args := c.Args()

						targetBranch := args.First()
						if targetBranch == "" {
							targetBranch = "master"
						}

						title := args.Get(1)
						if title == "" {
							// Generate auto title
							title = strings.Replace(currentBranch, "-", " ", -1)
							title = strings.Replace(title, "_", " ", -1)
						}

						createdMergeRequest, err := server.createMergeRequest(remoteUrl.path, currentBranch, targetBranch, title)
						if nil != err {
							log.Fatal(err)
						}

						log.Println("Created merge request:", server.getMergeRequestUrl(remoteUrl.path, createdMergeRequest.Iid))
					},
				},
				{
					Name:  "browse",
					Usage: "Browse current merge request or by ID.",
					Flags: mergeRequestFlags,
					Action: func(c *cli.Context) {
						server := needGitlab(c)
						remoteUrl := needRemoteUrl(c)
						gitDir := needGitDir(c)

						mergeRequests, err := needMergeRequests(c)
						if nil != err {
							log.Fatal(err)
						}

						if c.Args().First() != "" {
							mergeRequestId, err := strconv.Atoi(c.Args().First())
							if err != nil {
								log.Fatalf("You did not provide a valid ID")
							}

							for _, request := range mergeRequests {
								if request.Iid == mergeRequestId {
									browse(server.getMergeRequestUrl(remoteUrl.path, mergeRequestId))
									return
								}
							}

							log.Fatalf("Unable to find merge request with ID #%d\n", mergeRequestId)
						}

						currentBranch, err := gitDir.getCurrentBranch()
						if nil != err {
							log.Fatal(err)
						}

						for _, request := range mergeRequests {
							if request.SourceBranch == currentBranch {
								browse(server.getMergeRequestUrl(remoteUrl.path, request.Iid))
								return
							}
						}

						log.Fatalf("Could not find merge request for branch: %s on project %s\n", currentBranch, remoteUrl.path)
					},
				},
				{
					Name:  "diff",
					Usage: "Diff current merge request or by ID.",
					Flags: mergeRequestFlags,
					Action: func(c *cli.Context) {
						server := needGitlab(c)
						remoteUrl := needRemoteUrl(c)
						gitDir := needGitDir(c)

						mergeRequests, err := needMergeRequests(c)
						if nil != err {
							log.Fatal(err)
						}

						if c.Args().First() != "" {
							mergeRequestId, err := strconv.Atoi(c.Args().First())
							if err != nil {
								log.Fatalf("You did not provide a valid ID")
							}

							for _, request := range mergeRequests {
								if request.Iid == mergeRequestId {
									browse(server.getMergeRequestUrl(remoteUrl.path, mergeRequestId))
									return
								}
							}

							log.Fatalf("Unable to find merge request with ID #%d\n", mergeRequestId)
						}

						currentBranch, err := gitDir.getCurrentBranch()
						if nil != err {
							log.Fatal(err)
						}

						for _, request := range mergeRequests {
							if request.SourceBranch == currentBranch {
								gitDir.diff2(request.TargetBranch, request.SourceBranch)
								return
							}
						}

						log.Fatalf("Could not find merge request for branch: %s on project %s\n", currentBranch, remoteUrl.path)
					},
				},
				{
					Name:  "pick-diff",
					Usage: "Pick diff from merge requests",
					Flags: mergeRequestFlags,
					Action: func(c *cli.Context) {
						gitDir := needGitDir(c)

						request := promptForMergeRequest(c)
						if nil == request {
							return
						}

						gitDir.diff2(request.TargetBranch, request.SourceBranch)
					},
				},
				{
					Name:  "list",
					Usage: "List merge requests",
					Flags: mergeRequestFlags,
					Action: func(c *cli.Context) {
						mergeRequests, err := needMergeRequests(c)
						if nil != err {
							log.Fatal(err)
						}

						format := c.String("format")
						if format == "" {
							format = MergeRequestListTemplate
						}

						if format == "help" {
							fmt.Println(MergeRequestListTemplate)
							return
						}

						tmpl, err := newTemplate("default-merge-request", format, doColors(os.Stdout))
						if nil != err {
							log.Fatal(err)
						}

						for _, request := range mergeRequests {
							err = tmpl.Execute(os.Stdout, request)
							if err != nil {
								log.Fatal(err)
							}
						}

						countTmpl, err := newTemplate("count", "{{ .count | red | bold }} {{ \"merge requests\" | blue }}\n", true)
						err = countTmpl.Execute(os.Stderr, map[string]string{
							"count": strconv.Itoa(len(mergeRequests)),
						})
						if nil != err {
							log.Fatal(err)
						}
					},
				},
				{
					Name:  "checkout",
					Usage: "Checkout branch from merge request",
					Flags: mergeRequestFlags,
					Action: func(c *cli.Context) {
						mergeRequest := promptForMergeRequest(c)
						if mergeRequest == nil {
							return
						}
						fmt.Printf("Checkout out: \"%s\"...", mergeRequest.SourceBranch)
						gitDir := needGitDir(c)

						err := gitDir.checkout(mergeRequest.SourceBranch)
						if nil != err {
							log.Fatal(err)
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
