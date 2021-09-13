# amaze-github-automation
GitHub issue automation api function

API responsible to be used along with this [GitHub app](https://github.com/marketplace/issues-rest-api-function)

Want to automate GitHub issue creation? Use this app and get away with the burden of implementing GitHub authentication APIs

Steps:
---
- Add [this](https://github.com/apps/amaze-issue-automation) app to the repository and give it Issue read / write and metadata read permission.
- Generate the private key for your installation. Follow [this](https://docs.github.com/en/free-pro-team@latest/developers/apps/authenticating-with-github-apps) article for the same
- Deploy this API function on your preferred hosting.

Deployment: 
---
- Make sure to remove/modify the API endpoint in `api.go` Modify `http.HandleFunc("/", createIssue)` to your requirements.
- Set following environment variables before deployment
    - GITHUB_REPO_OWNER
    - GITHUB_REPO_NAME
    - GITHUB_APP_IDENTIFIER
    - GITHUB_APP_PRIVATE_KEY (Encode generated pem file above to base64)
    - API_TOKEN (see below)

Endpoint:
---
Endpoint supports following format:
- POST with request body:
```
{
    "title": "Dummy title",  // mandatory
    "body": "Dummy body",
    "milestone": 15,
    "assignees": ["GitHub Usernames"]
    "labels": ["dummy"]
}
```
- `token` as  query param. Compares with `API_TOKEN` added in environment variables. Both should match
- `channel` as query param. Creates and assigns a label for the GitHub issue being created in format `From-channel` format.
