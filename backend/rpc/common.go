package rpc

import (
	"ccm-desktop-v2/backend/internal/report"
)

type IssueItem struct {
	Severity      string `json:"severity"`
	Domain        string `json:"domain"`
	Message       string `json:"message"`
	Detail        string `json:"detail"`
	FixSuggestion string `json:"fix"`
}

func issueToItem(iss report.Issue) IssueItem {
	return IssueItem{
		Severity:      string(iss.Severity),
		Domain:        iss.Domain,
		Message:       iss.Message,
		Detail:        iss.Detail,
		FixSuggestion: iss.FixSuggestion,
	}
}
