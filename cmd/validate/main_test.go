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
	if !strings.Contains(got, "events/2026/corrida.yaml: sport: must be one of bike, run, roll") {
		t.Errorf("stderr lacks the problem line: %q", got)
	}
	if !strings.Contains(got, "1 problem(s) found in 1 event file(s) checked") {
		t.Errorf("stderr lacks the summary: %q", got)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}
}
