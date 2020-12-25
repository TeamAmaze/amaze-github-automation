package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Environment values, set these in your deployment
type Environment struct {
	repoOwner           string // add GITHUB_REPO_OWNER in deployment variables
	repoName            string // add GITHUB_REPO_NAME in deployment variables
	githubAppIdentifier string // add GITHUB_APP_IDENTIFIER in deployment variables
	githubAppPrivateKey string // add GITHUB_APP_PRIVATE_KEY in deployment variables
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
	Labels    []interface{} `json:"labels,omitempty`
}

var (
	environment   Environment
	issueResponse IssueResponse
	client        = &http.Client{
		Timeout: time.Second * 10,
	}
	issueRequest IssueRequest

	// GithubInstallationIDURI URL for fetching installation ID of GitHub app
	GithubInstallationIDURI string
	// GithubIssueURI URL for creating GitHub issue
	GithubIssueURI string

	// query params
	channel string // adds label From - `channel` to github issue
)

func init() {
	// environment.repoOwner = os.Getenv("GITHUB_REPO_OWNER")
	// environment.repoName = os.Getenv("GITHUB_REPO_NAME")
	// environment.githubAppIdentifier = os.Getenv("GITHUB_APP_IDENTIFIER")
	// environment.githubAppPrivateKey = os.Getenv("GITHUB_APP_PRIVATE_KEY")
	// environment.apiToken = os.Getenv("API_TOKEN")

	environment.apiToken = "test"
	environment.repoOwner = "TeamAmaze"
	environment.repoName = "amaze-telegram-bot"
	environment.githubAppIdentifier = "93675"
	environment.githubAppPrivateKey = "-----BEGIN RSA PRIVATE KEY-----\n" +
		"MIIEowIBAAKCAQEAxIuIPL+orDmJGhXdsTmjsrwNJN5oMl8qhLSclH9OVmdwLkZm\n" +
		"KkvxvRPTKJWGLlJ3yimWA6qy8R1UOXlD0rHXQM+uesABKGQrgBiR8Rxrp9qSlQJ5\n" +
		"A3ck1+twfz5nM296gECCu5ZRxRj2D+8/LM/IhO2AWk9CrI1sX3c1cBdFEIuBSM0M\n" +
		"+WhVPg9FSgFddFYjcNcAeYhy/NbgDGjj2usqH934xgkQfBGdiQbKFPqWGFV2Q+mP\n" +
		"6/S+w+cKvGPfXiSM28PVHInUh3vkKUWWLCdjnR5e5Ta/l9Ok0HAsjiQkej8y5x0W\n" +
		"ZrecRGmZC4ZsFrzggKBK7mEkZIOF04qgwCtZFQIDAQABAoIBAQCfWf3AOyg3UoKt\n" +
		"KpNOkEv/quYBQW07gdsIMyNMZpcOCNl0O1Gz81TwlrU6D1j2D5jdyK+/E1P3l27l\n" +
		"FkN9/QBnpLpy/V8y71wxhDo3QXKradQ0igexXpT5lwLjt6WWl0i72RHlo29ynNVL\n" +
		"gA85dtG9rI3HKsIFArieAhnKYqN1UCoD9cQVEFCVzgVZtCT/wkaLT91F73rPQo7T\n" +
		"JcDAYUAECkjTngBzJoHehnCyqs7wcS6uJVPLFy0Mc2py3uYer7kAu3Nrf9uhG4qm\n" +
		"0MXByQ83PEQ+OFj2U1OEoycx1G4+5fHZ/jm8LYXAR+Y2TYxivMFr4dFadMkEdK9n\n" +
		"iuCk9Wb1AoGBAOsj29fd9ZljtpsmdV2h4fcPVs9StwLQRuGrJQS8rIDlYPf8AalG\n" +
		"jhFOIXNwRehiPVNQZT5tgmsr2Bo2AQKaNfM6M9vIFtVml+PNA/u+N30nGTIuJMbf\n" +
		"QQkO6yHUbCmltgfSqRaFnMlvgP0X9BIhS/spyzEZJ/Cm/9sNs2/gQmq3AoGBANX7\n" +
		"KYxHUrJDwmrI2h5R6RQiX9ssC/erapxtpPZzX14z1VNlzD0jZNVdqk7qOJcXCPWh\n" +
		"4JJ06rnBP9KXAGJajilQSqXxDeMeBezuzQCIaZYbO9//QRNRRTWXmcAHUCJSM1uJ\n" +
		"7s4B0m9I6oY7Nxs7oLLOSWUHlZ82X4nHU/Wojn6TAoGAOK9pVS3eAj9mixqHWq4m\n" +
		"4j9hZxOCqPv6ynZOs0iksWIasU2gPOWUZBmYuNKNF8tvC0GrVpRhx2JHc3InZjA0\n" +
		"51DVpZsj3gggf7sxxaOCjvo4+b7kAMlbTUq6ZmpmNNgM/O/M8W/+bxUhXGJE5YX/\n" +
		"YioeINT2qu4nafBwnHzMphsCgYAjHz+Zk9diBTczGdabZWxxbpb3PYqVU2CDXofW\n" +
		"H+fGaZGZR7s3ScjyMJaUr2MsgY5p6vEWePRSGwMjyL86ZYyyAUjPZfqWjcYBNs0V\n" +
		"Sk6yYbP5N0dyKUPH4SNOXqTrjTx6yPAWhjwJIhnEgJGx+Z6N2sg3OgB4Co+x6LLC\n" +
		"PrFs2wKBgEB0m9HMfkFVW9WvpvZKAK8I+ZEcs3iv38ngV4YagE5xO7eY04fpIryn\n" +
		"VFEtxQj/1BVCx4zXvDb7750AwFA34fW1wMIicjzvLsvzQfQdaoiJFAX+wUb6yOSf\n" +
		"F9vbjorfyLVMe73mkoA0Nq/ExbZbho34mYZZQgT8RJp5mYETF/A0\n" +
		"-----END RSA PRIVATE KEY-----"

	GithubInstallationIDURI = fmt.Sprintf(GithubAPIBase+"/repos/%v/%v/installation", environment.repoOwner, environment.repoName)
	GithubIssueURI = fmt.Sprintf(GithubAPIBase+"/repos/%v/%v/issues", environment.repoOwner, environment.repoName)
}

func createIssue(w http.ResponseWriter, r *http.Request) {
	if !isRequestValid(r) {
		log.Printf("Invalid request with url %v and request %v", r.URL, r)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	issueResponse := createGithubIssue(environment, &issueRequest, channel)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(issueResponse)
}

func isRequestValid(r *http.Request) bool {
	body, _ := ioutil.ReadAll(r.Body)
	log.Printf("Processing request for isRequestValid %v", string(body))
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&issueRequest)
	fatal(err)
	defer r.Body.Close()

	params := r.URL.Query()
	if environment.apiToken != params.Get("token") {
		log.Printf("Invalid request: token doesn't match env %v", environment.apiToken)
		return false
	}

	if params.Get("channel") == null {
		log.Printf("Invalid request: channel param not present")
		return false
	}

	if issueRequest.Title == "" {
		log.Printf("Crash message empty issue title")
		return false
	}
	channel = params.Get("channel")
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
