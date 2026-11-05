package skills

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"gocode/internal/tools"
)

func TestSkillTool(t *testing.T) {
	dir := writeSkill(t, t.TempDir(), "pdf", "pdf", "PDFs")
	writeFiles(t, dir, map[string]string{"scripts/extract.py": "pass"})
	var installed []Skill
	tool := NewTool(func() []Skill { return installed })
	if tool.Spec().Name != "skill" || tool.Access() != tools.AccessRead {
		t.Fatalf("spec %q access %q", tool.Spec().Name, tool.Access())
	}
	if !json.Valid(tool.Spec().Parameters) {
		t.Fatal("parameters are not valid JSON")
	}
	// Installed after the tool was made: list must be read at run time.
	installed = []Skill{{Name: "pdf", Description: "PDFs", Dir: dir}}

	tests := []struct {
		name      string
		args      string
		wantError bool
		wantParts []string
	}{
		{"loads a skill", `{"name":"pdf"}`, false, []string{
			"Base directory for this skill: " + dir + "\n\n# pdf\nDo the thing.",
			"Files in this skill (relative to the base directory):\n- scripts/extract.py",
		}},
		{"unknown skill", `{"name":"docx"}`, true, []string{`unknown skill "docx"`, "available skills: pdf"}},
		{"bad arguments", `{"name":`, true, []string{"invalid arguments"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tool.Run(context.Background(), tools.Env{}, json.RawMessage(tt.args))
			if result.IsError != tt.wantError {
				t.Fatalf("IsError = %v: %s", result.IsError, result.Output)
			}
			for _, part := range tt.wantParts {
				if !strings.Contains(result.Output, part) {
					t.Errorf("output %q lacks %q", result.Output, part)
				}
			}
		})
	}
	if got := tool.Summary(json.RawMessage(`{"name":"pdf"}`)); got != "Load skill pdf" {
		t.Errorf("summary = %q", got)
	}
}
