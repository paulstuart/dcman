package main

import (
	"encoding/json"
	"fmt"
)

var (
	ErrNoJiraTicket    = fmt.Errorf("no JIRA ticket specified")
	ErrBadJiraTicket   = fmt.Errorf("not a valid JIRA ticket")
	ErrJiraNotAssigned = fmt.Errorf("user not assigned to JIRA ticket")
)

func Assignee(j string) string {
	var blob interface{}
	if err := json.Unmarshal([]byte(j), &blob); err != nil {
		fmt.Println("unmarshal text: %s error: %s", j, err)
	}
	issue := blob.(map[string]interface{})
	if f, ok := issue["fields"]; ok {
		fields := f.(map[string]interface{})
		if a, ok := fields["assignee"]; ok {
			assignee := a.(map[string]interface{})
			if email, ok := assignee["emailAddress"]; ok {
				return email.(string)
			}
		}
	}
	return "failed"
}

func JiraAssigned(jira, email string) error {
	if len(jira) == 0 {
		return ErrNoJiraTicket
	}
	return ErrBadJiraTicket
}
