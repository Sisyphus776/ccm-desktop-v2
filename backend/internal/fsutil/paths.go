package fsutil

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// NormalizePath converts mixed separators and MSYS2 paths to canonical form.
func NormalizePath(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	// MSYS2/Cygwin: /c/Users/... -> C:/Users/...
	if matched, _ := regexp.MatchString(`^/[a-zA-Z]/`, p); matched {
		drive := strings.ToUpper(string(p[1]))
		p = drive + ":" + p[2:]
	}
	return p
}

// IsSymlink checks if a path is a symlink/reparse point (Windows-compatible via Readlink).
func IsSymlink(path string) (bool, error) {
	_, err := os.Readlink(path)
	return err == nil, nil
}

// ReadSymlink returns the target of a symlink.
func ReadSymlink(path string) (string, error) {
	return os.Readlink(path)
}

// PathExists checks if a path exists.
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// FindAbsolutePaths extracts absolute path references from text content.
// Returns all matches: Windows paths (C:\..., D:\...), MSYS2 paths (/c/.../...),
// and Unix absolute paths (/usr/...).
func FindAbsolutePaths(content string) []PathRef {
	var refs []PathRef
	seen := map[string]bool{}

	// Windows absolute paths: C:\..., D:\... (with quotes or whitespace boundary)
	winRe := regexp.MustCompile(`[A-Za-z]:[/\\][^\s"'\n\r\t]+`)
	for _, m := range winRe.FindAllString(content, -1) {
		normalized := NormalizePath(m)
		if !seen[normalized] {
			seen[normalized] = true
			refs = append(refs, PathRef{Raw: m, Normalized: normalized})
		}
	}

	// MSYS2/Cygwin paths: /c/..., /d/... (driver-letter style)
	msysRe := regexp.MustCompile(`/[a-zA-Z]/(?:Users|home|opt|tmp|var|etc|usr)/[^\s"'\n\r\t]+`)
	for _, m := range msysRe.FindAllString(content, -1) {
		normalized := NormalizePath(m)
		if !seen[normalized] {
			seen[normalized] = true
			refs = append(refs, PathRef{Raw: m, Normalized: normalized})
		}
	}

	// Unix absolute paths (but not MSYS2 driver-letter paths)
	unixRe := regexp.MustCompile(`(?:\s|"|'|^)(/(?:usr|opt|etc|tmp|var|home)/[^\s"'\n\r\t]+)`)
	for _, m := range unixRe.FindAllStringSubmatch(content, -1) {
		if len(m) > 1 {
			normalized := NormalizePath(m[1])
			if !seen[normalized] && !strings.HasPrefix(normalized, "/") {
				seen[normalized] = true
				refs = append(refs, PathRef{Raw: m[1], Normalized: normalized})
			}
		}
	}

	return refs
}

// FindEnvVarRefs extracts environment variable references like %VAR% or $VAR.
func FindEnvVarRefs(content string) []string {
	var refs []string
	seen := map[string]bool{}
	re := regexp.MustCompile(`%(\w+)%`)
	for _, m := range re.FindAllStringSubmatch(content, -1) {
		if len(m) > 1 && !seen[m[1]] {
			seen[m[1]] = true
			refs = append(refs, m[1])
		}
	}
	return refs
}

type PathRef struct {
	Raw        string
	Normalized string
}

// IsPortable checks if a path is likely portable across machines.
func IsPortable(path, homeDir string) (bool, string) {
	normalized := NormalizePath(path)
	normalizedHome := NormalizePath(homeDir)

	// Relative paths are portable
	if !filepath.IsAbs(path) && !strings.HasPrefix(normalized, "/") && !regexp.MustCompile(`^[A-Za-z]:`).MatchString(normalized) {
		return true, "relative"
	}

	// Env var references are portable (runtime resolution)
	if strings.Contains(path, "%") || strings.Contains(path, "$") {
		return true, "env-ref"
	}

	// Inside home directory - username might differ
	if strings.HasPrefix(strings.ToLower(normalized), strings.ToLower(normalizedHome)) {
		return false, "home-dir"
	}

	// D: drive paths - drive may not exist on target
	if strings.HasPrefix(normalized, "D:") {
		return false, "d-drive"
	}

	// C: drive system paths - usually exist
	if strings.HasPrefix(normalized, "C:/Windows") || strings.HasPrefix(normalized, "C:/Program Files") {
		return true, "system"
	}

	// Other absolute paths
	if regexp.MustCompile(`^[A-Za-z]:`).MatchString(normalized) {
		return false, "absolute-drive"
	}

	return false, "unknown"
}

// PortabilitySeverity returns the severity of a portability issue.
func PortabilitySeverity(reason string) string {
	switch reason {
	case "d-drive", "absolute-drive":
		return "critical"
	case "home-dir":
		return "warning"
	default:
		return "info"
	}
}
