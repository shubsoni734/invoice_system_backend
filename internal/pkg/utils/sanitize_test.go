package utils

import "testing"

func TestSanitizeString(t *testing.T) {
	if got := SanitizeString("  Acme  "); got != "Acme" {
		t.Fatalf("got %q", got)
	}
}

func TestSanitizeSlug(t *testing.T) {
	cases := map[string]string{
		"Acme Books":     "acme-books",
		"  Hello--World ": "hello-world",
		"GST & Co.":      "gst-co",
		"---":            "",
	}
	for input, want := range cases {
		if got := SanitizeSlug(input); got != want {
			t.Errorf("SanitizeSlug(%q) = %q, want %q", input, got, want)
		}
	}
}
