//go:build windows

package tools

import "testing"

func TestNativePath(t *testing.T) {
	cases := map[string]string{
		"/c/Users/me/x.go":          `C:\Users\me\x.go`,
		"/cygdrive/d/work":          `D:\work`,
		"/mnt/c/Users":              `C:\Users`,
		"/c":                        `C:\`,
		`C:\Users\me`:               `C:\Users\me`,
		"internal/tools/bash.go":    "internal/tools/bash.go",
		"/codes/not-a-drive-letter": "/codes/not-a-drive-letter",
	}
	for in, want := range cases {
		if got := nativePath(in); got != want {
			t.Errorf("nativePath(%q) = %q, want %q", in, got, want)
		}
	}
}
