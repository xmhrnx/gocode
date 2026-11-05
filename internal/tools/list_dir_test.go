package tools

import (
	"strings"
	"testing"
)

func TestListDir(t *testing.T) {
	env := newTestEnv(t, map[string]string{
		"b.txt":              "",
		"a/inner.go":         "",
		"a/deep/deeper/x.go": "",
		"node_modules/m.js":  "",
		".gitignore":         "*.log\n",
		"debug.log":          "",
	})
	res := runTool(t, ListDir{}, env, listDirArgs{})
	want := "./\n  a/\n    deep/\n    inner.go\n  .gitignore\n  b.txt"
	if res.IsError || res.Output != want {
		t.Errorf("got:\n%s\nwant:\n%s", res.Output, want)
	}
	deep := runTool(t, ListDir{}, env, listDirArgs{Path: "a", Depth: 9})
	if !strings.Contains(deep.Output, "    deeper/\n      x.go") {
		t.Errorf("depth not applied:\n%s", deep.Output)
	}
	if file := runTool(t, ListDir{}, env, listDirArgs{Path: "b.txt"}); !file.IsError {
		t.Errorf("listing a file should fail: %+v", file)
	}
}

func TestListDirCapsEntries(t *testing.T) {
	files := map[string]string{}
	for i := 0; i < maxListEntries+10; i++ {
		files["d/"+strings.Repeat("f", 1+i%7)+string(rune('a'+i%26))+strings.Repeat("x", i/26)] = ""
	}
	env := newTestEnv(t, files)
	res := runTool(t, ListDir{}, env, listDirArgs{})
	if !strings.Contains(res.Output, "[listing stopped at 500 entries") {
		t.Errorf("no cap note in output of %d lines", strings.Count(res.Output, "\n"))
	}
}
