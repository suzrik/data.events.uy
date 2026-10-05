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
	maxEventDays   = 31 // longest gap between date and end_date
)

var (
	sports   = []string{"bike", "run", "roll"}
	statuses = []string{"confirmed", "tentative"}
	entries  = []string{"open", "license", "elite"}
)

type event struct {
	Name        string            `yaml:"name"`
	Date        string            `yaml:"date"`
	EndDate     string            `yaml:"end_date"`
	Status      string            `yaml:"status"`
	Entry       string            `yaml:"entry"`
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

// Messages for fields whose value has the wrong shape.
const (
	msgSingleValue  = "must be a single value on one line, not a list or nested lines"
	msgDistances    = "must be a list in square brackets, for example [10K, 5K]"
	msgDescription  = `must list the languages on separate indented lines, for example "es: ..."`
	msgLinks        = `must list the links on separate indented lines, for example "site: https://..."`
	msgSecondDoc    = `the file must contain one YAML document; remove the "---" line and everything after it`
	msgNotAMapping  = `the file must be a set of "field: value" lines, as in templates/event.yaml`
	msgDuplicateKey = "appears more than once; keep only one"
)

var eventFields = []string{"name", "date", "end_date", "status", "entry", "sport", "city", "venue", "distances", "links", "description"}
var linkFields = []string{"site", "instagram", "facebook", "register"}

// Event checks the contents of one event file. year is the name of the
// directory the file lives in; cities is the set read from cities.yaml, or
// nil when that list is unavailable (the city-list rule is then skipped).
func Event(path string, src []byte, year string, cities map[string]bool) []Problem {
	var ps []Problem
	add := func(field, format string, args ...any) {
		ps = append(ps, Problem{Path: path, Field: field, Msg: fmt.Sprintf(format, args...)})
	}

	dec := yaml.NewDecoder(bytes.NewReader(src))
	var doc yaml.Node
	if err := dec.Decode(&doc); err != nil {
		if errors.Is(err, io.EOF) {
			add("", "file is empty")
		} else {
			add("", "%s", cleanYAMLError(err))
		}
		return ps
	}
	var second yaml.Node
	if err := dec.Decode(&second); !errors.Is(err, io.EOF) && (err != nil || !isBlankDocument(&second)) {
		add("", "%s", msgSecondDoc)
	}

	root := doc.Content[0]
	if root.Kind == yaml.ScalarNode && root.Tag == "!!null" && root.Value == "" {
		add("", "file is empty")
		return ps
	}
	if root.Kind != yaml.MappingNode {
		add("", "%s", msgNotAMapping)
		return ps
	}

	// checkComment reports text that YAML drops after " #" on a value's line.
	checkComment := func(field string, n *yaml.Node) {
		if c := strings.TrimSpace(n.LineComment); c != "" {
			add(field, `the text after " #" is ignored by YAML (%q); put the whole value in double quotes, or remove the comment`, c)
		}
	}
	// scalar reads a one-line text value; ok is false (and a problem is
	// reported on field) when the value is a list or a mapping.
	scalar := func(field string, n *yaml.Node, dst *string) (ok bool) {
		n = resolveAlias(n)
		if n.Kind != yaml.ScalarNode || n.Decode(dst) != nil {
			add(field, "%s", msgSingleValue)
			return false
		}
		checkComment(field, n)
		return true
	}
	isNull := func(n *yaml.Node) bool {
		n = resolveAlias(n)
		return n.Kind == yaml.ScalarNode && n.Tag == "!!null"
	}

	var (
		e            event
		seen         = map[string]bool{}
		descShapeBad bool
		linksBad     bool
		descBad      = map[string]bool{}
		distBad      = map[int]bool{}
		shapeBad     = map[string]bool{} // scalar fields with a list or mapping value
	)
	for i := 0; i+1 < len(root.Content); i += 2 {
		key, val := root.Content[i].Value, root.Content[i+1]
		switch {
		case !slices.Contains(eventFields, key):
			add("", "unknown field %q", key)
			continue
		case seen[key]:
			add(key, "%s", msgDuplicateKey)
			continue
		}
		seen[key] = true
		val = resolveAlias(val)

		switch key {
		case "name":
			shapeBad[key] = !scalar(key, val, &e.Name)
		case "date":
			shapeBad[key] = !scalar(key, val, &e.Date)
		case "end_date":
			shapeBad[key] = !scalar(key, val, &e.EndDate)
		case "status":
			shapeBad[key] = !scalar(key, val, &e.Status)
		case "entry":
			shapeBad[key] = !scalar(key, val, &e.Entry)
		case "sport":
			shapeBad[key] = !scalar(key, val, &e.Sport)
		case "city":
			shapeBad[key] = !scalar(key, val, &e.City)
		case "venue":
			shapeBad[key] = !scalar(key, val, &e.Venue)
		case "distances":
			switch {
			case isNull(val):
			case val.Kind != yaml.SequenceNode:
				add(key, "%s", msgDistances)
				shapeBad[key] = true
			default:
				checkComment(key, val)
				e.Distances = []string{}
				for j, item := range val.Content {
					var d string
					if !scalar(fmt.Sprintf("distances[%d]", j), item, &d) {
						distBad[j] = true
					}
					e.Distances = append(e.Distances, d)
				}
			}
		case "links":
			switch {
			case isNull(val):
			case val.Kind != yaml.MappingNode:
				add(key, "%s", msgLinks)
				linksBad = true
			default:
				linkSeen := map[string]bool{}
				for j := 0; j+1 < len(val.Content); j += 2 {
					name, v := val.Content[j].Value, val.Content[j+1]
					var dst *string
					switch name {
					case "site":
						dst = &e.Links.Site
					case "instagram":
						dst = &e.Links.Instagram
					case "facebook":
						dst = &e.Links.Facebook
					case "register":
						dst = &e.Links.Register
					default:
						add("links", "unknown field %q", name)
						continue
					}
					if linkSeen[name] {
						add("links."+name, "%s", msgDuplicateKey)
						continue
					}
					linkSeen[name] = true
					if !scalar("links."+name, v, dst) {
						linksBad = true
					}
				}
			}
		case "description":
			switch {
			case isNull(val):
			case val.Kind != yaml.MappingNode:
				add(key, "%s", msgDescription)
				descShapeBad = true
			default:
				e.Description = map[string]string{}
				for j := 0; j+1 < len(val.Content); j += 2 {
					lang, v := val.Content[j].Value, val.Content[j+1]
					field := "description"
					if langKey.MatchString(lang) {
						field += "." + lang
					}
					if _, dup := e.Description[lang]; dup {
						add("description", "language %q appears more than once; keep only one", lang)
						continue
					}
					var text string
					if !scalar(field, v, &text) {
						descBad[lang] = true
					}
					e.Description[lang] = text
				}
			}
		}
	}
	if !shapeBad["name"] {
		switch n := utf8.RuneCountInString(strings.TrimSpace(e.Name)); {
		case n == 0:
			add("name", "is required")
		case n > maxName:
			add("name", "must be at most %d characters, got %d", maxName, n)
		}
	}

	var (
		start      time.Time // the date field, when it is a real day
		monthOnly  bool      // the date field is a real YYYY-MM
		dateIsGood bool      // the date field is a real day or a real month
	)
	if !shapeBad["date"] {
		day, dayErr := time.Parse("2006-01-02", e.Date)
		_, monthErr := time.Parse("2006-01", e.Date)
		switch {
		case e.Date == "":
			add("date", "is required")
		case dayErr != nil && monthErr != nil:
			add("date", "must be a real date written as YYYY-MM-DD, or YYYY-MM when the day is not announced yet, got %q", e.Date)
		default:
			dateIsGood = true
			if dayErr == nil {
				start = day
			} else {
				monthOnly = true
			}
			if e.Date[:4] != year {
				add("date", "year %s does not match the folder events/%s; move the file to events/%[1]s/", e.Date[:4], year)
			}
		}
	}

	if e.EndDate != "" && !shapeBad["end_date"] {
		end, err := time.Parse("2006-01-02", e.EndDate)
		switch {
		case err != nil:
			add("end_date", "must be a real date written as YYYY-MM-DD, got %q", e.EndDate)
		case dateIsGood && monthOnly:
			add("end_date", "needs a start date with a day: write date as YYYY-MM-DD, or remove end_date")
		case dateIsGood && !end.After(start):
			add("end_date", "must be later than date (%s); for a one-day event remove end_date", e.Date)
		case dateIsGood && end.After(start.AddDate(0, 0, maxEventDays)):
			add("end_date", "is more than %d days after date (%s); check the month and the year", maxEventDays, e.Date)
		}
	}

	if e.Status != "" && !shapeBad["status"] && !slices.Contains(statuses, e.Status) {
		add("status", "must be confirmed or tentative, got %q", e.Status)
	}

	if e.Entry != "" && !shapeBad["entry"] && !slices.Contains(entries, e.Entry) {
		msg := "must be one of " + strings.Join(entries, ", ") + ", got %q"
		if e.Entry == "licence" {
			msg += "; write it as license"
		}
		add("entry", msg, e.Entry)
	}

	if !shapeBad["sport"] {
		switch {
		case e.Sport == "":
			add("sport", "is required")
		case !slices.Contains(sports, e.Sport):
			add("sport", "must be one of %s, got %q", strings.Join(sports, ", "), e.Sport)
		}
	}

	if !shapeBad["city"] {
		switch {
		case e.City == "":
			add("city", "is required")
		case cities != nil && !cities[e.City]:
			add("city", "%q is not in cities.yaml; add it there in the same pull request", e.City)
		}
	}

	if n := utf8.RuneCountInString(e.Venue); !shapeBad["venue"] && n > maxVenue {
		add("venue", "must be at most %d characters, got %d", maxVenue, n)
	}

	if !shapeBad["distances"] {
		switch {
		case e.Distances != nil && len(e.Distances) == 0:
			add("distances", "must list at least one distance, or remove the field")
		case len(e.Distances) > maxDistances:
			add("distances", "must have at most %d values, got %d", maxDistances, len(e.Distances))
		}
		for i, d := range e.Distances {
			if distBad[i] {
				continue
			}
			field := fmt.Sprintf("distances[%d]", i)
			switch n := utf8.RuneCountInString(strings.TrimSpace(d)); {
			case n == 0:
				add(field, "must not be empty")
			case n > maxDistance:
				add(field, "must be at most %d characters, got %d", maxDistance, n)
			}
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
	if !linksBad && e.Links.Site == "" && e.Links.Instagram == "" && e.Links.Facebook == "" {
		add("links", "at least one of site, instagram, facebook is required: an official organizer page")
	}

	if descShapeBad {
		return ps
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
		if !langKey.MatchString(k) {
			add("description", "language code %q must be two lowercase letters", k)
			continue
		}
		if descBad[k] {
			continue
		}
		field := "description." + k
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

// isBlankDocument reports whether a YAML document has no content at all,
// as after a trailing "---" line.
func isBlankDocument(n *yaml.Node) bool {
	if len(n.Content) == 0 {
		return true
	}
	c := n.Content[0]
	return c.Kind == yaml.ScalarNode && c.Tag == "!!null" && c.Value == ""
}

func resolveAlias(n *yaml.Node) *yaml.Node {
	if n.Kind == yaml.AliasNode && n.Alias != nil {
		return n.Alias
	}
	return n
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
	msg := strings.Join(lines, "; ")
	hint := ""
	switch {
	case strings.Contains(msg, "mapping values are not allowed"):
		hint = ` — a value that contains ": " must be put in double quotes`
	case strings.Contains(msg, "cannot start any token"):
		hint = ` — a value that starts with "@" must be put in double quotes; links are full https:// addresses`
	}
	if hint != "" && !strings.HasPrefix(msg, "line ") {
		msg = "line 1: " + msg // yaml.v3 leaves the line out for line 1
	}
	return "invalid YAML: " + msg + hint
}
