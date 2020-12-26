package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// GithubAPIBase github api uri
var GithubAPIBase = "https://api.github.com"

func createGithubIssue(env Environment, issueRequest *IssueRequest, channel string) IssueResponse {
	request, err := getCreateIssueRequest(env, issueRequest, channel)
	var issueResponse IssueResponse
	fatal(err)
	if err != nil {
		setIssueResponseError(&issueResponse, errors.New("Failed to perform getInstallationIDRequest"))
	}
	body, err2 := processRequest(request)
	fatal(err2)
	if err2 != nil {
		setIssueResponseError(&issueResponse, errors.New("Failed to perform processRequest(request)"))
	}
	err3 := json.NewDecoder(bytes.NewReader(body)).Decode(&issueResponse)
	fatal(err3)
	if err3 != nil {
		setIssueResponseError(&issueResponse, errors.New("Failed to perform json.NewDecoder(response.Body)"))
	}
	return issueResponse
}

func getCreateIssueRequest(env Environment, issueRequest *IssueRequest, channel string) (*http.Request, error) {
	token, err := getInstallationToken(environment)
	fatal(err)
	if err != nil {
		return nil, errors.New("Failed to get installation token")
	}
	issueRequest.Labels = append(issueRequest.Labels, "From-"+channel)
	postBody, _ := json.Marshal(issueRequest)
	log.Printf("Create request for new issue using installation token %v and request body %v", token, string(postBody))
	requestBody := bytes.NewBuffer(postBody)
	request, err3 := http.NewRequest("POST", GithubIssueURI, requestBody)
	fatal(err3)
	if err3 != nil {
		return nil, errors.New("Failed to perform http.NewRequest")
	}
	getJwtHeaders(request, token)
	return request, nil
}

func setIssueResponseError(issueResponse *IssueResponse, err error) {
	issueResponse.Message = err.Error()
}
