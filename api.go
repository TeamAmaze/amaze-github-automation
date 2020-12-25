package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// Environment values, set these in your deployment
type Environment struct {
	repoOwner           string // add GITHUB_REPO_OWNER in deployment variables
	repoName            string // add GITHUB_REPO_NAME in deployment variables
	githubAppIdentifier string // add GITHUB_APP_IDENTIFIER in deployment variables
	githubAppPrivateKey string // add GITHUB_APP_PRIVATE_KEY in deployment variables
	apiToken            string // add API_TOKEN in deployment variables
}

var req struct {
	IssueTitle     string   `json:"title"`
	IssueBody      string   `json:"body"`
	IssueAssignee  string   `json:"assignee"`
	IssueMilestone int      `json:"milestone"`
	IssueLabels    []string `json:"labels`
	IssueAssignees []string `json:"assignees"`

	// query params
	channel string // adds label From - `channel` to github issue
}

var environment Environment

func init() {
	environment.repoOwner = os.Getenv("GITHUB_REPO_OWNER")
	environment.repoName = os.Getenv("GITHUB_REPO_NAME")
	environment.githubAppIdentifier = os.Getenv("GITHUB_APP_IDENTIFIER")
	environment.githubAppPrivateKey = os.Getenv("GITHUB_APP_PRIVATE_KEY")
	environment.apiToken = os.Getenv("API_TOKEN")
}

func createIssue(w http.ResponseWriter, r *http.Request) {
	if !isRequestValid(r) {
		log.Printf("Invalid request with url %v and request %v", r.URL, req)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	token, err := getInstallationToken(environment)
	fatal(err)
	fmt.Fprint(w, token)
}

func isRequestValid(r *http.Request) bool {
	err := json.NewDecoder(r.Body).Decode(&req)
	fatal(err)

	params := r.URL.Query()
	if environment.apiToken != params.Get("token") {
		log.Printf("Invalid request: token doesn't match env %v", environment.apiToken)
		return false
	}

	if req.IssueTitle == "" {
		log.Printf("Crash message empty issue title")
		return false
	}

	return true
}

func main() {
	http.HandleFunc("/", createIssue)
	log.Fatal(http.ListenAndServe(":10000", nil))
}

func fatal(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
