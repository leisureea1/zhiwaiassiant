package jwxt

import "testing"

func TestExtractCurrentSemesterNameFromHomeHTML(t *testing.T) {
	html := "<div class=\"modulebody\">" +
		"\"\u6559\u5b66\u7ba1\u7406\u7cfb\u7edf,\u4eca\u5929\u662f2026\u5e7405\u670825\u53f7/\u661f\u671f\u4e00,\u672c\u5468\u4e3a\"" +
		"<font color=\"blue\">2025-2026\u5b66\u5e74\u7b2c2\u5b66\u671f\u7684</font>" +
		"\"\u7b2c\"" +
		"<font color=\"red\" size=\"5\">13</font>" +
		"\"\u6559\u5b66\u5468\u3002\"" +
		"</div>"

	got := extractCurrentSemesterNameFromHTML(html)
	if got != "2025-2026-2" {
		t.Fatalf("expected current semester 2025-2026-2, got %q", got)
	}
}

func TestExtractCurrentSemesterNameDoesNotUseTeachingWeek(t *testing.T) {
	html := "2025-2026\u5b66\u5e74\u5b66\u671f\u7684\u7b2c13\u6559\u5b66\u5468"

	got := extractCurrentSemesterNameFromHTML(html)
	if got != "" {
		t.Fatalf("expected no semester when term digit is missing, got %q", got)
	}
}

func TestNormalizeSemesterTextMatchesListAndHomeFormats(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"2025-2026\u5b66\u5e74\u7b2c2\u5b66\u671f", "2025-2026-2"},
		{"2025-2026\u5b66\u5e74\u7b2c\u4e8c\u5b66\u671f", "2025-2026-2"},
		{"2025-2026\u5b66\u5e742\u5b66\u671f", "2025-2026-2"},
		{"2025-2026-2", "2025-2026-2"},
		{"2026-2027\u5b66\u5e74\u7b2c1\u5b66\u671f", "2026-2027-1"},
	}

	for _, tc := range cases {
		if got := normalizeSemesterText(tc.input); got != tc.want {
			t.Fatalf("normalizeSemesterText(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestMergeSemesterOptionsKeepsFallbackCurrentCandidate(t *testing.T) {
	dataQuery := []map[string]any{
		{"id": "249", "name": "2026-2027\u5b66\u5e741\u5b66\u671f"},
	}
	fallback := []map[string]any{
		{"id": "248", "name": "2025-2026\u5b66\u5e74\u7b2c2\u5b66\u671f"},
		{"id": "249", "name": "2026-2027\u5b66\u5e741\u5b66\u671f"},
	}

	options := mergeSemesterOptions(dataQuery, fallback)
	currentName := "2025-2026-2"
	current := ""
	for _, sem := range options {
		name := sem["name"].(string)
		if normalizeSemesterText(name) == normalizeSemesterText(currentName) {
			current = sem["id"].(string)
		}
	}

	if current != "248" {
		t.Fatalf("expected merged options to include current semester id 248, got %q from %#v", current, options)
	}
}

func TestExtractSemesterOptionsSkipsWeekOptions(t *testing.T) {
	html := "<option value=\"1\">\u7b2c1\u5468</option>" +
		"<option value=\"249\">2026-2027\u5b66\u5e741\u5b66\u671f</option>"

	options := extractSemesterOptions(html)
	if len(options) != 1 {
		t.Fatalf("expected only semester option, got %#v", options)
	}
	if options[0]["id"] != "249" {
		t.Fatalf("expected semester id 249, got %#v", options[0])
	}
}

func TestInferSemesterIDFromFutureOption(t *testing.T) {
	options := []map[string]any{
		{"id": "249", "name": "2026-2027\u5b66\u5e741\u5b66\u671f"},
	}

	id := inferSemesterIDFromOptions("2025-2026-2", options)
	if id != "248" {
		t.Fatalf("expected inferred id 248, got %q", id)
	}
}
