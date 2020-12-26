package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

// Environment values, set these in your deployment
type Environment struct {
	repoOwner           string // add GITHUB_REPO_OWNER in deployment variables
	repoName            string // add GITHUB_REPO_NAME in deployment variables
	githubAppIdentifier string // add GITHUB_APP_IDENTIFIER in deployment variables
	githubAppPrivateKey []byte // add GITHUB_APP_PRIVATE_KEY in base64 in deployment variables
	apiToken            string // add API_TOKEN in deployment variables
}

// IssueResponse created by github api
type IssueResponse struct {
	Number  int    `json:"number,omitempty"`
	Message string `json:"message"`
	Errors  []struct {
		Value    interface{} `json:"value"`
		Resource string      `json:"resource"`
		Field    string      `json:"field"`
		Code     string      `json:"code"`
	} `json:"errors"`
}

// IssueRequest github issue request
type IssueRequest struct {
	Title     string        `json:"title"`
	Body      string        `json:"body,omitempty"`
	Assignees []string      `json:"assignees,omitempty"`
	Milestone int           `json:"milestone,omitempty"`
	Labels    []interface{} `json:"labels,omitempty"`
}

var (
	environment Environment
	client      = &http.Client{
		Timeout: time.Second * 10,
	}

	// GithubInstallationIDURI URL for fetching installation ID of GitHub app
	GithubInstallationIDURI string
	// GithubIssueURI URL for creating GitHub issue
	GithubIssueURI string
)

func init() {
	environment.repoOwner = os.Getenv("GITHUB_REPO_OWNER")
	environment.repoName = os.Getenv("GITHUB_REPO_NAME")
	environment.githubAppIdentifier = os.Getenv("GITHUB_APP_IDENTIFIER")
	environment.githubAppPrivateKey = os.Getenv("GITHUB_APP_PRIVATE_KEY")
	environment.apiToken = os.Getenv("API_TOKEN")
	envPrivateKey := os.Getenv("GITHUB_APP_PRIVATE_KEY")
	environment.githubAppPrivateKey, _ = base64.StdEncoding.DecodeString(envPrivateKey)
	GithubInstallationIDURI = fmt.Sprintf(GithubAPIBase+"/repos/%v/%v/installation", environment.repoOwner, environment.repoName)
	GithubIssueURI = fmt.Sprintf(GithubAPIBase+"/repos/%v/%v/issues", environment.repoOwner, environment.repoName)
}

// CreateIssue main function responsible for creating GitHub issue
func CreateIssue(w http.ResponseWriter, r *http.Request) {
	validRequest, issueRequest, channel := isRequestValid(r)
	if !validRequest {
		log.Printf("Invalid request with url %v and request %v", r.URL, r)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	issueResponse := createGithubIssue(environment, &issueRequest, channel)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(issueResponse)
}

func isRequestValid(r *http.Request) (bool, IssueRequest, string) {
	var issueRequest IssueRequest
	body, _ := ioutil.ReadAll(r.Body)
	log.Printf("Processing request for isRequestValid %v", string(body))
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&issueRequest)
	fatal(err)

	params := r.URL.Query()
	if environment.apiToken != params.Get("token") {
		log.Printf("Invalid request: token doesn't match env %v", environment.apiToken)
		return false, IssueRequest{}, ""
	}

	if params.Get("channel") == "" {
		log.Printf("Invalid request: channel param not present")
		return false, IssueRequest{}, ""
	}

	if issueRequest.Title == "" {
		log.Printf("Crash message empty issue title")
		return false, IssueRequest{}, ""
	}
	channel := params.Get("channel")
	return true, issueRequest, channel
}

func main() {
	http.HandleFunc("/", CreateIssue)
	log.Fatal(http.ListenAndServe(":10000", nil))
}

func fatal(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func processRequest(request *http.Request) ([]byte, error) {
	log.Printf("Final request %v", request)
	resp, err2 := client.Do(request)
	fatal(err2)
	if err2 != nil {
		return nil, errors.New("failed to perform client.Do(request)")
	}
	body, err4 := ioutil.ReadAll(resp.Body)
	fatal(err4)
	if err4 != nil {
		return nil, errors.New("failed to perform ioutil.ReadAll(resp.Body)")
	}
	log.Printf("Response from processRequest is %v", string(body))
	defer resp.Body.Close()
	return body, nil
}
