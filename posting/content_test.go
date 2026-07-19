package posting

import "testing"

func TestTruncateRunesUTF8(t *testing.T) {
	got := truncateRunes("こんにちは世界", 4)
	if got != "こ..." {
		t.Fatalf("got %q", got)
	}
}

func TestStripMarkdownLight(t *testing.T) {
	in := "# Title\n\n**bold** and `code`"
	got := stripMarkdownLight(in)
	if got == "" || got[0] == '#' {
		t.Fatalf("unexpected %q", got)
	}
}
