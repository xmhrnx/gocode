package jsonstr

import "testing"

func TestDecode(t *testing.T) {
	cases := []struct {
		raw  string
		want string
		ok   bool
	}{
		{`"plain"`, "plain", true},
		{`"tab\tnew\nquote\"slash\/back\\"`, "tab\tnew\nquote\"slash/back\\", true},
		{`"é😀"`, "é😀", true},
		{"\"half\xe2\x86\"", "half\xe2\x86", true},
		{`"bad\x"`, "", false},
		{`notastring`, "", false},
	}
	for _, c := range cases {
		got, ok := Decode([]byte(c.raw))
		if got != c.want || ok != c.ok {
			t.Errorf("Decode(%q) = %q, %v; want %q, %v", c.raw, got, ok, c.want, c.ok)
		}
	}
}

func TestSplitRuneRejoins(t *testing.T) {
	first, _ := Decode([]byte("\"a\xe2\x86\""))
	second, _ := Decode([]byte("\"\x92b\""))
	if first+second != "a→b" {
		t.Fatalf("got %q", first+second)
	}
}
