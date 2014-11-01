package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type mergeRequest struct {
	Title       string `json:title`
	Description string `json:description`
}

type gitlab struct {
	scheme  string
	host    string
	apiPath string
	token   string
}

func newGitlab(host, token string) gitlab {
	return gitlab{"http", host, "/api/v3", token}
}

func (g gitlab) getApiUrl(pathSegments ...string) string {
	return g.scheme + "://" + g.host + g.apiPath + "/" + strings.Join(pathSegments, "/") + "?private_token=" + g.token
}

func (g gitlab) getOpaqueApiUrl(pathSegments ...string) string {
	return "//" + g.host + g.apiPath + "/" + strings.Join(pathSegments, "/") + "?private_token=" + g.token
}

func (g gitlab) querymergeRequests(projectId string) ([]mergeRequest, error) {
	addr := g.getApiUrl("projects", url.QueryEscape(projectId), "merge_requests")

	req, err := http.NewRequest("GET", addr, nil)
	req.URL = &url.URL{
		Scheme: g.scheme,
		Host:   g.host,
		// Use opaque url to preserve "%2F"
		Opaque: g.getOpaqueApiUrl("projects", url.QueryEscape(projectId), "merge_requests"),
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
