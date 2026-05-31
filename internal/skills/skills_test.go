package skills

import (
	"testing"
)

func TestExtractTriggers(t *testing.T) {
	tests := []struct {
		name string
		desc string
		want int
	}{
		{
			name: "triggers include list",
			desc: "Use this for creating documents. Triggers include: \"create\", \"edit\", \"delete\".",
			want: 3,
		},
		{
			name: "no triggers",
			desc: "A simple description without trigger keywords.",
			want: 0,
		},
		{
			name: "empty description",
			desc: "",
			want: 0,
		},
		{
			name: "comma separated",
			desc: "Triggers include: create, edit, delete, modify",
			want: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTriggers(tt.desc)
			if len(got) != tt.want {
				t.Errorf("extractTriggers(%q) = %v (len=%d), want len=%d", tt.desc, got, len(got), tt.want)
			}
		})
	}
}

func TestExtractDeps(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "npm install",
			body: "Some text\nnpm install some-package\nmore text",
			want: "npm install some-package",
		},
		{
			name: "no deps",
			body: "# Heading\n\nJust some instructions.",
			want: "",
		},
		{
			name: "pip install",
			body: "Run this:\npip install requests\n\nThen continue.",
			want: "pip install requests",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDeps(tt.body)
			if got != tt.want {
				t.Errorf("extractDeps(%q) = %q, want %q", tt.body, got, tt.want)
			}
		})
	}
}
