package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type IssueTypeName string

const (
	EpicIssue IssueTypeName = "Epic"
	TaskIssue IssueTypeName = "Task"
)

type IssuePriorityName string

const MajorPriority IssuePriorityName = "Major"

type IssueSearch struct {
	Issues []Issue `json:"issues"`
}

type Issue struct {
	Fields IssueFields `json:"fields"`
	Key    string      `json:"key,omitempty"`
}

type IssueFields struct {
	Summary       string          `json:"summary"`
	Description   string          `json:"description"`
	Project       Project         `json:"project"`
	IssueType     IssueType       `json:"issuetype"`
	FixVersions   []FixVersion    `json:"fixVersions,omitempty"`
	TargetVersion []TargetVersion `json:"customfield_12319940,omitempty"`
	StartDate     string          `json:"customfield_12313941,omitempty"`
	EndDate       string          `json:"customfield_12313942,omitempty"`
	Security      *Security       `json:"security,omitempty"`
	EpicName      string          `json:"customfield_12311141,omitempty"`
	EpicLink      string          `json:"customfield_12311140,omitempty"`
	Labels        []string        `json:"labels,omitempty"`
	Priority      *Priority       `json:"priority,omitempty"`
}

type Priority struct {
	Name IssuePriorityName `json:"name"`
}

type FixVersion struct {
	Name string `json:"name"`
}
type TargetVersion struct {
	Name string `json:"version"`
}

type Project struct {
	ID  *string `json:"id,omitempty"`
	Key *string `json:"key,omitempty"`
}

type Security struct {
	Name string `json:"name"`
}

type IssueType struct {
	Name IssueTypeName `json:"name"`
}

type IssueCreationResponse struct {
	Id  string `json:"id"`
	Key string `json:"key"`
}

func Search(query string) ([]Issue, error) {
	searchURL, err := constructURL("/search", url.Values{"jql": []string{query}})
	if err != nil {
		return nil, err
	}
	body, err := apiRequest(http.MethodGet, searchURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var searchResults IssueSearch
	if err = json.Unmarshal(body, &searchResults); err != nil {
		return nil, err
	}
	return searchResults.Issues, nil
}

type remoteLink struct {
	Object remoteLinkDetails `json:"object"`
}

type remoteLinkDetails struct {
	Summary string `json:"summary,omitempty"`
	Title   string `json:"title"`
	URL     string `json:"url"`
}

func GetRemoteLinks(issueKey string) ([]remoteLink, error) {
	remoteLinkURL, err := constructURL("/issue/"+issueKey+"/remotelink", nil)
	if err != nil {
		return nil, err
	}
	body, err := apiRequest(http.MethodGet, remoteLinkURL.String(), nil)
	if err != nil {
		return nil, err
	}
	links := new([]remoteLink)
	err = json.Unmarshal(body, &links)
	if err != nil {
		return nil, err
	}
	return *links, nil
}

func AddRemoteLink(issueKey, linkURL, title string) error {
	remoteLinkURL, err := constructURL("/issue/"+issueKey+"/remotelink", nil)
	if err != nil {
		return err
	}
	link := remoteLink{
		Object: remoteLinkDetails{
			Title: title,
			URL:   linkURL,
		},
	}
	reqBody, err := json.Marshal(&link)
	if err != nil {
		return err
	}
	body, err := apiRequest(http.MethodPost, remoteLinkURL.String(), reqBody)
	if err != nil {
		return err
	}
	fmt.Println(string(body))
	return nil

}

func constructURL(endpoint string, queries url.Values) (*url.URL, error) {
	apiURL := "https://issues.redhat.com/rest/api/2"
	joined, err := url.JoinPath(apiURL, endpoint)
	if err != nil {
		return nil, err
	}
	searchURL, err := url.Parse(joined)
	if err != nil {
		return nil, err
	}
	if queries != nil {
		searchURL.RawQuery = queries.Encode()
	}
	return searchURL, nil
}

func getJIRAAPIToken() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	creds, err := os.ReadFile(filepath.Join(homedir, ".jira/token"))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(creds)), nil
}

func CreateIssue(issue *Issue) (*IssueCreationResponse, error) {
	createIssueURL, err := constructURL("/issue", nil)
	if err != nil {
		return nil, err
	}
	issueBytes, err := json.Marshal(issue)
	if err != nil {
		return nil, err
	}
	res, err := apiRequest(http.MethodPost, createIssueURL.String(), issueBytes)
	if err != nil {
		return nil, err
	}
	var response IssueCreationResponse
	err = json.Unmarshal(res, &response)
	return &response, err
}

func UpdateIssue(key, updateBody string) error {
	updateIssueURL, err := constructURL("/issue/"+key, nil)
	if err != nil {
		return err
	}
	_, err = apiRequest(http.MethodPut, updateIssueURL.String(), []byte(updateBody))
	if err != nil {
		return err
	}
	return nil
}

func GetIssue(issueKey string) (*Issue, error) {
	results, err := Search("key = " + issueKey)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("issue not found")
	}
	if len(results) > 1 {
		return nil, fmt.Errorf("unexpected issues found")
	}
	return &results[0], nil
}

// apiRequest makes the request, returns the response body
func apiRequest(httpMethod, url string, body []byte) ([]byte, error) {
	apiKey, err := getJIRAAPIToken()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(httpMethod, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+apiKey)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode > 299 {
		return nil, fmt.Errorf("%d: %s: %s", res.StatusCode, res.Status, string(resBody))
	}
	return resBody, nil
}
