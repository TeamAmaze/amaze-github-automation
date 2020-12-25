package main

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

var (
	signKey *rsa.PrivateKey
	client  = &http.Client{
		Timeout: time.Second * 10,
	}
)

var installationTokenResponse struct {
	Token string `json:"token"`
}

var installationIDResponse struct {
	ID int `json:"id"`
}

func getInstallationToken(env Environment) (string, error) {
	request, err1 := getInstallationTokenRequest(env)
	if err1 != nil {
		return "", errors.New("failed to getInstallationTokenRequest")
	}
	log.Printf("Final request for getInstallationToken %v", request)
	resp, err2 := client.Do(request)
	fatal(err2)
	if err2 != nil {
		return "", errors.New("failed to perform client.Do(request)")
	}
	body, err4 := ioutil.ReadAll(resp.Body)
	fatal(err4)
	if err4 != nil {
		return "", errors.New("failed to perform ioutil.ReadAll(resp.Body)")
	}
	log.Printf("Response from getInstallationToken is %v", string(body))
	err3 := json.NewDecoder(bytes.NewReader(body)).Decode(&installationTokenResponse)
	fatal(err3)
	if err3 != nil {
		return "", errors.New("failed to perform json.NewDecoder(bytes.NewReader(body)).Decode(&installationTokenResponse)")
	}
	defer resp.Body.Close()
	log.Printf("Found installation token response %v", installationTokenResponse)
	if &installationTokenResponse.Token == nil {
		return "", errors.New("failed to get installation token response")
	}
	return installationTokenResponse.Token, nil
}

func getInstallationID(env Environment) (int, error) {
	githubInstallationIDURI := fmt.Sprintf(GithubApiBase+"/repos/%v/%v/installation", env.repoOwner, env.repoName)
	request, err := getInstallationIDRequest(githubInstallationIDURI, env)
	fatal(err)
	if err != nil {
		return 0, errors.New("Failed to perform getInstallationIDRequest")
	}
	log.Printf("Create request for getInstallationID %v", request)
	response, err1 := client.Do(request)
	fatal(err1)
	if err1 != nil {
		return 0, errors.New("Failed to perform client.Do(request)")
	}
	body, err4 := ioutil.ReadAll(response.Body)
	fatal(err4)
	if err4 != nil {
		return 0, errors.New("failed to perform ioutil.ReadAll(response.Body)")
	}
	log.Printf("Response from getInstallationID is %v", string(body))
	err2 := json.NewDecoder(bytes.NewReader(body)).Decode(&installationIDResponse)
	fatal(err2)
	if err2 != nil {
		return 0, errors.New("Failed to perform json.NewDecoder(response.Body)")
	}
	defer response.Body.Close()
	log.Printf("Found installation id response %v from uri %v", installationIDResponse, githubInstallationIDURI)
	if &installationIDResponse.ID == nil {
		return 0, errors.New("failed to get installation token response")
	}
	return installationIDResponse.ID, nil
}

func getInstallationIDRequest(uri string, env Environment) (*http.Request, error) {
	jwt, err1 := getJwt(env.githubAppPrivateKey, env.githubAppIdentifier)
	if err1 != nil {
		return nil, errors.New("Failed to perform getJwt")
	}
	req, err := http.NewRequest("GET", uri, nil)
	fatal(err)
	if err != nil {
		return nil, errors.New("Failed to perform getInstallationIDRequest")
	}
	getJwtHeaders(req, jwt)
	return req, nil
}

func getInstallationTokenRequest(env Environment) (*http.Request, error) {
	installaionID, err1 := getInstallationID(env)
	if err1 != nil {
		return nil, errors.New("Failed to perform getInstallationTokenRequest")
	}
	githubInstallationTokenURI := fmt.Sprintf(GithubApiBase+"/app/installations/%v/access_tokens", installaionID)
	log.Printf("Create request for getInstallationTokenRequest %v", githubInstallationTokenURI)
	postBody, err2 := json.Marshal(map[string][]string{
		"repositories": []string{env.repoName},
	})
	fatal(err2)
	if err2 != nil {
		return nil, errors.New("Failed to perform json.Marshal")
	}
	requestBody := bytes.NewBuffer(postBody)
	jwt, err := getJwt(env.githubAppPrivateKey, env.githubAppIdentifier)
	if err != nil {
		return nil, errors.New("Failed to perform getJwt")
	}
	req, err3 := http.NewRequest("POST", githubInstallationTokenURI, requestBody)
	fatal(err3)
	if err3 != nil {
		return nil, errors.New("Failed to perform http.NewRequest")
	}
	getJwtHeaders(req, jwt)
	return req, nil
}

func getGithubBaseHeaders(req *http.Request) {
	req.Header.Add("Accept", "application/vnd.github.v3+json")
}

func getJwtHeaders(req *http.Request, jwt string) {
	getGithubBaseHeaders(req)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", jwt))
}

func initSignKey(privateKey string) {
	signKey, _ = jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
}

func getJwt(privateKey string, appIdentifier string) (string, error) {
	initSignKey(privateKey)
	t := jwt.New(jwt.GetSigningMethod("RS256"))
	claims := t.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Minute * 1).Unix()
	claims["iat"] = time.Now().Unix()
	claims["iss"] = appIdentifier
	tokenString, err := t.SignedString(signKey)
	fatal(err)
	if err != nil {
		return "", errors.New("Failed to perform t.SignedString(signKey)")
	}
	return tokenString, nil
}
