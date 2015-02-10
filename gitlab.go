package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

type errorResponse struct {
	Errors []string `json:"error"`
}

type sessionRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

const MERGE_REQUEST_STATE_OPENED string = "opened"

type gitlab struct {
	scheme  string
	host    string
	apiPath string
	token   string
}

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
		return nil, g.getErrorFromResponse(resp)
	}

	responseDecoder := json.NewDecoder(resp.Body)
	var newMergeRequest mergeRequest
	err = responseDecoder.Decode(&newMergeRequest)
	if nil != err {
		return nil, err
	}

	return &newMergeRequest, nil
}

func (g gitlab) getErrorFromResponse(resp *http.Response) error {
	// Try getting error response from gitlab
	responseDecoder := json.NewDecoder(resp.Body)
	var errorResp errorResponse

	err := responseDecoder.Decode(&errorResp)
	if nil != err || len(errorResp.Errors) == 0 {
		return fmt.Errorf("Expected status 201, got %d\n", resp.StatusCode)
	} else {
		return fmt.Errorf("Gitlab: %s\n", strings.Join(errorResp.Errors, ", "))
	}
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

func (g gitlab) acceptMergeRequest(projectId string, mergeRequestId int) error {

	addr := g.getApiUrl(
		"projects",
		url.QueryEscape(projectId),
		"merge_request",
		url.QueryEscape(strconv.Itoa(mergeRequestId)),
		"merge",
	)

	req, err := http.NewRequest("PUT", addr, nil)
	req.URL = &url.URL{
		Scheme: g.scheme,
		Host:   g.host,
		// Use opaque url to preserve "%2F"
		Opaque: g.getOpaqueApiUrl("projects", url.QueryEscape(projectId), "merge_request", strconv.Itoa(mergeRequestId), "merge"),
	}

	client := http.Client{}
	resp, err := client.Do(req)

	if nil != err {
		return err
	}

	if resp.StatusCode == 404 {
		if g.token != "" {
			addr = strings.Replace(addr, g.token, "***", -1)
		}
		return fmt.Errorf("404: %s %s\n", req.Method, addr)
	}

	if resp.StatusCode != 200 {
		return g.getErrorFromResponse(resp)
	}

	return nil
}
