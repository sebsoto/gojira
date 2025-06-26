package git

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v72/github"

	"github.com/sebsoto/gojira/pkg/semver"
)

type FilterFunction func(*github.Commit) bool

type Commit struct {
	Message string
	SHA     string
}

type Repo interface {
	GetTags() ([]Tag, error)
	ListCommits(string, string, FilterFunction) ([]Commit, error)
	MergeBase(string, string) (string, error)
}

type Tag struct {
	Name string
	Sha  string
}

type GithubRepo struct {
	owner  string
	name   string
	client *github.Client
}

func NewRepo(gitURL string) (Repo, error) {
	u, err := url.Parse(gitURL)
	if err != nil {
		return nil, err
	}
	if u.Host != "github.com" {
		return nil, fmt.Errorf("unsupported git provider: %s", u.Host)
	}
	urlSplit := strings.Split(u.Path, "/")
	if len(urlSplit) < 3 {
		return nil, fmt.Errorf("unexpected URL path: %s", gitURL)
	}
	return NewGithubRepo(urlSplit[1], urlSplit[2]), nil
}

func NewGithubRepo(owner string, name string) *GithubRepo {
	client := github.NewClient(nil)
	token, err := getGithubAPIToken()
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: unable to access github api token: %s", err)
	} else {
		client = client.WithAuthToken(token)
	}
	return &GithubRepo{
		owner:  owner,
		name:   name,
		client: client,
	}
}

func (r *GithubRepo) GetTags() ([]Tag, error) {
	tags, _, err := r.client.Repositories.ListTags(context.Background(), r.owner, r.name, &github.ListOptions{
		PerPage: 100,
	})
	if err != nil {
		return nil, err
	}
	tagList := make([]Tag, 0)
	for _, tag := range tags {
		tagList = append(tagList, Tag{Name: tag.GetName(), Sha: tag.GetCommit().GetSHA()})
	}
	return tagList, nil
}

func (r *GithubRepo) ListCommits(startSHA, endSHA string, filter FilterFunction) ([]Commit, error) {
	var commitList []Commit
	commits, resp, err := r.client.Repositories.ListCommits(context.Background(), r.owner, r.name, &github.CommitsListOptions{
		SHA: startSHA,
		ListOptions: github.ListOptions{
			Page:    0,
			PerPage: 100,
		},
	})
	if err != nil {
		return nil, err
	}
	for _, commit := range commits {
		if commit.GetSHA() == endSHA {
			return commitList, nil
		}
		if !filter(commit.GetCommit()) {
			commitList = append(commitList, Commit{
				Message: commit.GetCommit().GetMessage(),
				SHA:     commit.GetSHA(),
			})
		}
	}
	for nextPage := resp.NextPage; nextPage != resp.LastPage; nextPage = resp.NextPage {
		fmt.Fprintf(os.Stderr, "checking page: %d/%d\n", resp.NextPage, resp.LastPage)
		commits, resp, err = r.client.Repositories.ListCommits(context.Background(), r.owner, r.name, &github.CommitsListOptions{
			SHA: startSHA,
			ListOptions: github.ListOptions{
				Page:    nextPage,
				PerPage: 100,
			},
		})
		if err != nil {
			return nil, err
		}
		for _, commit := range commits {
			if commit.GetSHA() == endSHA {
				return commitList, nil
			}
			if !filter(commit.GetCommit()) {
				commitList = append(commitList, Commit{
					Message: commit.GetCommit().GetMessage(),
					SHA:     commit.GetSHA(),
				})
			}
		}
	}
	return commitList, nil
}

func (r *GithubRepo) MergeBase(sha1, sha2 string) (string, error) {
	comparison, _, err := r.client.Repositories.CompareCommits(context.Background(), r.owner, r.name, sha1, sha2, nil)
	if err != nil {
		return "", err
	}
	return comparison.MergeBaseCommit.GetSHA(), nil
}

// FindPreviousTag returns the commit of the previous tag
func FindPreviousTag(repo Repo, currentTag semver.Semver) (string, error) {
	tags, err := repo.GetTags()
	if err != nil {
		return "", err
	}
	if currentTag.Patch != 0 {
		prevTagName := fmt.Sprintf("v%d.%d.%d", currentTag.Major, currentTag.Minor, currentTag.Patch-1)
		for _, tag := range tags {
			if tag.Name == prevTagName {
				fmt.Printf("Tag %s found\nCommit %s\n", tag.Name, tag.Sha)
				return tag.Sha, nil
			}
			return "", fmt.Errorf("no previous tag found")
		}
	}
	var prevTag Tag
	var prevTagSemver semver.Semver
	for _, tag := range tags {
		tagSemver, err := semver.New(tag.Name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: unable to parse semver %s: %s\n", tag.Name, err)
			continue
		}
		if tagSemver.Major == currentTag.Major && tagSemver.Minor == currentTag.Minor-1 && tagSemver.Patch > prevTagSemver.Patch {
			prevTagSemver = *tagSemver
			prevTag = tag
			break
		}
	}
	return prevTag.Sha, nil
}

func IsMerge(commit *github.Commit) bool {
	return len(commit.Parents) > 1
}

func getGithubAPIToken() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	creds, err := os.ReadFile(filepath.Join(homedir, ".github/token"))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(creds)), nil
}
