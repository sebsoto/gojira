package konflux

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"text/tabwriter"

	applicationv1alpha1 "github.com/konflux-ci/application-api/api/v1alpha1"
	releasev1alpha1 "github.com/konflux-ci/release-service/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	clientconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/yaml"

	"github.com/google/go-github/v72/github"
	"github.com/sebsoto/gojira/pkg/git"
	"github.com/sebsoto/gojira/pkg/jira"
)

type releaseData struct {
	ReleaseNotes releaseNotes `json:"releaseNotes"`
}
type releaseNotes struct {
	Type   string `json:"type"`
	CVEs   []cve  `json:"cves"`
	Issues issues `json:"issues"`
}
type cve struct {
	Component string `json:"component"`
	Key       string `json:"key"`
}

type issues struct {
	Fixed []issue `json:"fixed"`
}
type issue struct {
	ID     string `json:"id"`
	Source string `json:"source"`
}

func NewRelease(releaseplan, version string, jiraProjects []string, branchCommit string) (*releasev1alpha1.Release, error) {
	config, err := clientconfig.GetConfig()
	if err != nil {
		return nil, err
	}
	c, err := client.New(config, client.Options{})
	if err != nil {
		return nil, err
	}
	releasev1alpha1.AddToScheme(c.Scheme())
	applicationv1alpha1.AddToScheme(c.Scheme())
	var rp releasev1alpha1.ReleasePlan
	err = c.Get(context.Background(), types.NamespacedName{Name: releaseplan, Namespace: "windows-machine-conf-tenant"}, &rp)
	if err != nil {
		return nil, err
	}
	var relList releasev1alpha1.ReleaseList
	err = c.List(context.Background(), &relList, client.MatchingLabels{"appstudio.openshift.io/application": rp.Spec.Application}, client.InNamespace(rp.GetNamespace()))
	if err != nil {
		return nil, err
	}
	lastRelease, err := latestRelease(relList.Items)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Latest release is: %s\n", lastRelease.GetName())
	fmt.Printf("Uses snapshot: %s\n", lastRelease.Spec.Snapshot)
	var snap applicationv1alpha1.Snapshot
	err = c.Get(context.Background(), types.NamespacedName{Name: lastRelease.Spec.Snapshot, Namespace: "windows-machine-conf-tenant"}, &snap)
	if err != nil {
		return nil, err
	}
	var gitURL, snapshotCommit string
	componentName := snap.Labels["appstudio.openshift.io/component"]
	for _, component := range snap.Spec.Components {
		if component.Name == componentName {
			gitURL = component.Source.GitSource.URL
			snapshotCommit = component.Source.GitSource.Revision
		}
	}
	fmt.Printf("Snapshot timestamp: %v\n", snap.GetCreationTimestamp())
	fmt.Printf("Snapshot commit: %v\n", snapshotCommit)
	var component applicationv1alpha1.Component
	err = c.Get(context.Background(), types.NamespacedName{Name: componentName, Namespace: "windows-machine-conf-tenant"}, &component)
	if err != nil {
		return nil, err
	}

	repo, err := git.NewRepo(gitURL)
	if err != nil {
		return nil, err
	}

	mergesSinceRelease, err := repo.ListCommits(component.Spec.Source.GitSource.Revision, snapshotCommit, git.IsMerge)
	if err != nil {
		return nil, err
	}
	fmt.Printf("-----\n\n")
	fmt.Printf("%d recent merges not included in release:\n", len(mergesSinceRelease))
	for i, mergeCommit := range mergesSinceRelease {
		fmt.Printf("%d: %s\n", i+1, mergeCommit.GetMessage())
	}
	fmt.Printf("-----\n\n")

	fromSHA := branchCommit
	if branchCommit == "" {
		fromSHA, err = git.FindReleaseTail(repo, version)
		if err != nil {
			return nil, err
		}
	}
	commits, err := repo.ListCommits(snapshotCommit, fromSHA, git.IsMerge)
	if err != nil {
		return nil, err
	}
	jiraTickets, err := getJiraIssues(jiraProjects, commits)
	if err != nil {
		return nil, err
	}

	// Print contents
	fmt.Printf("Jira issues included in this release:\n")
	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "Issue\tSummary\tFix Version")
	fmt.Fprintln(w, "___\t___\t___")
	for _, ticket := range jiraTickets {
		fmt.Fprintf(w, "%s\t%s\t%s\n", ticket.Key, ticket.Fields.Summary, ticket.Fields.FixVersions)
	}
	w.Flush()
	fmt.Printf("-----\n\n")

	r, err := newRelease(componentName, jiraTickets, rp.GetName(), snap.GetName())
	if err != nil {
		return nil, err
	}

	yamlNotes, err := yaml.Marshal(r)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(yamlNotes))

	return r, nil
}

func latestRelease(relList []releasev1alpha1.Release) (*releasev1alpha1.Release, error) {
	// grab recent release
	successfulReleases := slices.DeleteFunc(relList, func(a releasev1alpha1.Release) bool {
		for _, condition := range a.Status.Conditions {
			if condition.Type == "Released" && condition.Status == "True" {
				return false
			}
		}
		return true
	})
	if len(successfulReleases) == 0 {
		return nil, fmt.Errorf("no sucessful releases found")
	}
	latestRelease := slices.MaxFunc(successfulReleases, func(a, b releasev1alpha1.Release) int {
		return a.CreationTimestamp.Compare(b.CreationTimestamp.Time)
	})
	return &latestRelease, nil
}

func newRelease(component string, jiraIssues []*jira.Issue, releaseplan, snapshot string) (*releasev1alpha1.Release, error) {
	var cveNames []string
	for _, jiraIssue := range jiraIssues {
		if cve := getCVEName(jiraIssue.Fields.Summary); cve != "" {
			cveNames = append(cveNames, cve)
		}
	}

	releaseType := "RHBA"
	var cves []cve
	if len(cveNames) != 0 {
		releaseType = "RHSA"
		for _, cveName := range cveNames {
			cves = append(cves, cve{Component: component, Key: cveName})
		}
	}
	var fixedIssues []issue
	for _, jiraIssue := range jiraIssues {
		fixedIssues = append(fixedIssues, issue{
			ID:     jiraIssue.Key,
			Source: "issues.redhat.com",
		})
	}
	releaseNotes := releaseData{
		ReleaseNotes: releaseNotes{
			Type:   releaseType,
			CVEs:   cves,
			Issues: issues{Fixed: fixedIssues},
		},
	}
	data, err := json.Marshal(releaseNotes)
	if err != nil {
		return nil, err
	}
	return &releasev1alpha1.Release{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Release",
			APIVersion: releasev1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: releasev1alpha1.ReleaseSpec{
			Snapshot:    snapshot,
			ReleasePlan: releaseplan,
			Data: &runtime.RawExtension{
				Raw: data,
			},
		},
	}, nil

}

func getCVEName(summary string) string {
	if strings.HasPrefix(summary, "CVE-") {
		return strings.Split(summary, " ")[0]
	}
	return ""
}

func ticketRegex(projects []string) (*regexp.Regexp, error) {
	regex := fmt.Sprintf("%s-[0-9]*", projects[0])
	if len(projects) > 1 {
		for _, project := range projects[1:] {
			regex += fmt.Sprintf("|%s-[0-9]*", project)
		}
	}
	return regexp.Compile(regex)
}

func getJiraIssues(projects []string, commits []*github.Commit) ([]*jira.Issue, error) {
	re, err := ticketRegex(projects)
	if err != nil {
		return nil, err
	}
	var jiraIssues []string
	for _, commit := range commits {
		matches := re.FindAllString(commit.GetMessage(), -1)
		for _, match := range matches {
			jiraIssues = append(jiraIssues, match)
		}
	}

	// Remove duplicate tickets
	foundJiraTickets := make(map[string]*jira.Issue)
	for _, ticket := range jiraIssues {
		if _, found := foundJiraTickets[ticket]; !found {
			jiraIssue, err := jira.GetIssue(ticket)
			if err != nil {
				return nil, err
			}
			foundJiraTickets[ticket] = jiraIssue
		}
	}
	var uniqueTickets []*jira.Issue
	for _, jiraTicket := range foundJiraTickets {
		uniqueTickets = append(uniqueTickets, jiraTicket)
	}
	return uniqueTickets, nil
}
