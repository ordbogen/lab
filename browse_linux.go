package main

// +build !darwin

import (
	"os"
	"os/exec"
	"syscall"
)

func browsePlatform(url string) error {
	if os.Getenv("DISPLAY") != "" {
		// x session
		return exec.Command("xdg-open", url).Run()
	} else {
		// text
		return syscall.Exec("/usr/bin/www-browser", []string{"www-browser", url}, os.Environ())
	}
}
