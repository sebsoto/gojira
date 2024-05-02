package jira

import (
	"fmt"
	"testing"
)

func TestSearch(t *testing.T) {
	issues, err := Search("project = WINC AND issuetype = Epic AND labels in (OperatorProductization) AND statusCategory != \"Done\"")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	for _, issue := range issues {
		fmt.Printf("%s\t%s\thttps://issues.redhat.com/browse/%s\n", issue.Key, issue.Fields.EpicName, issue.Key)
	}
}
