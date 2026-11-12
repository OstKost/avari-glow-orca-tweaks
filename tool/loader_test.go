package main

import (
	"bytes"
	"testing"
)

func TestMakeLoaderPreservesLength(t *testing.T) {
	original := []byte(`<head><!-- CSP is relaxed during development; electron-vite injects a stricter policy for production builds --><link href="./assets/App-test.css"></head>`)
	old := []byte(`href="./assets/App-test.css"`)
	next := []byte(`href="file:///tmp/example/app.asar.unpacked/out/renderer/assets/App-test.css"`)
	sentinel := []byte(`<!-- CSP is relaxed during development; electron-vite injects a stricter policy for production builds -->`)
	got, err := makeLoader(original, old, next, sentinel)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(original) || !bytes.Contains(got, next) {
		t.Fatalf("invalid loader result: %s", got)
	}
	if _, err = makeLoader([]byte("unexpected"), old, next, sentinel); err == nil {
		t.Fatal("expected unsupported layout error")
	}
}
