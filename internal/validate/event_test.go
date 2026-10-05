package validate

import (
	"strings"
	"testing"
)

const validEvent = `name: Corrida Rambla 10K
date: 2026-10-11
sport: run
city: Montevideo
venue: Rambla de Pocitos
distances: [10K, 5K]
links:
  site: https://example.org
  instagram: https://instagram.com/example
  register: https://example.org/inscripcion
description:
  es: Recorrido plano por la Rambla.
  en: Flat course along the waterfront.
`

var testCities = map[string]bool{"Montevideo": true}

// with returns validEvent with old replaced by new. It panics when old is
// absent, so a test cannot pass by checking the unmodified fixture.
func with(old, new string) string {
	if !strings.Contains(validEvent, old) {
		panic("fixture has no " + old)
	}
	return strings.Replace(validEvent, old, new, 1)
}

func problems(src string) []string {
	var out []string
	for _, p := range Event("events/2026/x.yaml", []byte(src), "2026", testCities) {
		out = append(out, p.String())
	}
	return out
}

func TestEventAcceptsValidFile(t *testing.T) {
	if got := problems(validEvent); len(got) != 0 {
		t.Fatalf("valid event reported problems: %v", got)
	}
}

func TestEventAcceptsEquivalentSpellings(t *testing.T) {
	tests := map[string]string{
		"quoted date":          with("date: 2026-10-11", `date: "2026-10-11"`),
		"numeric distances":    with("distances: [10K, 5K]", "distances: [42, 21]"),
		"no optional fields":   with("venue: Rambla de Pocitos\ndistances: [10K, 5K]\n", ""),
		"only instagram":       with("  site: https://example.org\n", ""),
		"only one language":    with("  en: Flat course along the waterfront.\n", ""),
		"windows line endings": strings.ReplaceAll(validEvent, "\n", "\r\n"),
		"byte order mark":      "\ufeff" + validEvent,
		"120 non-ASCII chars":  with("name: Corrida Rambla 10K", "name: "+strings.Repeat("ñ", 120)),
	}
	for name, src := range tests {
		t.Run(name, func(t *testing.T) {
			if got := problems(src); len(got) != 0 {
				t.Errorf("reported problems: %v", got)
			}
		})
	}
}

func TestEventRules(t *testing.T) {
	const descBlock = "description:\n  es: Recorrido plano por la Rambla.\n  en: Flat course along the waterfront.\n"
	tests := []struct{ name, src, want string }{
		{"empty file", "", "events/2026/x.yaml: file is empty"},
		{"broken yaml", "name: [a\n", "events/2026/x.yaml: invalid YAML"},
		{"unknown field", validEvent + "nmae: typo\n", "field nmae not found"},
		{"name missing", with("name: Corrida Rambla 10K\n", ""), "name: is required"},
		{"name blank", with("name: Corrida Rambla 10K", `name: "   "`), "name: is required"},
		{"name too long", with("name: Corrida Rambla 10K", "name: "+strings.Repeat("a", 121)), "name: must be at most 120 characters, got 121"},
		{"date missing", with("date: 2026-10-11\n", ""), "date: is required"},
		{"date impossible", with("2026-10-11", "2026-02-30"), `date: must be a real date written as YYYY-MM-DD, got "2026-02-30"`},
		{"date wrong format", with("2026-10-11", "11/10/2026"), "date: must be a real date written as YYYY-MM-DD"},
		{"date in another year", with("2026-10-11", "2027-01-10"), "date: year 2027 does not match the folder events/2026"},
		{"sport missing", with("sport: run\n", ""), "sport: is required"},
		{"sport unknown", with("sport: run", "sport: swim"), `sport: must be one of bike, run, roll, got "swim"`},
		{"city missing", with("city: Montevideo\n", ""), "city: is required"},
		{"city not listed", with("city: Montevideo", "city: Atlantida"), `city: "Atlantida" is not in cities.yaml`},
		{"venue too long", with("venue: Rambla de Pocitos", "venue: "+strings.Repeat("a", 121)), "venue: must be at most 120 characters, got 121"},
		{"distances empty", with("[10K, 5K]", "[]"), "distances: must list at least one distance"},
		{"distances too many", with("[10K, 5K]", "[1, 2, 3, 4, 5, 6, 7]"), "distances: must have at most 6 values, got 7"},
		{"distance too long", with("[10K, 5K]", "["+strings.Repeat("a", 21)+"]"), "distances[0]: must be at most 20 characters, got 21"},
		{"link not https", with("site: https://example.org", "site: http://example.org"), `links.site: must be an https:// URL, got "http://example.org"`},
		{"link not a url", with("register: https://example.org/inscripcion", "register: example"), `links.register: must be an https:// URL, got "example"`},
		{"no organizer link", with("  site: https://example.org\n  instagram: https://instagram.com/example\n", ""), "links: at least one of site, instagram, facebook is required"},
		{"description missing", with(descBlock, ""), "description: is required"},
		{"description without es or en", with(descBlock, "description:\n  pt: Percurso plano.\n"), "description: must include es or en"},
		{"description bad language code", with("  en: Flat", "  eng: Flat"), "description.eng: language code must be two lowercase letters"},
		{"description empty", with("  en: Flat course along the waterfront.", `  en: ""`), "description.en: must not be empty"},
		{"description too long", with("Flat course along the waterfront.", strings.Repeat("a", 501)), "description.en: must be at most 500 characters, got 501"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := problems(tt.src)
			if len(got) != 1 {
				t.Fatalf("want exactly one problem, got %d: %v", len(got), got)
			}
			if !strings.Contains(got[0], tt.want) {
				t.Errorf("problem = %q, want it to contain %q", got[0], tt.want)
			}
		})
	}
}

func TestEventReportsAllProblemsAtOnce(t *testing.T) {
	src := with("name: Corrida Rambla 10K\n", "")
	src = strings.Replace(src, "sport: run", "sport: swim", 1)
	src = strings.Replace(src, "site: https://example.org", "site: http://example.org", 1)
	got := problems(src)
	if len(got) != 3 {
		t.Fatalf("want 3 problems, got %d: %v", len(got), got)
	}
	for i, field := range []string{"name:", "sport:", "links.site:"} {
		if !strings.Contains(got[i], field) {
			t.Errorf("problem %d = %q, want field %q", i, got[i], field)
		}
	}
}

func TestProblemString(t *testing.T) {
	if got := (Problem{Path: "a.yaml", Field: "name", Msg: "is required"}).String(); got != "a.yaml: name: is required" {
		t.Errorf("with field: %q", got)
	}
	if got := (Problem{Path: "a.yaml", Msg: "file is empty"}).String(); got != "a.yaml: file is empty" {
		t.Errorf("without field: %q", got)
	}
}
