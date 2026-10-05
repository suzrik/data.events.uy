package validate

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

const repoRoot = "../.."

// fieldNames lists every field of the event format as the guides spell it:
// name, date, ..., links.site, links.instagram, ...
func fieldNames() []string {
	var names []string
	t := reflect.TypeOf(event{})
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("yaml")
		if f.Type.Kind() != reflect.Struct {
			names = append(names, tag)
			continue
		}
		for j := 0; j < f.Type.NumField(); j++ {
			names = append(names, tag+"."+f.Type.Field(j).Tag.Get("yaml"))
		}
	}
	return names
}

func TestRepositoryCatalogIsValid(t *testing.T) {
	rep, err := Dir(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range rep.Problems {
		t.Error(p)
	}
}

func TestTemplateIsAValidEvent(t *testing.T) {
	src, err := os.ReadFile(repoRoot + "/templates/event.yaml")
	if err != nil {
		t.Fatal(err)
	}
	cities, ps, err := loadCities(os.DirFS(repoRoot))
	if err != nil || len(ps) != 0 {
		t.Fatalf("cities.yaml: %v %v", ps, err)
	}
	// The template's example date is in 2026.
	for _, p := range Event("templates/event.yaml", src, "2026", cities) {
		t.Error(p)
	}
	for _, name := range fieldNames() {
		key := name[strings.LastIndex(name, ".")+1:] + ":"
		if !strings.Contains(string(src), key) {
			t.Errorf("template has no %q", key)
		}
	}
}

// The template travels inside a GitHub new-file URL, so it must stay short.
func TestTemplateFitsInAURL(t *testing.T) {
	src, err := os.ReadFile(repoRoot + "/templates/event.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if len(src) > 3000 {
		t.Errorf("template is %d bytes; keep it at 3000 or fewer", len(src))
	}
}

func TestGuidesDescribeEveryField(t *testing.T) {
	for _, lang := range []string{"es", "en"} {
		src, err := os.ReadFile(repoRoot + "/docs/add-event." + lang + ".md")
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range fieldNames() {
			if !strings.Contains(string(src), "`"+name+"`") {
				t.Errorf("add-event.%s.md does not describe `%s`", lang, name)
			}
		}
	}
}

func TestLicensesArePresent(t *testing.T) {
	for file, want := range map[string]string{
		"LICENSE":      "Attribution 4.0 International",
		"LICENSE-CODE": "MIT License",
	} {
		src, err := os.ReadFile(repoRoot + "/" + file)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(src), want) {
			t.Errorf("%s does not contain %q", file, want)
		}
	}
}
