package parser

import (
	"testing"
)

func TestExtractFrontmatter(t *testing.T) {
	input := "---\nname: test-skill\ndescription: A test skill\n---\n\n# Body here"
	fm, body, err := ExtractFrontmatter([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(fm) == 0 {
		t.Error("frontmatter should not be empty")
	}
	if string(body) != "# Body here" {
		t.Errorf("body = %q, want %q", string(body), "# Body here")
	}
}

func TestExtractFrontmatterNoFrontmatter(t *testing.T) {
	input := "# Just a heading\n\nSome content"
	_, body, err := ExtractFrontmatter([]byte(input))
	if err == nil {
		t.Error("expected error for no frontmatter, got nil")
	}
	if string(body) != input {
		t.Errorf("body should be the original content, got %q", string(body))
	}
}

func TestParseSkillFrontmatter(t *testing.T) {
	yaml := "name: test-skill\ndescription: A test skill description\n"
	sf, err := ParseSkillFrontmatter([]byte(yaml))
	if err != nil {
		t.Fatal(err)
	}
	if sf.Name != "test-skill" {
		t.Errorf("name = %q, want %q", sf.Name, "test-skill")
	}
	if sf.Description != "A test skill description" {
		t.Errorf("desc = %q", sf.Description)
	}
}

func TestParseSkillFrontmatterMissingName(t *testing.T) {
	yaml := "description: No name field\n"
	sf, err := ParseSkillFrontmatter([]byte(yaml))
	if err != nil {
		t.Fatal(err)
	}
	if sf.Name != "" {
		t.Errorf("expected empty name, got %q", sf.Name)
	}
}
