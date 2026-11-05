package skills

import (
	"net/url"
	"testing"
)

func TestParseRepoURL(t *testing.T) {
	tests := []struct {
		raw     string
		want    repoSource
		wantErr bool
	}{
		{raw: "https://github.com/anthropics/skills", want: repoSource{CloneURL: "https://github.com/anthropics/skills.git"}},
		{raw: "https://github.com/anthropics/skills.git", want: repoSource{CloneURL: "https://github.com/anthropics/skills.git"}},
		{raw: "https://www.github.com/anthropics/skills/", want: repoSource{CloneURL: "https://github.com/anthropics/skills.git"}},
		{
			raw:  "https://github.com/anthropics/skills/tree/main",
			want: repoSource{CloneURL: "https://github.com/anthropics/skills.git", Branch: "main"},
		},
		{
			raw:  "https://github.com/anthropics/skills/tree/v2/document-skills/pdf",
			want: repoSource{CloneURL: "https://github.com/anthropics/skills.git", Branch: "v2", Subdir: "document-skills/pdf"},
		},
		{raw: "https://gitlab.com/group/repo.git", want: repoSource{CloneURL: "https://gitlab.com/group/repo.git"}},
		{raw: "https://github.com/anthropics", wantErr: true},
		{raw: "https://github.com/anthropics/skills/blob/main/README.md", wantErr: true},
		{raw: "https://github.com/anthropics/skills/tree", wantErr: true},
		{raw: "https://github.com/anthropics/skills/tree/main/../../etc", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			u, err := url.Parse(tt.raw)
			if err != nil {
				t.Fatal(err)
			}
			got, err := parseRepoURL(u)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
