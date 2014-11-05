package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type gitRemote struct {
	base string
	path string
}

type gitDir string

/// Get working directory
func (here gitDir) Getwd() (string, error) {
	return filepath.Abs(strings.TrimSuffix(string(here), ".git"))
}

type ErrUnknownRemote string

func (e ErrUnknownRemote) Error() string {
	return fmt.Sprintf("Could not find remote: %s\n", e)
}

func getRemoteUrlFromRemoteVOutput(remoteName string, output []byte) (string, error) {
	lines := strings.Split(string(output), "\n")
	prefix := remoteName + "\t"
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			remoteLine := strings.TrimPrefix(line, prefix)
			remoteLine = strings.TrimSuffix(remoteLine, " (fetch)")
			remoteLine = strings.TrimSuffix(remoteLine, " (push)")
			return remoteLine, nil
		}
	}

	return "", ErrUnknownRemote(remoteName)
}

func (here gitDir) checkout(arg string) error {
	// Get working directory
	wd, err := here.Getwd()
	if nil != err {
		return err
	}

	// Fetch first
	fetchCmd := exec.Command("git", "fetch")
	fetchCmd.Dir = wd
	fetchCmd.Stdout = os.Stdout
	fetchCmd.Stdin = os.Stdin
	err = fetchCmd.Run()
	if nil != err {
		return err
	}

	cmd := exec.Command("git", "checkout", arg)
	cmd.Dir = wd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (here gitDir) diff2(left, right string) error {
	// Get working directory
	wd, err := here.Getwd()
	if nil != err {
		return err
	}

	remotes := map[string]bool{"origin": true}

	// Left remote?
	if strings.Contains(left, "/") {
		leftRemote := strings.Split(left, "/")[0]
		remotes[leftRemote] = true
	} else {
		left = "origin/" + left
	}

	if strings.Contains(right, "/") {
		rightRemote := strings.Split(right, "/")[0]
		remotes[rightRemote] = true
	} else {
		right = "origin/" + right
	}

	for remote, _ := range remotes {
		// Fetch first
		fetchCmd := exec.Command("git", "fetch", remote)
		fetchCmd.Dir = wd
		fetchCmd.Stdout = os.Stdout
		fetchCmd.Stdin = os.Stdin
		err = fetchCmd.Run()
		if nil != err {
			return err
		}
	}

	cmd := exec.Command("git", "diff", left+".."+right, "--")
	cmd.Dir = wd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (here gitDir) getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "--git-dir", string(here), "branch")
	output, err := cmd.CombinedOutput()

	if nil != err {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "* ") {
			return strings.TrimPrefix(line, "* "), nil
		}
	}

	return "", fmt.Errorf("Could not find current branch in %s\n", here)
}

func parseRemote(remoteAddr string) (remote gitRemote) {
	// Strip user info
	if atIndex := strings.IndexByte(remoteAddr, '@'); atIndex >= 0 {
		remoteAddr = remoteAddr[atIndex+1 : len(remoteAddr)]
	}

	if schemeIndex := strings.Index(remoteAddr, "://"); schemeIndex >= 0 {
		remoteAddr = remoteAddr[schemeIndex+3 : len(remoteAddr)]
	}

	if i := strings.IndexAny(remoteAddr, ":/"); i >= 0 {
		remote = gitRemote{
			remoteAddr[0:i],
			remoteAddr[i+1 : len(remoteAddr)],
		}
	} else {
		// relative remote? Not interested anyway
	}

	remote.path = strings.TrimSuffix(remote.path, ".git")

	return remote
}

/// Get origin for given remote name
func (here gitDir) getRemoteUrl(remoteName string) (string, error) {
	cmd := exec.Command("git", "--git-dir", string(here), "remote", "-v")
	output, err := cmd.CombinedOutput()

	if nil != err {
		return "", fmt.Errorf("%s\n", output)
	}

	return getRemoteUrlFromRemoteVOutput(remoteName, output)
}
