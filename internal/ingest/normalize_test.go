package ingest

import "testing"

func TestNormalizeStripsInvisible(t *testing.T) {
	zwsp, zwj, rlo := string(rune(0x200B)), string(rune(0x200D)), string(rune(0x202E))
	in := "ig" + zwsp + "no" + zwj + "re" + rlo + " me"
	got := Normalize(in)
	if want := "ignore me"; got != want {
		t.Errorf("Normalize = %q, want %q", got, want)
	}
}

func TestNormalizeFoldsFullwidth(t *testing.T) {
	in := ""
	for _, r := range "ignore" {
		in += string(r + 0xFEE0)
	}
	got := Normalize(in)
	if want := "ignore"; got != want {
		t.Errorf("Normalize = %q, want %q", got, want)
	}
}

func TestNormalizeCollapsesWhitespace(t *testing.T) {
	got := Normalize("  ignore\t\t all\n\nprevious  ")
	if want := "ignore all previous"; got != want {
		t.Errorf("Normalize = %q, want %q", got, want)
	}
}

func TestNormalizePreservesCase(t *testing.T) {
	got := Normalize("Ignore ALL")
	if want := "Ignore ALL"; got != want {
		t.Errorf("Normalize = %q, want %q", got, want)
	}
}
