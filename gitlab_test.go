package main

import (
	"encoding/json"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetSession(t *testing.T) {
	var req sessionRequest
	var method string
	var requestUrl string

	sr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&req)
		if nil != err {
			t.Fatal(err)
		}

		requestUrl = r.URL.String()
		method = r.Method
	}))

	u, err := url.Parse(sr.URL)
	if nil != err {
		t.Fatal(err)
	}

	g := newGitlab(u.Host)
	g.getSession("user", "password")

	expectedUrl := fmt.Sprintf("http://%s/api/v3/session", u.Host)
	if "POST" != method || expectedUrl != requestUrl {
		t.Fatalf("Expected POST request to %q, got %s request to %q", expectedUrl, method, requestUrl)
	}

	expectedUser := "user"
	if expectedUser != req.Login {
		t.Fatalf("Expected login: %q, got %q\n", expectedUser, req.Login)
	}

	expectedPassword := "password"
	if expectedUser != req.Login {
		t.Fatalf("Expected password: %q, got %q\n", expectedPassword, req.Password)
	}
}

func testGetPrivateTokenUrl(t *testing.T) {
	Convey("Given a gitlab instance", t, func() {
		g := newGitlab("1.2.3.4")

		Convey("It should produce a url for obtaining a private token", t, func() {
			So(g.getPrivateTokenUrl(), ShouldEqual, "http://1.2.3.4/profile/account")
		})
	})
}
