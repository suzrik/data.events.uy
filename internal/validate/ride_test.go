package validate

import (
	"strings"
	"testing"
)

const validRide = `name: Salida de los Sábados
sport: road
city: Montevideo
venue: Plaza Virgilio
days: [sat]
time: "07:30"
links:
  instagram: https://instagram.com/example
description:
  es: Salida grupal por la rambla, a ritmo tranquilo.
  en: Group ride along the waterfront at an easy pace.
`

// withRide returns validRide with old replaced by new. It panics when old
// is absent, so a test cannot pass by checking the unmodified fixture.
func withRide(old, new string) string {
	if !strings.Contains(validRide, old) {
		panic("fixture has no " + old)
	}
	return strings.Replace(validRide, old, new, 1)
}

func rideProblems(src string) []string {
	var out []string
	for _, p := range Ride("rides/x.yaml", []byte(src), testCities) {
		out = append(out, p.String())
	}
	return out
}

func TestRideAcceptsValidFiles(t *testing.T) {
	tests := map[string]string{
		"weekly, one day":          validRide,
		"weekly, two days":         withRide("days: [sat]", "days: [tue, thu]"),
		"every day":                withRide("days: [sat]", "days: [mon, tue, wed, thu, fri, sat, sun]"),
		"monthly, third sunday":    withRide("days: [sat]", "days: [sun]\nweek: 3"),
		"monthly, last saturday":   withRide("days: [sat]", "days: [sat]\nweek: last"),
		"time without quotes":      withRide(`time: "07:30"`, "time: 18:45"),
		"two cycling disciplines":  withRide("sport: road", "sport: [road, gravel]"),
		"a running group":          withRide("sport: road", "sport: run"),
		"with a registration link": withRide("  instagram: https://instagram.com/example\n", "  instagram: https://instagram.com/example\n  register: https://example.org/form\n"),
		"only spanish":             withRide("  en: Group ride along the waterfront at an easy pace.\n", ""),
	}
	tests["no venue and no time"] = strings.Replace(withRide("venue: Plaza Virgilio\n", ""), "time: \"07:30\"\n", "", 1)
	for name, src := range tests {
		t.Run(name, func(t *testing.T) {
			if got := rideProblems(src); len(got) != 0 {
				t.Errorf("reported problems: %v", got)
			}
		})
	}
}

func TestRideRules(t *testing.T) {
	tests := []struct{ name, src, want string }{
		{"days missing", withRide("days: [sat]\n", ""), "rides/x.yaml: days: is required"},
		{"days empty", withRide("days: [sat]", "days: []"), "days: is required"},
		{"days as text", withRide("days: [sat]", "days: sat"), "days: must be a list in square brackets, for example [tue, thu]"},
		{"day in spanish", withRide("days: [sat]", "days: [sábado]"), `days: must use mon, tue, wed, thu, fri, sat or sun, got "sábado"; write it as sat`},
		{"day in english", withRide("days: [sat]", "days: [Thursday]"), `days: must use mon, tue, wed, thu, fri, sat or sun, got "Thursday"; write it as thu`},
		{"day in capitals", withRide("days: [sat]", "days: [SAT]"), `days: must use mon, tue, wed, thu, fri, sat or sun, got "SAT"; write it as sat`},
		{"unknown day", withRide("days: [sat]", "days: [sat, finde]"), `days: must use mon, tue, wed, thu, fri, sat or sun, got "finde"`},
		{"day repeated", withRide("days: [sat]", "days: [sat, sun, sat]"), `days: "sat" appears more than once`},
		{"day item as list", withRide("days: [sat]", "days: [sat, [sun]]"), "days[1]: must be a single value"},
		{"week out of range", withRide("days: [sat]", "days: [sat]\nweek: 5"), `week: must be 1, 2, 3, 4 or last, got "5"`},
		{"week with two days", withRide("days: [sat]", "days: [sat, sun]\nweek: 1"), "week: needs exactly one day in days, for example days: [sun] with week: 3"},
		{"time with am/pm", withRide(`time: "07:30"`, `time: "7:30 pm"`), `time: must be written as HH:MM on a 24-hour clock, for example "18:45", got "7:30 pm"`},
		{"time out of range", withRide(`time: "07:30"`, `time: "25:00"`), `time: must be written as HH:MM on a 24-hour clock, for example "18:45", got "25:00"`},
		{"time without minutes", withRide(`time: "07:30"`, "time: 19"), `time: must be written as HH:MM on a 24-hour clock, for example "18:45", got "19"`},
		{"a date", validRide + "date: 2026-10-11\n", `rides/x.yaml: unknown field "date"; a ride has no date: say when it runs with days`},
		{"a status", validRide + "status: tentative\n", `rides/x.yaml: unknown field "status"`},
		{"name missing", withRide("name: Salida de los Sábados\n", ""), "name: is required"},
		{"sport missing", withRide("sport: road\n", ""), "sport: is required"},
		{"sport unknown", withRide("sport: road", "sport: bike"), `sport: must be one of road, mtb, gravel, run, roll, trail, tri, got "bike"`},
		{"sports of two kinds", withRide("sport: road", "sport: [road, run]"), "sport: only the cycling disciplines road, mtb and gravel can be combined"},
		{"city not listed", withRide("city: Montevideo", "city: Atlantida"), `city: "Atlantida" is not in cities.yaml`},
		{"no organizer link", withRide("  instagram: https://instagram.com/example\n", "  register: https://example.org/form\n"), "links: at least one of site, instagram, facebook is required"},
		{"description missing", withRide("description:\n  es: Salida grupal por la rambla, a ritmo tranquilo.\n  en: Group ride along the waterfront at an easy pace.\n", ""), "description: is required"},
		{"comment after days", withRide("days: [sat]", "days: [sat] # sábados"), `days: the text after " #" is ignored by YAML ("# sábados"); remove it, or put the note in the description`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rideProblems(tt.src)
			if len(got) != 1 {
				t.Fatalf("want exactly one problem, got %d: %v", len(got), got)
			}
			if !strings.Contains(got[0], tt.want) {
				t.Errorf("problem = %q, want it to contain %q", got[0], tt.want)
			}
		})
	}
}

// An event file must not take the fields of a ride.
func TestEventRejectsRideFields(t *testing.T) {
	got := problems(validEvent + "days: [sat]\n")
	if len(got) != 1 || !strings.Contains(got[0], `unknown field "days"`) {
		t.Errorf("problems = %v", got)
	}
}

func TestDirChecksRides(t *testing.T) {
	rep, lines := report(t, map[string]string{
		"cities.yaml":                         citiesFile,
		"events/2026/corrida-rambla-10k.yaml": validEvent,
		"rides/salida-de-los-sabados.yaml":    validRide,
		"rides/salida-mensual.yaml":           withRide("days: [sat]", "days: [sun]\nweek: 3"),
		"rides/.gitkeep":                      "",
	})
	if len(lines) != 0 {
		t.Fatalf("problems: %v", lines)
	}
	if rep.Events != 1 || rep.Rides != 2 {
		t.Errorf("Events = %d, Rides = %d, want 1 and 2", rep.Events, rep.Rides)
	}
}

func TestDirReportsRideProblems(t *testing.T) {
	_, lines := report(t, map[string]string{
		"cities.yaml":            citiesFile,
		"rides/Salida.yaml":      validRide,
		"rides/2026/salida.yaml": validRide,
		"rides/sin-dias.yaml":    withRide("days: [sat]\n", ""),
	})
	want := []string{
		"rides/2026/salida.yaml: must be rides/<slug>.yaml; the slug uses lowercase letters, digits and hyphens, up to 80 characters",
		"rides/Salida.yaml: must be rides/<slug>.yaml; the slug uses lowercase letters, digits and hyphens, up to 80 characters",
		"rides/sin-dias.yaml: days: is required",
	}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Errorf("problems:\n%s\nwant:\n%s", strings.Join(lines, "\n"), strings.Join(want, "\n"))
	}
}
