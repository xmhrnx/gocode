package tools

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestReadFile(t *testing.T) {
	env := newTestEnv(t, map[string]string{
		"a.txt":     "one\ntwo\r\nthree\n",
		"empty.txt": "",
		"bin.dat":   "abc\x00def",
		"dir/x":     "",
	})
	tests := []struct {
		name    string
		args    readFileArgs
		want    string
		wantErr bool
	}{
		{"whole file", readFileArgs{Path: "a.txt"}, "     1\tone\n     2\ttwo\n     3\tthree\n", false},
		{"offset and limit", readFileArgs{Path: "a.txt", Offset: 2, Limit: 1},
			"     2\ttwo\n[showing lines 2-2 of 3; use offset=3 to read more]", false},
		{"absolute path", readFileArgs{Path: env.WorkDir + "/a.txt", Offset: 3}, "     3\tthree\n", false},
		{"empty", readFileArgs{Path: "empty.txt"}, "(empty file)", false},
		{"offset past end", readFileArgs{Path: "a.txt", Offset: 9}, "past the end", true},
		{"binary", readFileArgs{Path: "bin.dat"}, "binary", true},
		{"directory", readFileArgs{Path: "dir"}, "list_dir", true},
		{"missing", readFileArgs{Path: "nope"}, "not found", true},
		{"no path", readFileArgs{}, "path is required", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := runTool(t, ReadFile{}, env, tt.args)
			if res.IsError != tt.wantErr {
				t.Fatalf("IsError = %v, output %q", res.IsError, res.Output)
			}
			if tt.wantErr && !strings.Contains(res.Output, tt.want) || !tt.wantErr && res.Output != tt.want {
				t.Errorf("output %q, want %q", res.Output, tt.want)
			}
		})
	}
}

func TestReadFileTruncatesLongLines(t *testing.T) {
	env := newTestEnv(t, map[string]string{"long.txt": strings.Repeat("x", 2500) + "\n"})
	res := runTool(t, ReadFile{}, env, readFileArgs{Path: "long.txt"})
	if !strings.Contains(res.Output, "[line truncated]") || len(res.Output) > 2100 {
		t.Errorf("long line not truncated: %d bytes", len(res.Output))
	}
}

func TestReadFileBadArgs(t *testing.T) {
	res := ReadFile{}.Run(t.Context(), Env{}, json.RawMessage(`{"path": 3}`))
	if !res.IsError || !strings.Contains(res.Output, "invalid arguments") {
		t.Errorf("got %+v", res)
	}
}
