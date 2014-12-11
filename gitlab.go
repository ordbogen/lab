package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type mergeRequest struct {
	Id           int    `json:"id"`
	Iid          int    `json:"iid"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
}

type mergeRequestCreateRequest struct {
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	Title        string `json:"title"`
}

type session struct {
	PrivateToken string `json:"private_token"`
}

type sessionRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type gitlab struct {
	scheme  string
	host    string
	apiPath string
	token   string
}

type activityFeed struct {
	Title   string        `xml:"title"`
	Entries []*feedCommit `xml:"entry"`
}

type feedCommit struct {
	Title   string    `xml:"title"`
	Updated time.Time `xml:"updated"`
}

const MERGE_REQUEST_STATE_OPENED string = "opened"
const DASHBOARD_FEED_PATH string = "/dashboard.atom"

func (g gitlab) getProjectUrl(path string) string {
	return g.scheme + "://" + g.host + "/" + strings.TrimPrefix(path, "/")
}

func (g gitlab) getMergeRequestUrl(projectId string, mergeRequestId int) string {
	projectId, _ = url.QueryUnescape(projectId)
	projectId = strings.Trim(projectId, "/")
	return g.scheme + "://" + g.host + "/" + projectId + "/merge_requests/" + strconv.Itoa(mergeRequestId)
}

func newGitlab(host string) gitlab {
	return gitlab{"http", host, "/api/v3", ""}
}

func (g gitlab) getPrivateTokenUrl() string {
	return g.scheme + "://" + g.host + "/profile/account"
}

func (g gitlab) getFeedUrl() string {
	return g.scheme + "://" + g.host + DASHBOARD_FEED_PATH + "?private_token=" + g.token
}

func (g gitlab) buildFeed(method, url string, body []byte) ([]byte, error) {
	var req *http.Request
	var err error

	if body != nil {
		reader := bytes.NewReader(body)
		req, err = http.NewRequest(method, url, reader)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		panic("Error while building gitlab request")
	}

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("client.Do error: %q", err)
	}

	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("%s", err)
	}

	if resp.StatusCode >= 400 {
		err = fmt.Errorf("buildFeed failed: <%d> %s", resp.StatusCode, req.URL)
	}

	return contents, err
}

func (g gitlab) getApiUrl(pathSegments ...string) string {
	return g.getUnauthApiUrl(pathSegments...) + "?private_token=" + g.token
}

func (g gitlab) getUnauthApiUrl(pathSegments ...string) string {
	return g.scheme + "://" + g.host + g.apiPath + "/" + strings.Join(pathSegments, "/")
}

func (g gitlab) getOpaqueApiUrl(pathSegments ...string) string {
	return g.getUnauthOpaqueApiUrl(pathSegments...) + "?private_token=" + g.token
}

func (g gitlab) getUnauthOpaqueApiUrl(pathSegments ...string) string {
	return "//" + g.host + g.apiPath + "/" + strings.Join(pathSegments, "/")
}

func (g gitlab) createMergeRequest(projectId, sourceBranch, targetBranch, title string) (*mergeRequest, error) {
	requestBody := mergeRequestCreateRequest{
		SourceBranch: sourceBranch,
		TargetBranch: targetBranch,
		Title:        title,
	}

	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	err := encoder.Encode(requestBody)
	if nil != err {
		return nil, err
	}

	addr := g.getApiUrl("projects", url.QueryEscape(projectId), "merge_requests")

	req, err := http.NewRequest("POST", addr, buffer)
	req.URL = &url.URL{
		Scheme: g.scheme,
		Host:   g.host,
		// Use opaque url to preserve "%2F"
		Opaque: g.getOpaqueApiUrl("projects", url.QueryEscape(projectId), "merge_requests"),
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if nil != err {
		return nil, err
	}

	if resp.StatusCode == 404 {
		// Duplicate merge request, same source branch
		return nil, fmt.Errorf("There already exists a merge request for: %s\n", sourceBranch)
	}

	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("Expected status 201, got %d\n", resp.StatusCode)
	}

	responseDecoder := json.NewDecoder(resp.Body)
	var newMergeRequest mergeRequest
	err = responseDecoder.Decode(&newMergeRequest)
	if nil != err {
		return nil, err
	}

	return &newMergeRequest, nil
}

/// Request session for private token
func (g gitlab) getSession(login, password string) (*session, error) {
	requestBody := sessionRequest{
		Login:    login,
		Password: password,
	}

	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	err := encoder.Encode(requestBody)
	if nil != err {
		return nil, err
	}

	addr := g.getUnauthApiUrl("session")

	req, err := http.NewRequest("POST", addr, buffer)
	req.URL = &url.URL{
		Scheme: g.scheme,
		Host:   g.host,
		// Use opaque url to preserve "%2F"
		Opaque: g.getUnauthOpaqueApiUrl("session"),
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if nil != err {
		return nil, err
	}

	if resp.StatusCode != 201 {
		return nil, errors.New(resp.Status)
	}

	responseDecoder := json.NewDecoder(resp.Body)
	var session session
	err = responseDecoder.Decode(&session)
	if nil != err {
		return nil, err
	}

	return &session, nil
}

func (g gitlab) queryMergeRequests(projectId string, state string) ([]mergeRequest, error) {

	if state == "" {
		state = MERGE_REQUEST_STATE_OPENED
	}
	addr := g.getApiUrl("projects", url.QueryEscape(projectId), "merge_requests") + "&state=" + state

	req, err := http.NewRequest("GET", addr, nil)
	req.URL = &url.URL{
		Scheme: g.scheme,
		Host:   g.host,
		// Use opaque url to preserve "%2F"
		Opaque: g.getOpaqueApiUrl("projects", url.QueryEscape(projectId), "merge_requests") + "&state=" + state,
	}

	client := http.Client{}
	resp, err := client.Do(req)

	if nil != err {
		return nil, err
	}

	if resp.StatusCode == 404 {
		if g.token != "" {
			addr = strings.Replace(addr, g.token, "***", -1)
		}
		return nil, fmt.Errorf("404: %s\n", addr)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(resp.Status)
	}

	var mergeRequests []mergeRequest
	d := json.NewDecoder(resp.Body)

	// Check for leading "<" -> "api 404"
	err = d.Decode(&mergeRequests)

	if nil != err {
		return nil, err
	}

	return mergeRequests, nil
}
