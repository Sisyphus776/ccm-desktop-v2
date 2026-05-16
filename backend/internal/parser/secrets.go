package parser

import (
	"regexp"
	"strings"
)

type SecretPattern struct {
	Name  string
	Regex *regexp.Regexp
}

type SecretFinding struct {
	Pattern  string
	Line     int
	Match    string
	Redacted string
}

var secretPatterns = []SecretPattern{
	{Name: "Anthropic API Key", Regex: regexp.MustCompile(`sk-ant-[a-zA-Z0-9_-]{20,}`)},
	{Name: "OpenAI API Key", Regex: regexp.MustCompile(`sk-[a-zA-Z0-9]{32,}`)},
	{Name: "GitHub Token", Regex: regexp.MustCompile(`ghp_[a-zA-Z0-9]{36}`)},
	{Name: "Generic API Key (JSON)", Regex: regexp.MustCompile(`"(?:api_?key|auth_?token|secret|password)"\s*:\s*"[^"]{8,}"`)},
	{Name: "JWT Token", Regex: regexp.MustCompile(`eyJ[a-zA-Z0-9_-]{10,}\.[a-zA-Z0-9_-]{10,}\.[a-zA-Z0-9_-]{10,}`)},
	{Name: "DeepSeek API Key", Regex: regexp.MustCompile(`sk-[a-zA-Z0-9]{20,}`)},
}

// ScanForSecrets scans content for known secret patterns.
func ScanForSecrets(content []byte) []SecretFinding {
	var findings []SecretFinding

	lines := strings.Split(string(content), "\n")
	for lineNum, line := range lines {
		for _, pat := range secretPatterns {
			if match := pat.Regex.FindString(line); match != "" {
				findings = append(findings, SecretFinding{
					Pattern:  pat.Name,
					Line:     lineNum + 1,
					Match:    match,
					Redacted: redact(match),
				})
			}
		}
	}
	return findings
}

func redact(s string) string {
	if len(s) <= 8 {
		return "***"
	}
	return s[:4] + "..." + s[len(s)-4:]
}
