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
		{"unknown field", validEvent + "nmae: typo\n", `unknown field "nmae"`},
		{"name missing", with("name: Corrida Rambla 10K\n", ""), "name: is required"},
		{"name blank", with("name: Corrida Rambla 10K", `name: "   "`), "name: is required"},
		{"name too long", with("name: Corrida Rambla 10K", "name: "+strings.Repeat("a", 121)), "name: must be at most 120 characters, got 121"},
		{"date missing", with("date: 2026-10-11\n", ""), "date: is required"},
		{"date impossible", with("2026-10-11", "2026-02-30"), `date: must be a real date written as YYYY-MM-DD, or YYYY-MM when the day is not announced yet, got "2026-02-30"`},
		{"date wrong format", with("2026-10-11", "11/10/2026"), "date: must be a real date written as YYYY-MM-DD, or YYYY-MM when the day is not announced yet"},
		{"date in another year", with("2026-10-11", "2027-01-10"), "date: year 2027 does not match the folder events/2026; move the file to events/2027/"},
		{"sport missing", with("sport: run\n", ""), "sport: is required"},
		{"sport unknown", with("sport: run", "sport: swim"), `sport: must be one of bike, run, roll, got "swim"`},
		{"city missing", with("city: Montevideo\n", ""), "city: is required"},
		{"city not listed", with("city: Montevideo", "city: Atlantida"), `city: "Atlantida" is not in cities.yaml`},
		{"venue too long", with("venue: Rambla de Pocitos", "venue: "+strings.Repeat("a", 121)), "venue: must be at most 120 characters, got 121"},
		{"distances empty", with("[10K, 5K]", "[]"), "distances: must list at least one distance"},
		{"distances too many", with("[10K, 5K]", "[1, 2, 3, 4, 5, 6, 7]"), "distances: must have at most 6 values, got 7"},
		{"distance empty", with("[10K, 5K]", `[10K, ""]`), "distances[1]: must not be empty"},
		{"distance too long", with("[10K, 5K]", "["+strings.Repeat("a", 21)+"]"), "distances[0]: must be at most 20 characters, got 21"},
		{"link not https", with("site: https://example.org", "site: http://example.org"), `links.site: must be an https:// URL, got "http://example.org"`},
		{"link not a url", with("register: https://example.org/inscripcion", "register: example"), `links.register: must be an https:// URL, got "example"`},
		{"no organizer link", with("  site: https://example.org\n  instagram: https://instagram.com/example\n", ""), "links: at least one of site, instagram, facebook is required"},
		{"description missing", with(descBlock, ""), "description: is required"},
		{"description without es or en", with(descBlock, "description:\n  pt: Percurso plano.\n"), "description: must include es or en"},
		{"description bad language code", with("  en: Flat", "  eng: Flat"), `description: language code "eng" must be two lowercase letters`},
		{"unknown link", with("  register: https://example.org/inscripcion\n", "  register: https://example.org/inscripcion\n  web: https://example.org\n"), `links: unknown field "web"`},
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

func TestEventReportsEveryProblemOfShapeAndUnknownField(t *testing.T) {
	src := validEvent + "nmae: typo\n"
	src = strings.Replace(src, "sport: run", "sport: swim", 1)
	got := problems(src)
	if len(got) != 2 {
		t.Fatalf("want 2 problems, got %v", got)
	}
	joined := strings.Join(got, "\n")
	for _, want := range []string{`events/2026/x.yaml: unknown field "nmae"`, "sport: must be one of"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in %v", want, got)
		}
	}
}

func TestEventInlineComments(t *testing.T) {
	tests := []struct{ name, src, want string }{
		{"name", with("name: Corrida Rambla 10K", "name: Fecha #3 Campeonato Nacional"),
			`events/2026/x.yaml: name: the text after " #" is ignored by YAML ("#3 Campeonato Nacional"); put the whole value in double quotes, or remove the comment`},
		{"description", with("  es: Recorrido plano por la Rambla.", "  es: Recorrido plano. #MVD10K #running"),
			`events/2026/x.yaml: description.es: the text after " #" is ignored by YAML ("#MVD10K #running"); put the whole value in double quotes, or remove the comment`},
		{"sport", with("sport: run", "sport: run # running"),
			`events/2026/x.yaml: sport: the text after " #" is ignored by YAML ("# running")`},
		{"link", with("site: https://example.org", "site: https://example.org # oficial"),
			`events/2026/x.yaml: links.site: the text after " #" is ignored by YAML ("# oficial")`},
		{"flow list", with("[10K, 5K]", "[10K, 5K] # dos"),
			`events/2026/x.yaml: distances: the text after " #" is ignored by YAML ("# dos")`},
		{"list item", with("distances: [10K, 5K]", "distances:\n  - 10K\n  - 5K #x"),
			`events/2026/x.yaml: distances[1]: the text after " #" is ignored by YAML ("#x")`},
		{"windows line endings", strings.ReplaceAll(with("name: Corrida Rambla 10K", "name: Fecha #3"), "\n", "\r\n"),
			`events/2026/x.yaml: name: the text after " #" is ignored by YAML ("#3");`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := problems(tt.src)
			if len(got) != 1 || !strings.Contains(got[0], tt.want) {
				t.Fatalf("problems = %q, want one containing %q", got, tt.want)
			}
		})
	}
}

func TestEventCommentsThatAreFine(t *testing.T) {
	tests := map[string]string{
		"own-line comments":   "# Archivo del evento\n" + with("sport: run", "# deporte\nsport: run"),
		"quoted value":        with("name: Corrida Rambla 10K", `name: "Fecha #3 Campeonato Nacional"`),
		"hash without space":  with("name: Corrida Rambla 10K", "name: Corrida#10K"),
		"comment after key":   with("links:\n", "links: # enlaces\n"),
		"quoted with comment": with("name: Corrida Rambla 10K", `name: "Fecha #3" `),
	}
	for name, src := range tests {
		t.Run(name, func(t *testing.T) {
			if got := problems(src); len(got) != 0 {
				t.Errorf("reported problems: %v", got)
			}
		})
	}
}

func TestEventSecondDocument(t *testing.T) {
	const want = `events/2026/x.yaml: the file must contain one YAML document; remove the "---" line and everything after it`
	tests := map[string]string{
		"valid second document":   validEvent + "---\n" + validEvent,
		"invalid second document": validEvent + "---\nname: [a\n",
		"second with one line":    validEvent + "---\nx: 1\n",
	}
	for name, src := range tests {
		t.Run(name, func(t *testing.T) {
			got := problems(src)
			if len(got) != 1 || got[0] != want {
				t.Fatalf("problems = %q, want [%q]", got, want)
			}
		})
	}
	for name, src := range map[string]string{
		"leading marker":  "---\n" + validEvent,
		"closing marker":  validEvent + "...\n",
		"trailing marker": validEvent + "---\n",
	} {
		t.Run(name, func(t *testing.T) {
			if got := problems(src); len(got) != 0 {
				t.Errorf("reported problems: %v", got)
			}
		})
	}
}

func TestEventShapeMistakes(t *testing.T) {
	const descBlock = "description:\n  es: Recorrido plano por la Rambla.\n  en: Flat course along the waterfront.\n"
	const linksBlock = "links:\n  site: https://example.org\n  instagram: https://instagram.com/example\n  register: https://example.org/inscripcion\n"
	tests := []struct{ name, src, want string }{
		{"distances as text", with("[10K, 5K]", "10K, 5K"), "events/2026/x.yaml: distances: must be a list in square brackets, for example [10K, 5K]"},
		{"description as text", with(descBlock, "description: Recorrido plano.\n"), `events/2026/x.yaml: description: must list the languages on separate indented lines, for example "es: ..."`},
		{"links as text", with(linksBlock, "links: https://example.org\n"), `events/2026/x.yaml: links: must list the links on separate indented lines, for example "site: https://..."`},
		{"description list", with(descBlock, "description:\n  - es\n"), `events/2026/x.yaml: description: must list the languages`},
		{"name as list", with("name: Corrida Rambla 10K", "name: [a, b]"), "events/2026/x.yaml: name: must be a single value"},
		{"distance item as list", with("[10K, 5K]", "[10K, [5K]]"), "events/2026/x.yaml: distances[1]: must be a single value"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := problems(tt.src)
			if len(got) != 1 || !strings.Contains(got[0], tt.want) {
				t.Fatalf("problems = %q, want exactly one containing %q", got, tt.want)
			}
			for _, bad := range []string{"[]string", "map[string]", "!!"} {
				if strings.Contains(got[0], bad) {
					t.Errorf("message has Go jargon %q: %s", bad, got[0])
				}
			}
		})
	}
}

func TestEventSyntaxErrorHints(t *testing.T) {
	tests := []struct{ name, src, prefix, suffix string }{
		{"colon in value", "name: Vuelta Ciclista: Etapa 1\n", "events/2026/x.yaml: invalid YAML: line 1:",
			` — a value that contains ": " must be put in double quotes`},
		{"colon in value, later line", with("sport: run", "sport: run: fast"), "events/2026/x.yaml: invalid YAML: line 3:",
			` — a value that contains ": " must be put in double quotes`},
		{"at sign", with("  instagram: https://instagram.com/example", "  instagram: @corridarambla"), "events/2026/x.yaml: invalid YAML: line ",
			` — a value that starts with "@" must be put in double quotes; links are full https:// addresses`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := problems(tt.src)
			if len(got) != 1 || !strings.HasPrefix(got[0], tt.prefix) || !strings.HasSuffix(got[0], tt.suffix) {
				t.Fatalf("problems = %q, want prefix %q and suffix %q", got, tt.prefix, tt.suffix)
			}
		})
	}
	got := problems("name: [a\n")
	if len(got) != 1 || !strings.HasPrefix(got[0], "events/2026/x.yaml: invalid YAML: line 1:") || strings.Contains(got[0], "—") {
		t.Errorf("other syntax errors changed: %q", got)
	}
}

func TestEventUnknownFieldsDoNotHideOtherProblems(t *testing.T) {
	got := problems(with("sport: run", "sport: swim") + "hora: 9\n")
	if len(got) != 2 {
		t.Fatalf("problems = %q", got)
	}
}

func TestEventLanguageCodeWithLineBreakStaysOnOneLine(t *testing.T) {
	src := with("  en: Flat course along the waterfront.", "  \"a\\nb: c\": Flat")
	got := problems(src)
	if len(got) != 1 {
		t.Fatalf("problems = %q", got)
	}
	if strings.Contains(got[0], "\n") || !strings.HasPrefix(got[0], "events/2026/x.yaml: description: language code ") {
		t.Errorf("problem = %q", got[0])
	}
	// The same through the exported type, with a real line break in the key.
	p := Event("events/2026/x.yaml", []byte(strings.Replace(validEvent, "  en: Flat course along the waterfront.", "  \"a\\nb\": Flat", 1)), "2026", testCities)
	if len(p) != 1 || strings.ContainsAny(p[0].String(), "\r\n") {
		t.Errorf("problems = %v", p)
	}
}

func TestEventCitiesUnavailable(t *testing.T) {
	src := with("city: Montevideo", "city: Atlantida")
	if got := Event("events/2026/x.yaml", []byte(src), "2026", nil); len(got) != 0 {
		t.Errorf("nil cities still reported: %v", got)
	}
	if got := Event("events/2026/x.yaml", []byte(src), "2026", map[string]bool{}); len(got) != 1 {
		t.Errorf("empty (loaded) cities must apply the rule: %v", got)
	}
	got := Event("events/2026/x.yaml", []byte(with("city: Montevideo\n", "")), "2026", nil)
	if len(got) != 1 || !strings.Contains(got[0].String(), "city: is required") {
		t.Errorf("city is required must still apply: %v", got)
	}
}

const event2027 = `name: Vuelta de Prueba
date: 2027-02-25
sport: bike
city: Montevideo
links:
  site: https://example.org
description:
  es: Texto.
`

func problems2027(src string) []string {
	var out []string
	for _, p := range Event("events/2027/x.yaml", []byte(src), "2027", testCities) {
		out = append(out, p.String())
	}
	return out
}

func TestEventDateEndDateStatusEntry(t *testing.T) {
	const p = "events/2027/x.yaml: "
	const dateMsg = "date: must be a real date written as YYYY-MM-DD, or YYYY-MM when the day is not announced yet, got "
	const laterMsg = p + "end_date: must be later than date (2027-02-25); for a one-day event remove end_date"
	const entryMsg = "entry: must be one of open, license, elite, got "
	rep := func(old, new string) string { return strings.Replace(event2027, old, new, 1) }
	add := func(lines string) string { return event2027 + lines }
	tests := []struct {
		name, src string
		want      []string
	}{
		{"1 month only", rep("2027-02-25", "2027-04"), nil},
		{"2 month 13", rep("2027-02-25", "2027-13"), []string{p + dateMsg + `"2027-13"`}},
		{"3 month one digit", rep("2027-02-25", "2027-4"), []string{p + dateMsg + `"2027-4"`}},
		{"4 year only", rep("2027-02-25", "2027"), []string{p + dateMsg + `"2027"`}},
		{"5 impossible day", rep("2027-02-25", "2027-02-30"), []string{p + dateMsg + `"2027-02-30"`}},
		{"6 month only, other year", rep("2027-02-25", "2026-04"), []string{p + "date: year 2026 does not match the folder events/2027; move the file to events/2026/"}},
		{"7 end date", add("end_date: 2027-02-28\n"), nil},
		{"8 end date equals date", add("end_date: 2027-02-25\n"), []string{laterMsg}},
		{"9 end date before date", add("end_date: 2027-02-20\n"), []string{laterMsg}},
		{"10 exactly 31 days", add("end_date: 2027-03-28\n"), nil},
		{"11 32 days", add("end_date: 2027-03-29\n"), []string{p + "end_date: is more than 31 days after date (2027-02-25); check the month and the year"}},
		{"12 across the year", rep("2027-02-25", "2027-12-28") + "end_date: 2028-01-03\n", nil},
		{"13 end date month only", add("end_date: 2027-02\n"), []string{p + `end_date: must be a real date written as YYYY-MM-DD, got "2027-02"`}},
		{"14 start without day", rep("2027-02-25", "2027-02") + "end_date: 2027-02-28\n", []string{p + "end_date: needs a start date with a day: write date as YYYY-MM-DD, or remove end_date"}},
		{"15 bad date and end date", rep("2027-02-25", "pronto") + "end_date: 2027-02-28\n", []string{p + dateMsg + `"pronto"`}},
		{"16 end date as list", add("end_date: [2027-02-28]\n"), []string{p + "end_date: must be a single value on one line, not a list or nested lines"}},
		{"17 end date comment", add("end_date: 2027-02-28 # a confirmar\n"), []string{p + `end_date: the text after " #" is ignored by YAML ("# a confirmar"); put the whole value in double quotes, or remove the comment`}},
		{"18 status tentative", add("status: tentative\n"), nil},
		{"18 status confirmed", add("status: confirmed\n"), nil},
		{"19 status maybe", add("status: maybe\n"), []string{p + `status: must be confirmed or tentative, got "maybe"`}},
		{"20 entry open", add("entry: open\n"), nil},
		{"20 entry license", add("entry: license\n"), nil},
		{"20 entry elite", add("entry: elite\n"), nil},
		{"21 entry licence", add("entry: licence\n"), []string{p + entryMsg + `"licence"; write it as license`}},
		{"22 entry pro", add("entry: pro\n"), []string{p + entryMsg + `"pro"`}},
		{"23 empty values", add("end_date:\nstatus:\nentry:\n"), nil},
		{"24 status twice", add("status: tentative\nstatus: tentative\n"), []string{p + "status: appears more than once; keep only one"}},
		{"25 misspelled field", add("enddate: 2027-02-28\n"), []string{p + `unknown field "enddate"`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := problems2027(tt.src)
			if len(got) != len(tt.want) {
				t.Fatalf("problems = %q, want %q", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("problem %d = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
