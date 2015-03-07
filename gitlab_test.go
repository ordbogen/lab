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

func serveAndCatchJson(t *testing.T, jsonOut interface{}) (*httptest.Server, chan *http.Request) {
	reqChan := make(chan *http.Request, 1)
	serv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqChan <- r
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(jsonOut)
		if nil != err {
			t.Fatal(err)
		}
	}))
	return serv, reqChan
}

func urlMustParse(t *testing.T, urlStr string) *url.URL {
	u, err := url.Parse(urlStr)
	if nil != err {
		t.Fatal(err)
	}
	return u
}

func TestGetSession(t *testing.T) {
	Convey("Given a gitlab server", t, func() {
		var reqData sessionRequest

		sr, reqChan := serveAndCatchJson(t, &reqData)
		u := urlMustParse(t, sr.URL)

		g := newGitlab(u.Host)

		Convey("When requesting a session (private token)", func() {
			g.getSession("user", "password")

			Convey("The client should post the correct json to the api", func() {
				expectedUrl := fmt.Sprintf("http://%s/api/v3/session", u.Host)
				req := <-reqChan
				So(req.Method, ShouldEqual, "POST")
				So(req.URL.String(), ShouldEqual, expectedUrl)

				So(reqData, ShouldResemble, sessionRequest{"user", "password"})
			})
		})
	})
}

func testGetPrivateTokenUrl(t *testing.T) {
	Convey("Given a gitlab instance", t, func() {
		g := newGitlab("1.2.3.4")

		Convey("It should produce a url for obtaining a private token", func() {
			So("something", ShouldEqual, "something else")
			So(g.getPrivateTokenUrl(), ShouldEqual, "http://1.2.3.4/profile/account")
		})
	})
}

func TestCreateMergeRequest(t *testing.T) {
	Convey("Given a merge request", t, func() {
		var mr mergeRequestCreateRequest

		sr, reqChan := serveAndCatchJson(t, &mr)
		u := urlMustParse(t, sr.URL)
		g := newGitlab(u.Host)
		g.token = "my-private-token"

		Convey("When creating a merge request", func() {
			g.createMergeRequest("17", "source-branch", "target-branch", "my title")

			Convey("The request should match", func() {
				req := <-reqChan
				So(req.Method, ShouldEqual, "POST")
				So(
					req.URL.String(),
					ShouldEqual,
					fmt.Sprintf("http://%s/api/v3/projects/17/merge_requests?private_token=my-private-token", u.Host),
				)
				So(
					mr,
					ShouldResemble,
					mergeRequestCreateRequest{
						Title:        "my title",
						SourceBranch: "source-branch",
						TargetBranch: "target-branch",
					},
				)
			})
		})
	})
}

func TestQueryMergeRequests(t *testing.T) {
	Convey("Given a gitlab server", t, func() {
		var req *http.Request
		mrs := []mergeRequest{
			mergeRequest{
				Id:           13,
				Iid:          17,
				Title:        "my-title",
				Description:  "my-description",
				SourceBranch: "source-branch",
				TargetBranch: "target-branch",
			},
		}

		sr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			enc := json.NewEncoder(w)
			err := enc.Encode(mrs)
			if nil != err {
				t.Fatal(err)
			}
			req = r
		}))

		u := urlMustParse(t, sr.URL)
		g := newGitlab(u.Host)
		g.token = "my-private-token"

		Convey("When creating a merge request", func() {
			gottenMrs, err := g.queryMergeRequests("17", "shuffled")

			Convey("The request should match", func() {
				So(err, ShouldBeNil)
				So(gottenMrs, ShouldResemble, mrs)
				So(req.Method, ShouldEqual, "GET")
				So(
					req.URL.String(),
					ShouldEqual,
					fmt.Sprintf(
						"http://%s/api/v3/projects/17/merge_requests?private_token=my-private-token&state=shuffled",
						u.Host,
					),
				)
			})
		})
	})
}

func TestQueryMergeRequestsErrorMessage(t *testing.T) {
	Convey("Given a gitlab server", t, func() {
		var req *http.Request

		sr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(417)
			enc := json.NewEncoder(w)
			err := enc.Encode(map[string]string{
				"message": "my error message",
			})
			if nil != err {
				t.Fatal(err)
			}
			req = r
		}))

		u := urlMustParse(t, sr.URL)
		g := newGitlab(u.Host)
		g.token = "my-private-token"

		Convey("When creating a merge request", func() {
			_, err := g.queryMergeRequests("17", "shuffled")

			Convey("The request should match", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "my error message")
				So(req.Method, ShouldEqual, "GET")
				So(
					req.URL.String(),
					ShouldEqual,
					fmt.Sprintf(
						"http://%s/api/v3/projects/17/merge_requests?private_token=my-private-token&state=shuffled",
						u.Host,
					),
				)
			})
		})
	})
}
