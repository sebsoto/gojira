package release

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sigs.k8s.io/yaml"
	"text/template"
	"time"

	releasev1alpha1 "github.com/konflux-ci/release-service/api/v1alpha1"

	"github.com/sebsoto/gojira/pkg/jira"
)

type release struct {
	Zstream      bool
	HandoverDate string
	GADate       string
	Version      string
	Project      string
	Release      string
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

func newRelease(patch bool, version string, releaseDate time.Time, project string, konfluxRelease *releasev1alpha1.Release) *release {
	day := 24 * time.Hour
	qePeriod := 10 * day
	if patch {
		qePeriod = 7 * day
	}
	qeEndDate := roundDownToWeekday(releaseDate.Add(-day))
	qeHandoverDate := roundDownToWeekday(qeEndDate.Add(-qePeriod))
	releaseString, _ := yaml.Marshal(konfluxRelease)
	return &release{
		HandoverDate: formattedDate(qeHandoverDate),
		GADate:       formattedDate(releaseDate),
		Version:      version,
		Zstream:      patch,
		Project:      project,
		Release:      string(releaseString),
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
				Key: &r.Project,
			},
			IssueType: jira.IssueType{Name: jira.EpicIssue},
			TargetVersion: []jira.TargetVersion{
				{
					Name: fmt.Sprintf("WMCO %s", r.Version),
				},
			},
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
				Key: &r.Project,
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

func CreateIssues(jiraProject, version string, majorRelease bool, releaseDate time.Time, release *releasev1alpha1.Release) error {
	r := newRelease(majorRelease, version, releaseDate, jiraProject, release)
	epicKey, err := r.createReleaseEpic()
	if err != nil {
		return err
	}
	err = r.createReleaseTask(epicKey)
	if err != nil {
		return err
	}
	return nil
}

type fieldsUpdater struct {
	Fields fields `json:"fields"`
}

type fields struct {
	Description string `json:"description"`
}

func UpdateRelease(issue, jiraProject, version string, majorRelease bool, releaseDate time.Time, release *releasev1alpha1.Release) error {
	r := newRelease(majorRelease, version, releaseDate, jiraProject, release)
	t, err := template.New("release_task_template").ParseFiles("/home/sebsoto/code/openshift/gojira/templates/release_task_template")
	if err != nil {
		return err
	}
	description := new(bytes.Buffer)
	err = t.Execute(description, r)
	if err != nil {
		return err
	}
	update := fieldsUpdater{fields{Description: description.String()}}
	updateBody, err := json.Marshal(update)
	if err != nil {
		return err
	}
	fmt.Printf("Updating ticket with description:\n%s\n", string(updateBody))
	return jira.UpdateIssue(issue, string(updateBody))
}
