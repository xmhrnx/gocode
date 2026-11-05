package prompts

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Vars are the values substituted for {{name}} placeholders.
type Vars struct {
	WorkDir string
	Shell   string
}

// projectRuleFiles are read from the project root, in order, and appended
// to the prompt so the agent follows each project's own instructions.
var projectRuleFiles = []string{"AGENTS.md", "CLAUDE.md", ".gocode/rules.md"}

const maxRuleFileBytes = 64 << 10

// Build fills in template and appends the user's custom rules and, when
// loadProjectRules is set, the project's rule files.
func Build(template, customRules string, vars Vars, loadProjectRules bool) string {
	var b strings.Builder
	b.WriteString(expand(template, vars))
	if rules := strings.TrimSpace(customRules); rules != "" {
		b.WriteString("\n\n# User rules\n")
		b.WriteString(rules)
	}
	if loadProjectRules {
		for _, name := range projectRuleFiles {
			appendRuleFile(&b, vars.WorkDir, name)
		}
	}
	return b.String()
}

func appendRuleFile(b *strings.Builder, workDir, name string) {
	data, err := os.ReadFile(filepath.Join(workDir, name))
	if err != nil || len(data) == 0 {
		return // a project without rule files is normal
	}
	if len(data) > maxRuleFileBytes {
		data = data[:maxRuleFileBytes]
	}
	b.WriteString("\n\n# Project instructions (" + name + ")\n")
	b.Write(data)
}

func expand(template string, vars Vars) string {
	return strings.NewReplacer(
		"{{workdir}}", vars.WorkDir,
		"{{os}}", runtime.GOOS+"/"+runtime.GOARCH,
		"{{shell}}", vars.Shell,
		"{{date}}", time.Now().Format("2006-01-02"),
	).Replace(template)
}
