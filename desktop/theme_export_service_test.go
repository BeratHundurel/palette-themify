package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveThemeRejectsInvalidJSON(t *testing.T) {
	service := &ThemeExportService{}
	_, err := service.SaveThemeToEditorTarget("zed", "Broken", "{not json}")
	if err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Fatalf("expected invalid JSON error, got %v", err)
	}
}

func TestResolveZedThemeDirectoryFor(t *testing.T) {
	home := filepath.Join("users", "theme-user")
	tests := []struct {
		name     string
		goos     string
		env      map[string]string
		expected string
	}{
		{
			name:     "Windows APPDATA",
			goos:     "windows",
			env:      map[string]string{"APPDATA": filepath.Join("custom", "roaming")},
			expected: filepath.Join("custom", "roaming", "Zed", "themes"),
		},
		{
			name:     "Windows fallback",
			goos:     "windows",
			expected: filepath.Join(home, "AppData", "Roaming", "Zed", "themes"),
		},
		{
			name:     "macOS documented config directory",
			goos:     "darwin",
			expected: filepath.Join(home, ".config", "zed", "themes"),
		},
		{
			name:     "Linux XDG config directory",
			goos:     "linux",
			env:      map[string]string{"XDG_CONFIG_HOME": filepath.Join("custom", "config")},
			expected: filepath.Join("custom", "config", "zed", "themes"),
		},
		{
			name:     "Linux fallback",
			goos:     "linux",
			expected: filepath.Join(home, ".config", "zed", "themes"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lookupEnv := func(key string) (string, bool) {
				value, ok := test.env[key]
				return value, ok
			}
			actual, err := resolveZedThemeDirectoryFor(test.goos, home, lookupEnv)
			if err != nil {
				t.Fatal(err)
			}
			if actual != test.expected {
				t.Fatalf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestVSCodeSaveRollsBackThemeWhenManifestIsInvalid(t *testing.T) {
	home := t.TempDir()
	extensionDirectory := filepath.Join(home, ".vscode", "extensions", "themesmith-local")
	themesDirectory := filepath.Join(extensionDirectory, "themes")
	if err := os.MkdirAll(themesDirectory, 0o755); err != nil {
		t.Fatal(err)
	}

	themePath := filepath.Join(themesDirectory, "sample.json")
	oldTheme := []byte("old theme")
	if err := os.WriteFile(themePath, oldTheme, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(extensionDirectory, "package.json"), []byte("invalid"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := saveThemeToVSCodeAt(home, "Sample", `{"type":"dark"}`); err == nil {
		t.Fatal("expected invalid manifest error")
	}

	actual, err := os.ReadFile(themePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(actual) != string(oldTheme) {
		t.Fatalf("expected theme rollback to %q, got %q", oldTheme, actual)
	}
}

func TestVSCodeManifestReplacesEntriesThatShareAFileName(t *testing.T) {
	home := t.TempDir()
	if _, err := saveThemeToVSCodeAt(home, "A B", `{"type":"dark"}`); err != nil {
		t.Fatal(err)
	}
	if _, err := saveThemeToVSCodeAt(home, "A-B", `{"type":"light"}`); err != nil {
		t.Fatal(err)
	}

	packagePath := filepath.Join(home, ".vscode", "extensions", "themesmith-local", "package.json")
	content, err := os.ReadFile(packagePath)
	if err != nil {
		t.Fatal(err)
	}
	packageData := vscodePackageJSON{}
	if err := json.Unmarshal(content, &packageData); err != nil {
		t.Fatal(err)
	}
	if len(packageData.Contributes.Themes) != 1 {
		t.Fatalf("expected one manifest entry, got %d", len(packageData.Contributes.Themes))
	}
	if packageData.Contributes.Themes[0].Label != "A-B" {
		t.Fatalf("expected latest label, got %q", packageData.Contributes.Themes[0].Label)
	}
}

func TestWriteFileSafelyOverwritesReadOnlyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "theme.json")
	if err := os.WriteFile(path, []byte("old"), 0o400); err != nil {
		t.Fatal(err)
	}

	if err := writeFileSafely(path, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	actual, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(actual) != "new" {
		t.Fatalf("expected new content, got %q", actual)
	}
}

func TestFilesystemErrorIdentifiesPermissions(t *testing.T) {
	err := filesystemError("write theme file", "somewhere/theme.json", fs.ErrPermission)
	if !errors.Is(err, fs.ErrPermission) || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("expected permission error, got %v", err)
	}
}

func TestSanitizeThemeNameAvoidsProblematicFileNames(t *testing.T) {
	if actual := sanitizeThemeName("CON"); actual != "theme-con" {
		t.Fatalf("expected reserved name to be prefixed, got %q", actual)
	}

	actual := sanitizeThemeName(strings.Repeat("long name ", 20))
	if len([]rune(actual)) > 80 {
		t.Fatalf("expected at most 80 runes, got %d", len([]rune(actual)))
	}
	if actual != sanitizeThemeName(strings.Repeat("long name ", 20)) {
		t.Fatal("expected long-name shortening to be deterministic")
	}
}
