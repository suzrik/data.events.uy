package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const event = `name: Corrida Rambla 10K
date: 2026-10-11
sport: run
city: Montevideo
links:
  site: https://example.org
description:
  es: Recorrido plano.
`

func catalog(t *testing.T, eventBody string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "events", "2026"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "cities.yaml"), []byte("- Montevideo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "events", "2026", "corrida.yaml"), []byte(eventBody), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestRunValidCatalog(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(catalog(t, event), &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("code = %d, stderr = %q", code, stderr.String())
	}
	if got := stdout.String(); got != "OK: 1 event file(s) checked\n" {
		t.Errorf("stdout = %q", got)
	}
}

func TestRunInvalidCatalog(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(catalog(t, strings.Replace(event, "sport: run", "sport: swim", 1)), &stdout, &stderr)
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	got := stderr.String()
	if !strings.Contains(got, "events/2026/corrida.yaml: sport: must be one of road, mtb, gravel, run, roll") {
		t.Errorf("stderr lacks the problem line: %q", got)
	}
	if !strings.Contains(got, "1 problem(s) found in 1 event file(s) checked") {
		t.Errorf("stderr lacks the summary: %q", got)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunWrongDirectory(t *testing.T) {
	var stdout, stderr bytes.Buffer
	dir := filepath.Join(t.TempDir(), "nope")
	if code := run(dir, &stdout, &stderr); code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
	if want := "validate: catalog directory \"" + dir + "\" not found\n"; stderr.String() != want {
		t.Errorf("stderr = %q, want %q", stderr.String(), want)
	}
}

func TestRunArgsOptionIsUsageError(t *testing.T) {
	for _, arg := range []string{"--help", "-h", "-x"} {
		var stdout, stderr bytes.Buffer
		if code := runArgs([]string{arg}, &stdout, &stderr); code != 2 {
			t.Errorf("%s: code = %d, want 2", arg, code)
		}
		if got := stderr.String(); got != "usage: validate [catalog directory]\n" {
			t.Errorf("%s: stderr = %q", arg, got)
		}
		if stdout.Len() != 0 {
			t.Errorf("%s: stdout = %q", arg, stdout.String())
		}
	}
	var stdout, stderr bytes.Buffer
	if code := runArgs([]string{catalog(t, event)}, &stdout, &stderr); code != 0 {
		t.Errorf("directory argument: code = %d, stderr = %q", code, stderr.String())
	}
}
