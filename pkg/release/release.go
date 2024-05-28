package release

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/sebsoto/gojira/pkg/errata"
	"github.com/sebsoto/gojira/pkg/jira"
)

type release struct {
	Zstream    bool
	EngFreeze  string
	QEHandover string
	QEStart    string
	QEEnd      string
	GA         string
	Version    string
}

func formattedDate(t time.Time) string {
	return t.Format(time.DateOnly)
}

func roundDownToWeekday(t time.Time) time.Time {
	if t.Weekday() == time.Sunday {
		return t.Add(2 * -24 * time.Hour)
	} else if t.Weekday() == time.Saturday {
		return t.Add(-24 * time.Hour)
	}
	return t
}

func newRelease(patch bool, version string, releaseDate time.Time) *release {
	day := 24 * time.Hour
	qePeriod := 10 * day
	if patch {
		qePeriod = 7 * day
	}
	qeEndDate := roundDownToWeekday(releaseDate.Add(-day))
	qeStartDate := roundDownToWeekday(qeEndDate.Add(-qePeriod))
	qeHandOverDate := roundDownToWeekday(qeStartDate.Add(-day))
	engFreeze := roundDownToWeekday(qeHandOverDate.Add(-day))
	return &release{
		EngFreeze:  formattedDate(engFreeze),
		QEHandover: formattedDate(qeHandOverDate),
		QEStart:    formattedDate(qeStartDate),
		QEEnd:      formattedDate(qeEndDate),
		GA:         formattedDate(releaseDate),
		Version:    version,
		Zstream:    patch,
	}
}

// createReleaseEpic creates the release epic, and returns the key, e.g. WINC-1111
func (r *release) createReleaseEpic() (string, error) {
	t, err := template.New("epic_template").ParseFiles("/home/sebsoto/code/openshift/gojira/templates/epic_template")
	if err != nil {
		return "", err
	}
	epicDescription := new(bytes.Buffer)
	err = t.Execute(epicDescription, r)
	if err != nil {
		return "", err
	}
	newIssue := jira.Issue{
		Fields: jira.IssueFields{
			Summary:     fmt.Sprintf("Windows Machine Config Operator %s Release", r.Version),
			Description: epicDescription.String(),
			Project: jira.Project{
				ID:  nil,
				Key: strPtr("WINC"),
			},
			IssueType: jira.IssueType{Name: jira.EpicIssue},
			TargetVersion: []jira.TargetVersion{
				{
					Name: fmt.Sprintf("WMCO %s", r.Version),
				},
			},
			StartDate: r.EngFreeze,
			EndDate:   r.GA,
			Security: &jira.Security{
				Name: "Red Hat Employee",
			},
			EpicName: fmt.Sprintf("WMCO %s Release", r.Version),
			Labels:   []string{"OperatorProductization"},
			Priority: &jira.Priority{Name: jira.MajorPriority},
		},
	}
	response, err := jira.CreateIssue(&newIssue)
	if err != nil {
		return "", err
	}
	return response.Key, nil
}

func (r *release) addErrataLink(key, synopsisSearch string) error {
	errataList, err := errata.List(synopsisSearch)
	if err != nil {
		return fmt.Errorf("error listing errata: %w", err)
	}
	var match *errata.Summary
	for _, e := range errataList {
		if strings.Contains(e.Synopsis, r.Version) {
			match = &e
			break
		}
	}
	if match == nil {
		return fmt.Errorf("no matching errata found")
	}
	fmt.Printf("linking errata %d:%s to epic %s", match.ID, match.Synopsis, key)
	return jira.AddRemoteLink(key, errata.URL(match.ID), match.Synopsis)
}

// createReleaseEpic creates the post-release task, and returns the key, e.g. WINC-1111
func (r *release) createPostReleaseTask() error {
	nextVersion, err := nextPatch(r.Version)
	if err != nil {
		return err
	}
	description := fmt.Sprintf("Post Release issue for release of version %s of Red Hat OpenShift for Windows Containers.\n\n"+
		"Prepare for the next release, %s.\n\n"+
		"Post release work tasks are documented in the process doc [Releasing Red Hat WMCO|https://docs.google.com/document/d/19oMK2F7PiYGLHoCtMMZ_hjrkOiTvibVD4H6yDf3tkRo]", r.Version, nextVersion)
	newStory := jira.Issue{
		Fields: jira.IssueFields{
			Summary:     fmt.Sprintf("Red Hat OpenShift for Windows Containers %s Post Release", r.Version),
			Description: description,
			Project: jira.Project{
				ID:  nil,
				Key: strPtr("WINC"),
			},
			IssueType: jira.IssueType{Name: jira.TaskIssue},
			Priority:  &jira.Priority{Name: jira.MajorPriority},
		},
	}
	response, err := jira.CreateIssue(&newStory)
	if err == nil {
		fmt.Printf("%+v\n", response)
	}
	return err
}

func (r *release) createReleaseTask(epicTicketID string) error {
	t, err := template.New("release_task_template").ParseFiles("/home/sebsoto/code/openshift/gojira/templates/release_task_template")
	if err != nil {
		return err
	}
	description := new(bytes.Buffer)
	err = t.Execute(description, r)
	if err != nil {
		return err
	}
	newStory := jira.Issue{
		Fields: jira.IssueFields{
			Summary:     fmt.Sprintf("Red Hat OpenShift for Windows Containers %s Release", r.Version),
			Description: description.String(),
			Project: jira.Project{
				ID:  nil,
				Key: strPtr("WINC"),
			},
			IssueType: jira.IssueType{Name: jira.TaskIssue},
			EpicLink:  epicTicketID,
			Labels:    []string{"docs", "qe", "release"},
			Priority:  &jira.Priority{Name: jira.MajorPriority},
		},
	}
	_, err = jira.CreateIssue(&newStory)
	return err
}

func nextPatch(version string) (string, error) {
	versionSplit := strings.Split(version, ".")
	patch, err := strconv.Atoi(versionSplit[len(versionSplit)-1])
	if err != nil {
		return "", err
	}
	versionSplit[len(versionSplit)-1] = strconv.Itoa(patch + 1)
	return strings.Join(versionSplit, "."), nil
}

func strPtr(str string) *string {
	return &str
}

func CreateIssues(jiraProject, errataSearch, version string, majorRelease bool, releaseDate time.Time) error {
	r := newRelease(majorRelease, version, releaseDate)
	epicKey, err := r.createReleaseEpic()
	if err != nil {
		return err
	}
	err = r.addErrataLink(epicKey, errataSearch)
	if err != nil {
		fmt.Printf("error adding errata link to epic, skipping: %s\n", err)
	}
	err = r.createReleaseTask(epicKey)
	if err != nil {
		return err
	}
	err = r.createPostReleaseTask()
	if err != nil {
		return err
	}
	return nil
}
