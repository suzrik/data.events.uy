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
	Rides    int // ride files checked
	Problems []Problem
}

const maxSlug = 80

var (
	eventPath = regexp.MustCompile(`^events/(\d{4})/([a-z0-9]+(?:-[a-z0-9]+)*)\.yaml$`)
	ridePath  = regexp.MustCompile(`^rides/([a-z0-9]+(?:-[a-z0-9]+)*)\.yaml$`)
)

// Dir checks the catalog rooted at root: cities.yaml and every file under
// events/ and rides/. The error is non-nil only when the file system cannot
// be read.
func Dir(root string) (Report, error) {
	var rep Report
	if info, err := os.Stat(root); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return rep, err
	} else if err != nil || !info.IsDir() {
		return rep, fmt.Errorf("catalog directory %q not found", root)
	}
	fsys := os.DirFS(root)

	cities, ps, err := loadCities(fsys)
	if err != nil {
		return rep, err
	}
	rep.Problems = append(rep.Problems, ps...)

	// walk checks every file under dir; each receives the path and its content.
	walk := func(dir string, each func(p string, src []byte)) error {
		return fs.WalkDir(fsys, dir, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				if p == dir && errors.Is(err, fs.ErrNotExist) {
					return nil // a catalog without this folder is valid
				}
				return err
			}
			if p != dir && strings.HasPrefix(d.Name(), ".") {
				if d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
			if d.Type()&fs.ModeSymlink != 0 {
				rep.Problems = append(rep.Problems, Problem{Path: p, Msg: "must be a regular file, not a symbolic link"})
				return nil
			}
			if d.IsDir() {
				return nil
			}
			if !d.Type().IsRegular() {
				rep.Problems = append(rep.Problems, Problem{Path: p, Msg: "must be a regular file"})
				return nil
			}
			src, err := fs.ReadFile(fsys, p)
			if err != nil {
				return err
			}
			each(p, src)
			return nil
		})
	}

	err = walk("events", func(p string, src []byte) {
		m := eventPath.FindStringSubmatch(p)
		if m == nil || len(m[2]) > maxSlug {
			rep.Problems = append(rep.Problems, Problem{Path: p, Msg: fmt.Sprintf(
				"must be events/<year>/<slug>.yaml; the slug uses lowercase letters, digits and hyphens, up to %d characters", maxSlug)})
			return
		}
		rep.Events++
		rep.Problems = append(rep.Problems, Event(p, src, m[1], cities)...)
	})
	if err == nil {
		err = walk("rides", func(p string, src []byte) {
			m := ridePath.FindStringSubmatch(p)
			if m == nil || len(m[1]) > maxSlug {
				rep.Problems = append(rep.Problems, Problem{Path: p, Msg: fmt.Sprintf(
					"must be rides/<slug>.yaml; the slug uses lowercase letters, digits and hyphens, up to %d characters", maxSlug)})
				return
			}
			rep.Rides++
			rep.Problems = append(rep.Problems, Ride(p, src, cities)...)
		})
	}
	sort.SliceStable(rep.Problems, func(i, j int) bool { return rep.Problems[i].Path < rep.Problems[j].Path })
	return rep, err
}

// loadCities reads cities.yaml. The returned set is nil when the list is
// unavailable (missing or not a valid list), so that events are not also
// reported for cities that cannot be checked. The error is non-nil only
// when the file exists but cannot be read.
func loadCities(fsys fs.FS) (map[string]bool, []Problem, error) {
	const path = "cities.yaml"
	src, err := fs.ReadFile(fsys, path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, []Problem{{Path: path, Msg: "file is missing"}}, nil
	}
	if err != nil {
		return nil, nil, err
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(src, &doc); err != nil {
		return nil, []Problem{{Path: path, Msg: cleanYAMLError(err)}}, nil
	}
	set := map[string]bool{}
	if len(doc.Content) == 0 {
		return set, nil, nil // an empty file is an empty list
	}
	root := doc.Content[0]
	if root.Kind != yaml.SequenceNode {
		// Let the decoder describe what is wrong with the shape.
		var list []string
		if err := yaml.Unmarshal(src, &list); err != nil {
			return nil, []Problem{{Path: path, Msg: cleanYAMLError(err)}}, nil
		}
		return set, nil, nil // an empty document such as "~"
	}
	var ps []Problem
	for i, item := range root.Content {
		field := fmt.Sprintf("[%d]", i)
		var c string
		if item.Kind != yaml.ScalarNode || item.Decode(&c) != nil {
			ps = append(ps, Problem{Path: path, Field: field, Msg: "must be a city name"})
			continue
		}
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
	return set, ps, nil
}
