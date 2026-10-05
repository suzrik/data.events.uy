package validate

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Report is the result of checking a whole catalog.
type Report struct {
	Events   int // event files checked
	Problems []Problem
}

const maxSlug = 80

var eventPath = regexp.MustCompile(`^events/(\d{4})/([a-z0-9]+(?:-[a-z0-9]+)*)\.yaml$`)

// Dir checks the catalog rooted at root: cities.yaml and every file under
// events/. The error is non-nil only when the file system cannot be read.
func Dir(root string) (Report, error) {
	var rep Report
	fsys := os.DirFS(root)

	cities, ps := loadCities(fsys)
	rep.Problems = append(rep.Problems, ps...)

	err := fs.WalkDir(fsys, "events", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if p == "events" && errors.Is(err, fs.ErrNotExist) {
				return nil // a catalog without events is valid
			}
			return err
		}
		if p != "events" && strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		m := eventPath.FindStringSubmatch(p)
		if m == nil || len(m[2]) > maxSlug {
			rep.Problems = append(rep.Problems, Problem{Path: p, Msg: fmt.Sprintf(
				"must be events/<year>/<slug>.yaml; the slug uses lowercase letters, digits and hyphens, up to %d characters", maxSlug)})
			return nil
		}
		src, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		rep.Events++
		rep.Problems = append(rep.Problems, Event(p, src, m[1], cities)...)
		return nil
	})
	sort.SliceStable(rep.Problems, func(i, j int) bool { return rep.Problems[i].Path < rep.Problems[j].Path })
	return rep, err
}

func loadCities(fsys fs.FS) (map[string]bool, []Problem) {
	const path = "cities.yaml"
	set := map[string]bool{}
	src, err := fs.ReadFile(fsys, path)
	if err != nil {
		return set, []Problem{{Path: path, Msg: "file is missing"}}
	}
	var list []string
	if err := yaml.Unmarshal(src, &list); err != nil {
		return set, []Problem{{Path: path, Msg: cleanYAMLError(err)}}
	}
	var ps []Problem
	for i, c := range list {
		field := fmt.Sprintf("[%d]", i)
		switch {
		case strings.TrimSpace(c) == "":
			ps = append(ps, Problem{Path: path, Field: field, Msg: "must not be empty"})
			continue
		case c != strings.TrimSpace(c):
			ps = append(ps, Problem{Path: path, Field: field, Msg: "must not start or end with a space"})
		case set[c]:
			ps = append(ps, Problem{Path: path, Field: field, Msg: fmt.Sprintf("%q is listed twice", c)})
		}
		set[c] = true
	}
	return set, ps
}
