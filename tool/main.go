package main

import (
	"bytes"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

//go:embed avari-tweaks.css
var css string

const start = "/* === AVARI GLOW TWEAKS START [v1.0] === */"
const end = "/* === AVARI GLOW TWEAKS END === */"

var features = []string{"tabs", "aurora", "panels", "composer", "badges", "chips"}

func detect(app string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(app, "Contents/Resources/app.asar.unpacked/out/renderer/assets/App-*.css"))
	if err != nil {
		return "", err
	}
	if len(matches) != 1 {
		return "", fmt.Errorf("expected exactly one App-*.css in %s; found %d", app, len(matches))
	}
	return matches[0], nil
}
func strip(data []byte) ([]byte, bool, error) {
	s := string(data)
	a, b := strings.Count(s, start), strings.Count(s, end)
	if a == 0 && b == 0 {
		return data, false, nil
	}
	if a != 1 || b != 1 {
		return nil, false, errors.New("invalid or duplicate Avari markers; file left unchanged")
	}
	i, j := strings.Index(s, start), strings.Index(s, end)
	if j < i {
		return nil, false, errors.New("reversed Avari markers; file left unchanged")
	}
	j += len(end)
	if j < len(s) && s[j] == '\n' {
		j++
	}
	if i > 0 && s[i-1] == '\n' {
		i--
	}
	return []byte(s[:i] + s[j:]), true, nil
}
func selectedCSS(names []string) (string, error) {
	wanted := map[string]bool{}
	for _, n := range names {
		found := false
		for _, f := range features {
			if n == f {
				found = true
			}
		}
		if !found {
			return "", fmt.Errorf("unknown feature %q", n)
		}
		wanted[n] = true
	}
	var out strings.Builder
	chunks := strings.Split(css, "/* feature:")
	for _, chunk := range chunks[1:] {
		parts := strings.SplitN(chunk, " */", 2)
		if len(parts) == 2 && wanted[parts[0]] {
			out.WriteString("/* feature:" + chunk)
		}
	}
	return out.String(), nil
}
func backup(path string, data []byte, mode os.FileMode) error {
	f, err := os.OpenFile(path+".original", os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode.Perm())
	if err == nil {
		if _, err = f.Write(data); err != nil {
			f.Close()
			return err
		}
		if err = f.Close(); err != nil {
			return err
		}
	} else if !os.IsExist(err) {
		return err
	}
	f, err = os.OpenFile(path+".bak."+time.Now().UTC().Format("20060102T150405.000000000Z"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode.Perm())
	if err != nil {
		return err
	}
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
func update(path string, names []string, revert bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	clean, exists, err := strip(data)
	if err != nil {
		return err
	}
	var next []byte
	if revert {
		if !exists {
			return nil
		}
		next = clean
	} else {
		body, err := selectedCSS(names)
		if err != nil {
			return err
		}
		if len(names) == 0 {
			return errors.New("select at least one feature")
		}
		next = append(append([]byte{}, clean...), []byte("\n"+start+"\n"+body+end+"\n")...)
	}
	if bytes.Equal(data, next) {
		return loader(path, revert)
	}
	if err = backup(path, data, info.Mode()); err != nil {
		return fmt.Errorf("backup: %w", err)
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".avari-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if err = f.Chmod(info.Mode().Perm()); err == nil {
		_, err = f.Write(next)
	}
	if err == nil {
		err = f.Sync()
	}
	cerr := f.Close()
	if err != nil {
		return err
	}
	if cerr != nil {
		return cerr
	}
	if err = os.Rename(tmp, path); err != nil {
		return err
	}
	if err := loader(path, revert); err != nil {
		// Avoid leaving a CSS patch behind when the loader cannot be updated.
		if restoreErr := os.WriteFile(path, data, info.Mode().Perm()); restoreErr != nil {
			return fmt.Errorf("loader: %v; CSS restoration failed: %w", err, restoreErr)
		}
		return err
	}
	return nil
}
func installTheme(path string) error {
	cmd := exec.Command("orca", "plugin", "dev", "load", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
func status(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return err.Error()
	}
	_, ok, err := strip(data)
	if err != nil {
		return err.Error()
	}
	if !ok {
		return "not applied"
	}
	var names []string
	for _, n := range features {
		if strings.Contains(string(data), "/* feature:"+n+" */") {
			names = append(names, n)
		}
	}
	return "written to disk (not verified in Orca): " + strings.Join(names, ", ")
}

var menuItems = []string{
	"Выбрать визуальные твики", "Применить выбранные твики", "Отменить CSS-твики",
	"Загрузить тему", "Установить иконку Dock", "Вернуть исходную иконку", "Выход",
}
var featureLabels = map[string]string{
	"tabs": "Вкладки", "aurora": "Свечение Aurora", "panels": "Панели",
	"composer": "Поле ввода", "badges": "Бейджи", "chips": "Метки",
}

type model struct {
	path, theme, app, msg string
	cursor                int
	choosing              bool
	chosen                map[string]bool
}

func (m model) Init() tea.Cmd { return nil }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	v, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch v.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		if m.choosing {
			m.choosing = false
			m.cursor = 0
			m.msg = ""
		}
	case "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down":
		limit := len(menuItems)
		if m.choosing {
			limit = len(features) + 1
		}
		if m.cursor < limit-1 {
			m.cursor++
		}
	case "enter", " ":
		if m.choosing {
			if m.cursor == len(features) {
				m.choosing = false
				m.cursor = 0
			} else {
				n := features[m.cursor]
				m.chosen[n] = !m.chosen[n]
			}
			return m, nil
		}
		switch m.cursor {
		case 0:
			m.choosing = true
			m.cursor = 0
			m.msg = ""
		case 1:
			var names []string
			for _, f := range features {
				if m.chosen[f] {
					names = append(names, f)
				}
			}
			if e := update(m.path, names, false); e != nil {
				m.msg = e.Error()
			} else {
				m.msg = "Твики применены. Перезапустите Orca."
			}
		case 2:
			if e := update(m.path, nil, true); e != nil {
				m.msg = e.Error()
			} else {
				m.msg = "Твики отменены. Перезапустите Orca."
			}
		case 3:
			cmd := exec.Command("orca", "plugin", "dev", "load", m.theme)
			out, e := cmd.CombinedOutput()
			if e != nil {
				m.msg = string(out) + e.Error()
			} else {
				m.msg = "Тема загружена."
			}
		case 4:
			if e := changeIcon(m.app, false); e != nil {
				m.msg = e.Error()
			} else {
				m.msg = "Иконка установлена. Перезапустите Orca; macOS может показывать старую иконку из кэша."
			}
		case 5:
			if e := changeIcon(m.app, true); e != nil {
				m.msg = e.Error()
			} else {
				m.msg = "Исходная иконка восстановлена. Перезапустите Orca."
			}
		case 6:
			return m, tea.Quit
		}
	}
	return m, nil
}
func (m model) View() string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("AVARI ORCA THEME TWEAKER"))
	fmt.Fprintf(&b, "\nCSS: %s\nСтатус: %s\nИконка: %s\n\n", m.path, status(m.path), iconStatus(m.app))
	if m.choosing {
		b.WriteString("Выберите твики:\n\n")
		for i, n := range features {
			cursor, mark := " ", " "
			if i == m.cursor {
				cursor = ">"
			}
			if m.chosen[n] {
				mark = "x"
			}
			fmt.Fprintf(&b, "%s [%s] %s\n", cursor, mark, featureLabels[n])
		}
		cursor := " "
		if m.cursor == len(features) {
			cursor = ">"
		}
		fmt.Fprintf(&b, "%s Назад в меню\n", cursor)
		b.WriteString("\n↑/↓ выбор • Enter отметить • Esc назад\n")
	} else {
		for i, item := range menuItems {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}
			fmt.Fprintf(&b, "%s %s\n", cursor, item)
		}
		b.WriteString("\n↑/↓ выбор • Enter выполнить\n")
	}
	if m.msg != "" {
		fmt.Fprintf(&b, "\n%s\n", m.msg)
	}
	return b.String()
}
func run() error {
	app := os.Getenv("AVARI_ORCA_APP")
	if app == "" {
		app = "/Applications/Orca.app"
	}
	args := os.Args[1:]
	command := "tui"
	if len(args) > 0 {
		command = args[0]
		args = args[1:]
	}
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	appPath := fs.String("app", app, "path to Orca.app")
	cssPath := fs.String("css", "", "explicit CSS path (testing or custom installations)")
	theme := fs.String("theme", "../avari-glow-theme", "theme plugin directory")
	selection := fs.String("features", strings.Join(features, ","), "comma-separated features")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return errors.New("unexpected arguments")
	}
	if command == "icon-apply" {
		return changeIcon(*appPath, false)
	}
	if command == "icon-revert" {
		return changeIcon(*appPath, true)
	}
	if command == "install-theme" {
		return installTheme(*theme)
	}
	path := *cssPath
	var err error
	if path == "" {
		path, err = detect(*appPath)
		if err != nil {
			return err
		}
	}
	switch command {
	case "status":
		fmt.Printf("CSS: %s\nStatus: %s\nIcon: %s\n", path, status(path), iconStatus(*appPath))
		return nil
	case "apply":
		names := strings.Split(*selection, ",")
		sort.Strings(names)
		return update(path, names, false)
	case "revert":
		return update(path, nil, true)
	case "tui":
		chosen := map[string]bool{}
		for _, f := range features {
			chosen[f] = true
		}
		_, err = tea.NewProgram(model{path: path, theme: *theme, app: *appPath, chosen: chosen}).Run()
		return err
	default:
		return fmt.Errorf("unknown command %q (use apply, revert, status, tui, install-theme, icon-apply or icon-revert)", command)
	}
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "avari-orca-tweak:", err)
		os.Exit(1)
	}
}

// Dock icon replacement is not included in the catalog edition.
func iconStatus(string) string { return "not included in catalog edition" }
func changeIcon(string, bool) error {
	return errors.New("Dock icon replacement is not included in catalog edition")
}
