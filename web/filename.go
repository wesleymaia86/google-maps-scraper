package web

import (
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var unsafeFilenameChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]+`)

// SanitizeDownloadName builds a safe CSV filename from the job name.
func SanitizeDownloadName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "resultados.csv"
	}

	var b strings.Builder
	for _, r := range name {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		case r == ' ' || r == '_' || r == '-' || r == '.':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}

	cleaned := unsafeFilenameChars.ReplaceAllString(b.String(), "-")
	cleaned = strings.Trim(cleaned, " .-_")
	for strings.Contains(cleaned, "--") {
		cleaned = strings.ReplaceAll(cleaned, "--", "-")
	}

	if cleaned == "" {
		cleaned = "resultados"
	}

	if len(cleaned) > 80 {
		cleaned = cleaned[:80]
	}

	if !strings.HasSuffix(strings.ToLower(cleaned), ".csv") {
		cleaned += ".csv"
	}

	return cleaned
}

// ContentDispositionAttachment builds Content-Disposition with ASCII fallback and UTF-8 filename*.
func ContentDispositionAttachment(filename string) string {
	ascii := filepath.Base(filename)
	ascii = strings.Map(func(r rune) rune {
		if r > 127 || r < 32 {
			return '_'
		}
		switch r {
		case '"', '\\':
			return '_'
		default:
			return r
		}
	}, ascii)

	if ascii == "" || ascii == ".csv" {
		ascii = "resultados.csv"
	}

	return `attachment; filename="` + ascii + `"; filename*=UTF-8''` + url.PathEscape(filename)
}
