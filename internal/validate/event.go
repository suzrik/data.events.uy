// Package validate checks the events catalog against the format rules.
// It is the source of truth for the event file format.
package validate

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"go.yaml.in/yaml/v3"
)

// Problem is one rule violation in one file.
type Problem struct {
	Path  string // slash-separated, relative to the catalog root
	Field string // empty when the problem concerns the whole file
	Msg   string
}

func (p Problem) String() string {
	if p.Field == "" {
		return p.Path + ": " + p.Msg
	}
	return p.Path + ": " + p.Field + ": " + p.Msg
}

const (
	maxName        = 120
	maxVenue       = 120
	maxDistances   = 6
	maxDistance    = 20
	maxDescription = 500
)

var sports = []string{"bike", "run", "roll"}

type event struct {
	Name        string            `yaml:"name"`
	Date        string            `yaml:"date"`
	Sport       string            `yaml:"sport"`
	City        string            `yaml:"city"`
	Venue       string            `yaml:"venue"`
	Distances   []string          `yaml:"distances"`
	Links       links             `yaml:"links"`
	Description map[string]string `yaml:"description"`
}

type links struct {
	Site      string `yaml:"site"`
	Instagram string `yaml:"instagram"`
	Facebook  string `yaml:"facebook"`
	Register  string `yaml:"register"`
}

var (
	langKey   = regexp.MustCompile(`^[a-z]{2}$`)
	goTypeRef = regexp.MustCompile(` in type \S+`)
)

// Event checks the contents of one event file. year is the name of the
// directory the file lives in; cities is the set read from cities.yaml.
func Event(path string, src []byte, year string, cities map[string]bool) []Problem {
	var ps []Problem
	add := func(field, format string, args ...any) {
		ps = append(ps, Problem{Path: path, Field: field, Msg: fmt.Sprintf(format, args...)})
	}

	var e event
	dec := yaml.NewDecoder(bytes.NewReader(src))
	dec.KnownFields(true)
	if err := dec.Decode(&e); err != nil {
		if errors.Is(err, io.EOF) {
			add("", "file is empty")
		} else {
			add("", "%s", cleanYAMLError(err))
		}
		return ps
	}

	switch n := utf8.RuneCountInString(strings.TrimSpace(e.Name)); {
	case n == 0:
		add("name", "is required")
	case n > maxName:
		add("name", "must be at most %d characters, got %d", maxName, n)
	}

	switch _, err := time.Parse("2006-01-02", e.Date); {
	case e.Date == "":
		add("date", "is required")
	case err != nil:
		add("date", "must be a real date written as YYYY-MM-DD, got %q", e.Date)
	case e.Date[:4] != year:
		add("date", "year %s does not match the folder events/%s", e.Date[:4], year)
	}

	switch {
	case e.Sport == "":
		add("sport", "is required")
	case !slices.Contains(sports, e.Sport):
		add("sport", "must be one of %s, got %q", strings.Join(sports, ", "), e.Sport)
	}

	switch {
	case e.City == "":
		add("city", "is required")
	case !cities[e.City]:
		add("city", "%q is not in cities.yaml; add it there in the same pull request", e.City)
	}

	if n := utf8.RuneCountInString(e.Venue); n > maxVenue {
		add("venue", "must be at most %d characters, got %d", maxVenue, n)
	}

	switch {
	case e.Distances != nil && len(e.Distances) == 0:
		add("distances", "must list at least one distance, or remove the field")
	case len(e.Distances) > maxDistances:
		add("distances", "must have at most %d values, got %d", maxDistances, len(e.Distances))
	}
	for i, d := range e.Distances {
		field := fmt.Sprintf("distances[%d]", i)
		switch n := utf8.RuneCountInString(strings.TrimSpace(d)); {
		case n == 0:
			add(field, "must not be empty")
		case n > maxDistance:
			add(field, "must be at most %d characters, got %d", maxDistance, n)
		}
	}

	for _, l := range []struct{ field, value string }{
		{"links.site", e.Links.Site},
		{"links.instagram", e.Links.Instagram},
		{"links.facebook", e.Links.Facebook},
		{"links.register", e.Links.Register},
	} {
		if l.value != "" && !isHTTPS(l.value) {
			add(l.field, "must be an https:// URL, got %q", l.value)
		}
	}
	if e.Links.Site == "" && e.Links.Instagram == "" && e.Links.Facebook == "" {
		add("links", "at least one of site, instagram, facebook is required: an official organizer page")
	}

	if len(e.Description) == 0 {
		add("description", "is required")
		return ps
	}
	langs := make([]string, 0, len(e.Description))
	for k := range e.Description {
		langs = append(langs, k)
	}
	sort.Strings(langs)
	for _, k := range langs {
		field := "description." + k
		if !langKey.MatchString(k) {
			add(field, "language code must be two lowercase letters")
			continue
		}
		switch n := utf8.RuneCountInString(strings.TrimSpace(e.Description[k])); {
		case n == 0:
			add(field, "must not be empty")
		case n > maxDescription:
			add(field, "must be at most %d characters, got %d", maxDescription, n)
		}
	}
	if _, es := e.Description["es"]; !es {
		if _, en := e.Description["en"]; !en {
			add("description", "must include es or en")
		}
	}
	return ps
}

func isHTTPS(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme == "https" && u.Host != ""
}

// cleanYAMLError turns a yaml.v3 error into one line without Go type names.
func cleanYAMLError(err error) string {
	s := strings.TrimPrefix(err.Error(), "yaml: ")
	s = strings.TrimPrefix(s, "unmarshal errors:\n")
	s = goTypeRef.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "validate.", "")
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	return "invalid YAML: " + strings.Join(lines, "; ")
}
