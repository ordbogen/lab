package main

// +build darwin

import (
	"os/exec"
)

func browsePlatform(url string) error {
	log.Printf("Opening \"%s\"...\n", url)
	return exec.Command("open", url).Run()
}

func browseGUIPlatform(url string) error {
	return browsePlatform(url)
}
