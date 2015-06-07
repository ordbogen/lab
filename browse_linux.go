package main

// +build !darwin

import (
	"errors"
	"log"
	"os"
	"os/exec"
)

var NoGUIError error

func init() {
	NoGUIError = errors.New("No gui. Not browsing.")
}

// Only browse if there's a display available. No text browser.
func browseGUIPlatform(url string) error {
	if os.Getenv("DISPLAY") != "" {
		// x session
		log.Printf("Opening \"%s\"...\n", url)
		return exec.Command("xdg-open", url).Run()
	}

	return NoGUIError
}
