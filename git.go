package main

import (
	"fmt"
)

var GithubApiBase = "https://api.github.com"
var GITHUB_API_BASE_URI = "https://api.github.com/repos/%v/%v/issues"
var AMAZE_OWNER = "VishalNehra"
var AMAZE_REPO = "Green_Day_Lyrics"
var GITHUB_API_CREATE_ISSUE = fmt.Sprint(GITHUB_API_BASE_URI, AMAZE_OWNER, AMAZE_REPO)

func helloWorld() {
	fmt.Printf("jwt.getJwt()")
}
