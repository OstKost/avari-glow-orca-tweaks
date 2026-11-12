package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBundledLoaderRoundTrip(t *testing.T) {
	base, err := os.MkdirTemp("/tmp", "at-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(base)
	root := filepath.Join(base, "app.asar.unpacked", "out", "renderer")
	assets := filepath.Join(root, "assets")
	if err := os.MkdirAll(assets, 0755); err != nil {
		t.Fatal(err)
	}
	cssPath := filepath.Join(assets, "App-test.css")
	originalCSS := []byte("body{color:red}\n")
	index := []byte(`<head><!-- CSP is relaxed during development; electron-vite injects a stricter policy for production builds --><link rel="modulepreload" href="./assets/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.js"><link rel="modulepreload" href="./assets/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb.js"><link href="./assets/App-test.css"></head>` + strings.Repeat(" ", 1000))
	indexPath := filepath.Join(root, "index.html")
	if err := os.WriteFile(cssPath, originalCSS, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(indexPath, index, 0644); err != nil {
		t.Fatal(err)
	}
	if err := update(cssPath, []string{"tabs", "panels"}, false); err != nil {
		t.Fatal(err)
	}
	if err := update(cssPath, nil, true); err != nil {
		t.Fatal(err)
	}
	gotCSS, _ := os.ReadFile(cssPath)
	gotIndex, _ := os.ReadFile(indexPath)
	if !bytes.Equal(gotCSS, originalCSS) || !bytes.Equal(gotIndex, index) {
		t.Fatal("revert did not restore fixture")
	}
	if _, err := os.Stat(filepath.Join(assets, "avari-tilt.js")); !os.IsNotExist(err) {
		t.Fatal("tilt script was not removed")
	}
}
