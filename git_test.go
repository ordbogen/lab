package main

import (
	"testing"
)

func TestGetRemoteUrlOrigin(t *testing.T) {
	output := []byte(`
origin	https://github.com/educas/lab (fetch)
`)

	parsed, err := getRemoteUrlFromRemoteVOutput("origin", output)

	if err != nil {
		t.Fatal(err)
	}

	if parsed != "https://github.com/educas/lab" {
		t.Fatal(parsed)
	}

}

func TestGetRemoteUrlFromNothing(t *testing.T) {

	output := []byte("")
	parsed, err := getRemoteUrlFromRemoteVOutput("whatever", output)

	if parsed != "" {
		t.Fatal("Should not get parsed output:", parsed)
	}

	if _, ok := err.(ErrUnknownRemote); !ok {
		t.Fatalf("Expected ErrNoOrigin error, got: %+v\n", err, err)
	}
}

func TestGetOtherRemoteUrl(t *testing.T) {
	output := []byte(`
origin	https://something.com/wrong
something	https://something.com/right
`)

	parsed, err := getRemoteUrlFromRemoteVOutput("something", output)
	if parsed != "https://something.com/right" {
		t.Fatal("Expected remote: \"https://something.com/right\", got:", parsed)
	}

	if nil != err {
		t.Fatal(err)
	}
}

func TestParseGitHttpRemote(t *testing.T) {
	remote := parseRemote("https://git.something.org/someday/somewhere.git")

	if remote.base != "git.something.org" {
		t.Fatal("Expected remote base: \"git.something.org\", got:", remote.base)
	}

	if remote.path != "someday/somewhere" {
		t.Fatal("Expected remote path: \"someday/somewhere\", got:", remote.path)
	}
}
