package errata

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type errataSearch struct {
	Data []errata `json:"data"`
}

type errata struct {
	ID                        int      `json:"id"`
	Type                      string   `json:"type"`
	TextOnly                  bool     `json:"text_only"`
	AdvisoryName              string   `json:"advisory_name"`
	Synopsis                  string   `json:"synopsis"`
	Revision                  int      `json:"revision"`
	Status                    string   `json:"status"`
	SecurityImpact            string   `json:"security_impact"`
	IsOperatorHotfix          bool     `json:"is_operator_hotfix"`
	IsOperatorPrerelease      bool     `json:"is_operator_prerelease"`
	SkipCustomerNotifications bool     `json:"skip_customer_notifications"`
	PreventAutoPushReady      bool     `json:"prevent_auto_push_ready"`
	SuppressPushRequestJira   bool     `json:"suppress_push_request_jira"`
	RespinCount               int      `json:"respin_count"`
	Pushcount                 int      `json:"pushcount"`
	ContentTypes              []string `json:"content_types"`
	Timestamps                struct {
		IssueDate      time.Time `json:"issue_date"`
		UpdateDate     time.Time `json:"update_date"`
		ReleaseDate    any       `json:"release_date"`
		StatusTime     time.Time `json:"status_time"`
		SecuritySLA    any       `json:"security_sla"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
		ActualShipDate any       `json:"actual_ship_date"`
		PublishDate    any       `json:"publish_date"`
		EmbargoDate    any       `json:"embargo_date"`
	} `json:"timestamps"`
	Flags struct {
		TextReady      bool `json:"text_ready"`
		Pushed         bool `json:"pushed"`
		Published      bool `json:"published"`
		Deleted        bool `json:"deleted"`
		QaComplete     bool `json:"qa_complete"`
		RhnComplete    bool `json:"rhn_complete"`
		DocComplete    bool `json:"doc_complete"`
		Rhnqa          bool `json:"rhnqa"`
		Closed         bool `json:"closed"`
		SignRequested  bool `json:"sign_requested"`
		EmbargoUndated bool `json:"embargo_undated"`
	} `json:"flags"`
	Product struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		ShortName string `json:"short_name"`
	} `json:"product"`
	Release struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"release"`
	People struct {
		AssignedTo       string `json:"assigned_to"`
		Reporter         string `json:"reporter"`
		QeGroup          string `json:"qe_group"`
		DocsGroup        string `json:"docs_group"`
		DocReviewer      string `json:"doc_reviewer"`
		DevelGroup       string `json:"devel_group"`
		PackageOwner     string `json:"package_owner"`
		SecurityReviewer any    `json:"security_reviewer"`
	} `json:"people"`
	Content struct {
		Topic       string `json:"topic"`
		Description string `json:"description"`
		Solution    string `json:"solution"`
		Keywords    string `json:"keywords"`
	} `json:"content"`
	Labels []any `json:"labels"`
}

type Summary struct {
	ID       int
	Synopsis string
}

func List() ([]Summary, error) {
	// Don't want to figure out a safe library for SPNEGO, so just using curl instead
	apiURL := "https://errata.devel.redhat.com/api/v1/"
	wmcoErrataSearch := fmt.Sprintf("%s/erratum/search?product[]=79&show_state_IN_PUSH=1&show_state_NEW_FILES=1&show_state_PUSH_READY=1&show_state_QE=1&show_state_REL_PREP=1&synopsis_text=Windows+Containers", apiURL)
	curlCmd := exec.Command("curl", "--negotiate", "--user", ":", wmcoErrataSearch)
	out, err := curlCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error running curl: %s: %w", string(out), err)
	}
	var errataList errataSearch
	err = json.Unmarshal(out, &errataList)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response %s: %w", string(out), err)
	}
	var summaries []Summary
	for _, e := range errataList.Data {
		summaries = append(summaries, Summary{ID: e.ID, Synopsis: e.Synopsis})
	}
	return summaries, nil
}

// URL returns the url for the given errata
func URL(id int) string {
	return fmt.Sprintf("https://errata.devel.redhat.com/advisory/%d", id)
}
