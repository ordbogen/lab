package main

import (
	"fmt"
	"os/exec"
	"strings"
)

type gitRemote struct {
	base string
	path string
}

type gitDir string

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
