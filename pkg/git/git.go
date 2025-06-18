package git

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/go-github/v72/github"
)

type FilterFunction func(*github.Commit) bool

type Commit interface {
	GetMessage() string
	GetSHA() string
}

type Repo interface {
	GetTags() ([]Tag, error)
	ListCommits(string, string, FilterFunction) ([]Commit, error)
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
	return &GithubRepo{
		owner:  owner,
		name:   name,
		client: github.NewClient(nil),
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
			commitList = append(commitList, commit.GetCommit())
		}
	}
	for nextPage := resp.NextPage; nextPage != resp.LastPage; {
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
				commitList = append(commitList, commit.GetCommit())
			}
		}
	}
	return commitList, nil
}

func FindReleaseTail(repo Repo, semver string) (string, error) {
	tags, err := repo.GetTags()
	if err != nil {
		return "", err
	}
	tagMap := make(map[string]string)
	for _, tag := range tags {
		tagMap[tag.Sha] = tag.Name
	}
	prevVersion := ""
	semverSplit := strings.Split(semver, ".")
	if len(semverSplit) != 3 {
		return "", fmt.Errorf("expected a semver of format vX.Y.Z")
	}
	if strings.HasSuffix(semver, ".0") {
		// Use last minor release instead
		minorVersion, err := strconv.Atoi(semverSplit[1])
		if err != nil {
			return "", err
		}
		prevVersion = fmt.Sprintf("%s.%d.0", semverSplit[0], minorVersion-1)
	} else {
		patchVersion, err := strconv.Atoi(semverSplit[2])
		if err != nil {
			return "", err
		}
		prevVersion = fmt.Sprintf("%s.%s.%d", semverSplit[0], semverSplit[1], patchVersion-1)
	}
	fmt.Println("looking for " + prevVersion)
	for _, tag := range tags {
		if tag.Name == prevVersion {
			fmt.Printf("Tag %s found\nCommit %s\n", tag.Name, tag.Sha)
			return tag.Sha, nil
		}
	}
	return "", fmt.Errorf("no valid tag found")
}

func IsMerge(commit *github.Commit) bool {
	return len(commit.Parents) > 1
}
