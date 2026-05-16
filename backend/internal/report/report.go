package report

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Severity string

const (
	Error   Severity = "error"
	Warning Severity = "warning"
	Info    Severity = "info"
	Success Severity = "success"
)

type Issue struct {
	Severity     Severity
	Domain       string
	Message      string
	Detail       string
	FixSuggestion string
}

type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
)

var (
	UseColor  = true
	Quiet     = false
	Format    = FormatText
)

func init() {
	// Enable ANSI on Windows
	enableANSIIfNeeded()
}

func PrintIssue(issue Issue) {
	if Quiet && issue.Severity == Info {
		return
	}

	icon := severityIcon(issue.Severity)
	color := severityColor(issue.Severity)

	if Format == FormatJSON {
		data, _ := json.Marshal(map[string]interface{}{
			"severity":       issue.Severity,
			"domain":         issue.Domain,
			"message":        issue.Message,
			"detail":         issue.Detail,
			"fix_suggestion": issue.FixSuggestion,
		})
		fmt.Println(string(data))
		return
	}

	line := fmt.Sprintf("  %s [%s] %s: %s", colorize(icon, color), colorize(string(issue.Severity), color), issue.Domain, issue.Message)
	fmt.Println(line)

	if issue.Detail != "" {
		fmt.Printf("    %s\n", issue.Detail)
	}
	if issue.FixSuggestion != "" {
		fmt.Printf("    %s %s\n", colorize("Fix:", color), issue.FixSuggestion)
	}
}

func Header(title string) {
	if Format == FormatJSON {
		return
	}
	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println(colorize(title, colorCyan))
	fmt.Println(strings.Repeat("=", 60))
}

func SubHeader(title string) {
	if Format == FormatJSON {
		return
	}
	fmt.Printf("\n%s\n", colorize(title, colorYellow))
}

func Summary(ok, warn, err int) {
	if Format == FormatJSON {
		data, _ := json.Marshal(map[string]int{"ok": ok, "warning": warn, "error": err})
		fmt.Println(string(data))
		return
	}
	fmt.Println()
	fmt.Printf("  %s %d OK  ", colorize("✓", colorGreen), ok)
	if warn > 0 {
		fmt.Printf("%s %d WARN  ", colorize("⚠", colorYellow), warn)
	}
	if err > 0 {
		fmt.Printf("%s %d ERR", colorize("✗", colorRed), err)
	}
	fmt.Println()
}

func InfoLine(msg string) {
	if Quiet || Format == FormatJSON {
		return
	}
	fmt.Printf("  %s\n", msg)
}

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

func colorize(s, color string) string {
	if !UseColor {
		return s
	}
	return color + s + colorReset
}

func severityIcon(s Severity) string {
	switch s {
	case Error:
		return "✗"
	case Warning:
		return "⚠"
	case Info:
		return "ℹ"
	case Success:
		return "✓"
	}
	return "?"
}

func severityColor(s Severity) string {
	switch s {
	case Error:
		return colorRed
	case Warning:
		return colorYellow
	case Info:
		return colorCyan
	case Success:
		return colorGreen
	}
	return colorReset
}

func enableANSIIfNeeded() {
	_ = os.Stdout
}

// Exported color helpers for other packages.
var (
	ColorRed    = colorRed
	ColorGreen  = colorGreen
	ColorYellow = colorYellow
	ColorCyan   = colorCyan
	ColorReset  = colorReset
)

func Colorize(s, color string) string {
	return colorize(s, color)
}
