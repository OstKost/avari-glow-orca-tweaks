package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//go:embed avari-tilt.js
var tiltJS []byte

// The asar protocol reports the archived length of unpacked assets. Point the
// existing stylesheet link at its real file so Electron reads the complete CSS.
func loader(path string, revert bool) error {
	marker := string(filepath.Separator) + "app.asar.unpacked" + string(filepath.Separator)
	if !strings.Contains(path, marker) {
		return nil
	} // Explicit standalone CSS fixtures.
	index := filepath.Join(filepath.Dir(filepath.Dir(path)), "index.html")
	data, err := os.ReadFile(index)
	if err != nil {
		return err
	}
	info, err := os.Stat(index)
	if err != nil {
		return err
	}
	old := []byte(`href="./assets/` + filepath.Base(path) + `"`)
	u := url.URL{Scheme: "file", Path: path}
	replacement := []byte(`href="` + u.String() + `"`)
	sentinel := []byte("<!-- CSP is relaxed during development; electron-vite injects a stricter policy for production builds -->")
	if revert {
		if !bytes.Contains(data, replacement) {
			return nil
		}
		original, err := os.ReadFile(index + ".avari-original")
		if err != nil {
			return fmt.Errorf("loader backup: %w", err)
		}
		expected, err := makeLoader(original, old, replacement, sentinel)
		if err != nil {
			return err
		}
		withScript, err := addTiltLoader(expected, path)
		if err != nil {
			return err
		}
		if !bytes.Equal(expected, data) && !bytes.Equal(withScript, data) {
			return fmt.Errorf("index.html changed since Avari was applied; refusing to overwrite it")
		}
		if err := os.WriteFile(index, original, info.Mode().Perm()); err != nil {
			return err
		}
		if err := os.Remove(filepath.Join(filepath.Dir(path), "avari-tilt.js")); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	if bytes.Contains(data, replacement) {
		original, err := os.ReadFile(index + ".avari-original")
		if err != nil {
			return err
		}
		expected, err := makeLoader(original, old, replacement, sentinel)
		if err != nil {
			return err
		}
		next, err := addTiltLoader(expected, path)
		if err != nil {
			return err
		}
		if !bytes.Equal(data, expected) && !bytes.Equal(data, next) {
			return fmt.Errorf("index.html changed since Avari was applied; refusing to overwrite it")
		}
		if err := writeTilt(path); err != nil {
			return err
		}
		if bytes.Equal(data, next) {
			return nil
		}
		return os.WriteFile(index, next, info.Mode().Perm())
	}
	next, err := makeLoader(data, old, replacement, sentinel)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(index+".avari-original", os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode().Perm())
	if os.IsExist(err) {
		b, e := os.ReadFile(index + ".avari-original")
		if e != nil || !bytes.Equal(b, data) {
			return fmt.Errorf("existing index backup differs; refusing to overwrite it")
		}
	} else if err != nil {
		return err
	} else {
		_, err = f.Write(data)
		if e := f.Close(); err == nil {
			err = e
		}
		if err != nil {
			return err
		}
	}
	next, err = addTiltLoader(next, path)
	if err != nil {
		return err
	}
	if err = writeTilt(path); err != nil {
		return err
	}
	return os.WriteFile(index, next, info.Mode().Perm())
}
func makeLoader(data, old, replacement, sentinel []byte) ([]byte, error) {
	if bytes.Count(data, old) != 1 || bytes.Count(data, sentinel) != 1 {
		return nil, fmt.Errorf("unsupported index.html layout; no files changed")
	}
	out := bytes.Replace(data, old, replacement, 1)
	out = bytes.Replace(out, sentinel, nil, 1)
	if len(out) > len(data) {
		return nil, fmt.Errorf("insufficient space in index.html")
	}
	return append(out, bytes.Repeat([]byte(" "), len(data)-len(out))...), nil
}

func writeTilt(path string) error {
	return os.WriteFile(filepath.Join(filepath.Dir(path), "avari-tilt.js"), tiltJS, 0644)
}
func addTiltLoader(data []byte, path string) ([]byte, error) {
	re := regexp.MustCompile(`<link rel="modulepreload"[^>]*>`)
	matches := re.FindAllIndex(data, 2)
	if len(matches) < 2 {
		return nil, fmt.Errorf("cannot safely insert tilt loader")
	}
	u := url.URL{Scheme: "file", Path: filepath.Join(filepath.Dir(path), "avari-tilt.js")}
	tag := []byte(`<script defer src="` + u.String() + `"></script>`)
	out := append([]byte{}, data[:matches[0][0]]...)
	out = append(out, tag...)
	out = append(out, data[matches[1][1]:]...)
	if len(out) > len(data) {
		return nil, fmt.Errorf("insufficient index space for tilt loader")
	}
	return append(out, bytes.Repeat([]byte(" "), len(data)-len(out))...), nil
}
