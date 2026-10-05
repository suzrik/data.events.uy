package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// tree writes files (slash-separated name -> content) into a temp dir.
func tree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, body := range files {
		p := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func report(t *testing.T, files map[string]string) (Report, []string) {
	t.Helper()
	rep, err := Dir(tree(t, files))
	if err != nil {
		t.Fatalf("Dir: %v", err)
	}
	var lines []string
	for _, p := range rep.Problems {
		lines = append(lines, p.String())
	}
	return rep, lines
}

const citiesFile = "- Montevideo\n- Punta del Este\n"

func TestDirAcceptsValidCatalog(t *testing.T) {
	rep, lines := report(t, map[string]string{
		"cities.yaml":                         citiesFile,
		"events/2026/corrida-rambla-10k.yaml": validEvent,
		"events/2027/corrida-rambla-10k.yaml": strings.Replace(validEvent, "2026-10-11", "2027-10-10", 1),
		"events/.gitkeep":                     "",
		"events/2026/.DS_Store":               "junk",
	})
	if len(lines) != 0 {
		t.Fatalf("problems: %v", lines)
	}
	if rep.Events != 2 {
		t.Errorf("Events = %d, want 2", rep.Events)
	}
}

func TestDirWithoutEventsFolder(t *testing.T) {
	rep, lines := report(t, map[string]string{"cities.yaml": citiesFile})
	if len(lines) != 0 || rep.Events != 0 {
		t.Fatalf("Events = %d, problems: %v", rep.Events, lines)
	}
}

func TestDirRejectsMisplacedFiles(t *testing.T) {
	const want = "must be events/<year>/<slug>.yaml"
	bad := []string{
		"events/2026/Corrida.yaml",
		"events/2026/corrida_rambla.yaml",
		"events/2026/corrida.yml",
		"events/corrida.yaml",
		"events/2026/sub/corrida.yaml",
		"events/26/corrida.yaml",
		"events/2026/" + strings.Repeat("a", 81) + ".yaml",
	}
	for _, name := range bad {
		t.Run(name, func(t *testing.T) {
			rep, lines := report(t, map[string]string{"cities.yaml": citiesFile, name: validEvent})
			if len(lines) != 1 || !strings.Contains(lines[0], want) || !strings.HasPrefix(lines[0], name+": ") {
				t.Fatalf("problems: %v", lines)
			}
			if rep.Events != 0 {
				t.Errorf("Events = %d, want 0", rep.Events)
			}
		})
	}
}

func TestDirAcceptsSlugOfMaximumLength(t *testing.T) {
	name := "events/2026/" + strings.Repeat("a", 80) + ".yaml"
	_, lines := report(t, map[string]string{"cities.yaml": citiesFile, name: validEvent})
	if len(lines) != 0 {
		t.Fatalf("problems: %v", lines)
	}
}

func TestDirReportsEveryBadFileSortedByPath(t *testing.T) {
	_, lines := report(t, map[string]string{
		"cities.yaml":           citiesFile,
		"events/2026/zeta.yaml": strings.Replace(validEvent, "sport: run", "sport: swim", 1),
		"events/2026/alfa.yaml": strings.Replace(validEvent, "city: Montevideo", "city: Atlantida", 1),
	})
	if len(lines) != 2 {
		t.Fatalf("want 2 problems, got %v", lines)
	}
	if !strings.HasPrefix(lines[0], "events/2026/alfa.yaml: city:") || !strings.HasPrefix(lines[1], "events/2026/zeta.yaml: sport:") {
		t.Errorf("order or content wrong: %v", lines)
	}
}

func TestDirCitiesFile(t *testing.T) {
	tests := []struct{ name, cities, want string }{
		{"duplicate", "- Montevideo\n- Salto\n- Montevideo\n", `cities.yaml: [2]: "Montevideo" is listed twice`},
		{"empty entry", "- Montevideo\n- \"\"\n", "cities.yaml: [1]: must not be empty"},
		{"surrounding space", "- \" Montevideo\"\n", "cities.yaml: [0]: must not start or end with a space"},
		{"not a list", "Montevideo: yes\n", "cities.yaml: invalid YAML"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, lines := report(t, map[string]string{"cities.yaml": tt.cities})
			if len(lines) != 1 || !strings.Contains(lines[0], tt.want) {
				t.Fatalf("problems: %v, want one containing %q", lines, tt.want)
			}
		})
	}
}

func TestDirMissingCitiesFile(t *testing.T) {
	_, lines := report(t, map[string]string{"events/.gitkeep": ""})
	if len(lines) != 1 || lines[0] != "cities.yaml: file is missing" {
		t.Fatalf("problems: %v", lines)
	}
}
