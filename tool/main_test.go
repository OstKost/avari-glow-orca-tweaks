package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyReapplyRevert(t *testing.T) {
	p := filepath.Join(t.TempDir(), "app.css")
	original := []byte("body { color: red; }\n")
	if err := os.WriteFile(p, original, 0644); err != nil {
		t.Fatal(err)
	}
	if err := update(p, []string{"tabs", "aurora"}, false); err != nil {
		t.Fatal(err)
	}
	if err := update(p, []string{"composer"}, false); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(p)
	if strings.Count(string(data), start) != 1 || strings.Contains(string(data), "feature:tabs") || !strings.Contains(string(data), "feature:composer") {
		t.Fatalf("unexpected patch: %s", data)
	}
	if err := update(p, nil, true); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(p)
	if string(data) != string(original) {
		t.Fatalf("revert changed original: %q", data)
	}
	golden, err := os.ReadFile(p + ".original")
	if err != nil || string(golden) != string(original) {
		t.Fatalf("golden backup: %q %v", golden, err)
	}
	backups, _ := filepath.Glob(p + ".bak.*")
	if len(backups) != 3 {
		t.Fatalf("got %d backups", len(backups))
	}
}
func TestMalformedMarkersRefused(t *testing.T) {
	p := filepath.Join(t.TempDir(), "app.css")
	data := []byte("hello " + start)
	os.WriteFile(p, data, 0644)
	if update(p, []string{"tabs"}, false) == nil {
		t.Fatal("expected error")
	}
	got, _ := os.ReadFile(p)
	if string(got) != string(data) {
		t.Fatal("file changed")
	}
}
func TestUnknownFeatureRefused(t *testing.T) {
	if _, err := selectedCSS([]string{"unknown"}); err == nil {
		t.Fatal("expected error")
	}
}
