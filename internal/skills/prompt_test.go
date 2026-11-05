package skills

import (
	"reflect"
	"strings"
	"testing"
)

func TestPromptSection(t *testing.T) {
	tests := []struct {
		name   string
		skills []Skill
		want   []string
	}{
		{name: "none", skills: nil, want: nil},
		{
			name: "lists each skill on one line",
			skills: []Skill{
				{Name: "pdf", Description: "Work with PDFs.\nUse for forms."},
				{Name: "xlsx", Description: "Spreadsheets."},
			},
			want: []string{"# Skills\n", "skill tool", "\n- pdf: Work with PDFs. Use for forms.", "\n- xlsx: Spreadsheets."},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PromptSection(tt.skills)
			if tt.want == nil && got != "" {
				t.Fatalf("got %q, want empty", got)
			}
			for _, part := range tt.want {
				if !strings.Contains(got, part) {
					t.Errorf("section %q lacks %q", got, part)
				}
			}
		})
	}
}

func TestEnabled(t *testing.T) {
	all := []Skill{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	tests := []struct {
		name     string
		disabled []string
		want     []Skill
	}{
		{"nothing disabled", nil, all},
		{"some disabled", []string{"b", "unknown"}, []Skill{{Name: "a"}, {Name: "c"}}},
		{"all disabled", []string{"a", "b", "c"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Enabled(all, tt.disabled); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
