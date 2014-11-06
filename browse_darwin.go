package main

// +build darwin

import (
	"os/exec"
)

func browsePlatform(url string) error {
	return exec.Command("open", url).Run()
}
