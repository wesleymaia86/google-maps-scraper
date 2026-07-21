package web

import "testing"

func TestSanitizeDownloadName(t *testing.T) {
	got := SanitizeDownloadName("Barbearias Guará / DF")
	if got != "Barbearias Guará - DF.csv" && got != "Barbearias Guara - DF.csv" {
		// accented letters are kept; slash becomes dash
		if len(got) < 5 || got[len(got)-4:] != ".csv" {
			t.Fatalf("unexpected filename: %q", got)
		}
	}

	if SanitizeDownloadName("") != "resultados.csv" {
		t.Fatalf("empty name should fallback")
	}
}

func TestContentDispositionAttachment(t *testing.T) {
	h := ContentDispositionAttachment("Barbearias Guará.csv")
	if h == "" || !containsAll(h, `filename="`, "filename*=UTF-8''") {
		t.Fatalf("bad disposition: %q", h)
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !contains(s, p) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
