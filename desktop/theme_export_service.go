package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"
)

type ThemeExportService struct{}

type vscodePackageJSON struct {
	Name        string                       `json:"name"`
	DisplayName string                       `json:"displayName"`
	Description string                       `json:"description"`
	Version     string                       `json:"version"`
	Publisher   string                       `json:"publisher"`
	Engines     map[string]string            `json:"engines"`
	Contributes vscodePackageJSONContributes `json:"contributes"`
}

type vscodePackageJSONContributes struct {
	Themes []vscodeThemeEntry `json:"themes"`
}

type vscodeThemeEntry struct {
	Label   string `json:"label"`
	UITheme string `json:"uiTheme"`
	Path    string `json:"path"`
}

type generatedThemeMeta struct {
	Type string `json:"type"`
}

type fileSnapshot struct {
	content []byte
	mode    fs.FileMode
	exists  bool
}

func (s *ThemeExportService) SaveThemeToEditorTarget(editorTarget string, themeName string, themeJSON string) (string, error) {
	if strings.TrimSpace(themeJSON) == "" {
		return "", errors.New("theme JSON cannot be empty")
	}
	if !json.Valid([]byte(themeJSON)) {
		return "", errors.New("theme JSON is invalid")
	}

	target := strings.ToLower(strings.TrimSpace(editorTarget))
	switch target {
	case "vscode", "cursor", "antigravity":
		return saveThemeToVSCodeFamily(target, themeName, themeJSON)
	case "zed":
		return saveThemeToZed(themeName, themeJSON)
	default:
		return "", fmt.Errorf("unsupported editor target: %s", editorTarget)
	}
}

func saveThemeToVSCodeFamily(target string, themeName string, themeJSON string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not resolve user home directory: %w", err)
	}
	return saveThemeToVSCodeFamilyAt(home, target, themeName, themeJSON)
}

func saveThemeToVSCodeAt(home string, themeName string, themeJSON string) (string, error) {
	return saveThemeToVSCodeFamilyAt(home, "vscode", themeName, themeJSON)
}

func saveThemeToVSCodeFamilyAt(home string, target string, themeName string, themeJSON string) (string, error) {
	resolvedThemeName := sanitizeThemeName(themeName)
	if resolvedThemeName == "" {
		resolvedThemeName = "generated-theme"
	}

	extensionDirectory, err := resolveVSCodeFamilyExtensionDirectory(home, target)
	if err != nil {
		return "", err
	}
	themesDirectory := filepath.Join(extensionDirectory, "themes")
	if err := os.MkdirAll(themesDirectory, 0o755); err != nil {
		return "", filesystemError("create VS Code theme directory", themesDirectory, err)
	}

	themeFileName := resolvedThemeName + ".json"
	themeFilePath := filepath.Join(themesDirectory, themeFileName)
	themeSnapshot, err := snapshotFile(themeFilePath)
	if err != nil {
		return "", filesystemError("read existing VS Code theme file", themeFilePath, err)
	}
	if err := writeFileSafely(themeFilePath, prettyJSON(themeJSON, "    "), 0o644); err != nil {
		return "", filesystemError("write VS Code theme file", themeFilePath, err)
	}

	themeLabel := strings.TrimSpace(themeName)
	if themeLabel == "" {
		themeLabel = "Generated Theme"
	}
	if err := updateVSCodePackageJSON(extensionDirectory, themeLabel, themeFileName, inferVSCodeUITheme(themeJSON)); err != nil {
		if rollbackErr := restoreFile(themeFilePath, themeSnapshot); rollbackErr != nil {
			return "", errors.Join(err, fmt.Errorf("failed to roll back VS Code theme file: %w", rollbackErr))
		}
		return "", err
	}

	return extensionDirectory, nil
}

func resolveVSCodeFamilyExtensionDirectory(home string, target string) (string, error) {
	var editorDirectory string
	switch strings.ToLower(strings.TrimSpace(target)) {
	case "vscode":
		editorDirectory = ".vscode"
	case "cursor":
		editorDirectory = ".cursor"
	case "antigravity":
		editorDirectory = ".antigravity-ide"
	default:
		return "", fmt.Errorf("unsupported VS Code family target: %s", target)
	}

	return filepath.Join(home, editorDirectory, "extensions", "themesmith-local"), nil
}

func saveThemeToZed(themeName string, themeJSON string) (string, error) {
	resolvedThemeName := sanitizeThemeName(themeName)
	if resolvedThemeName == "" {
		resolvedThemeName = "generated-theme"
	}

	targetDirectory, err := resolveZedThemeDirectory()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(targetDirectory, 0o755); err != nil {
		return "", filesystemError("create Zed theme directory", targetDirectory, err)
	}

	filePath := filepath.Join(targetDirectory, resolvedThemeName+".json")
	if err := writeFileSafely(filePath, prettyJSON(themeJSON, "    "), 0o644); err != nil {
		return "", filesystemError("write Zed theme file", filePath, err)
	}

	return filePath, nil
}

func resolveZedThemeDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not resolve user home directory: %w", err)
	}

	return resolveZedThemeDirectoryFor(runtime.GOOS, home, os.LookupEnv)
}

func resolveZedThemeDirectoryFor(goos string, home string, lookupEnv func(string) (string, bool)) (string, error) {
	switch goos {
	case "windows":
		if appData, ok := lookupEnv("APPDATA"); ok && strings.TrimSpace(appData) != "" {
			return filepath.Join(appData, "Zed", "themes"), nil
		}
		return filepath.Join(home, "AppData", "Roaming", "Zed", "themes"), nil
	case "darwin":
		return filepath.Join(home, ".config", "zed", "themes"), nil
	case "linux":
		if xdgConfigHome, ok := lookupEnv("XDG_CONFIG_HOME"); ok && strings.TrimSpace(xdgConfigHome) != "" {
			return filepath.Join(xdgConfigHome, "zed", "themes"), nil
		}
		return filepath.Join(home, ".config", "zed", "themes"), nil
	default:
		return "", fmt.Errorf("unsupported OS for Zed target: %s", goos)
	}
}

func updateVSCodePackageJSON(extensionDirectory string, themeLabel string, themeFileName string, uiTheme string) error {
	packagePath := filepath.Join(extensionDirectory, "package.json")

	packageData := defaultVSCodePackageData()
	if content, err := os.ReadFile(packagePath); err == nil {
		if unmarshalErr := json.Unmarshal(trimUTF8BOM(content), &packageData); unmarshalErr != nil {
			return fmt.Errorf("failed to parse existing VS Code package.json: %w", unmarshalErr)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to read VS Code package.json: %w", err)
	}

	if packageData.Contributes.Themes == nil {
		packageData.Contributes.Themes = []vscodeThemeEntry{}
	}

	filteredThemes := make([]vscodeThemeEntry, 0, len(packageData.Contributes.Themes)+1)
	replaced := false

	for _, entry := range packageData.Contributes.Themes {
		if strings.EqualFold(strings.TrimSpace(entry.Label), strings.TrimSpace(themeLabel)) ||
			strings.EqualFold(filepath.Clean(entry.Path), filepath.Clean("./themes/"+themeFileName)) {
			filteredThemes = append(filteredThemes, vscodeThemeEntry{
				Label:   themeLabel,
				UITheme: uiTheme,
				Path:    "./themes/" + themeFileName,
			})
			replaced = true
			continue
		}

		filteredThemes = append(filteredThemes, entry)
	}

	if !replaced {
		filteredThemes = append(filteredThemes, vscodeThemeEntry{
			Label:   themeLabel,
			UITheme: uiTheme,
			Path:    "./themes/" + themeFileName,
		})
	}

	packageData.Contributes.Themes = filteredThemes

	encoded, err := json.MarshalIndent(packageData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode VS Code package.json: %w", err)
	}

	encoded = append(encoded, '\n')
	if err := writeFileSafely(packagePath, encoded, 0o644); err != nil {
		return filesystemError("write VS Code extension manifest", packagePath, err)
	}

	return nil
}

func defaultVSCodePackageData() vscodePackageJSON {
	return vscodePackageJSON{
		Name:        "themesmith-local",
		DisplayName: "ThemeSmith",
		Description: "Custom themes generated by ThemeSmith",
		Version:     "0.0.1",
		Publisher:   "local",
		Engines: map[string]string{
			"vscode": "^1.70.0",
		},
		Contributes: vscodePackageJSONContributes{
			Themes: []vscodeThemeEntry{},
		},
	}
}

func trimUTF8BOM(content []byte) []byte {
	if len(content) >= 3 && content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
		return content[3:]
	}

	return content
}

func inferVSCodeUITheme(themeJSON string) string {
	meta := generatedThemeMeta{}
	if err := json.Unmarshal([]byte(themeJSON), &meta); err != nil {
		return "vs-dark"
	}

	if strings.EqualFold(strings.TrimSpace(meta.Type), "light") {
		return "vs"
	}

	return "vs-dark"
}

func prettyJSON(raw string, indent string) []byte {
	var decoded any
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return []byte(raw)
	}

	encoded, err := json.MarshalIndent(decoded, "", indent)
	if err != nil {
		return []byte(raw)
	}

	encoded = append(encoded, '\n')
	return encoded
}

func writeFileSafely(path string, content []byte, mode fs.FileMode) (returnErr error) {
	temporaryFile, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+"-*")
	if err != nil {
		return err
	}
	temporaryPath := temporaryFile.Name()
	temporaryFileOpen := true
	defer func() {
		if temporaryFileOpen {
			if closeErr := temporaryFile.Close(); returnErr == nil && closeErr != nil {
				returnErr = closeErr
			}
		}
		if removeErr := os.Remove(temporaryPath); returnErr == nil && removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			returnErr = removeErr
		}
	}()

	if err := temporaryFile.Chmod(mode); err != nil {
		return err
	}
	if _, err := temporaryFile.Write(content); err != nil {
		return err
	}
	if err := temporaryFile.Sync(); err != nil {
		return err
	}
	if err := temporaryFile.Close(); err != nil {
		return err
	}
	temporaryFileOpen = false

	// Chmod clears the read-only attribute on Windows for themes previously created by another tool.
	if _, err := os.Stat(path); err == nil {
		if err := os.Chmod(path, mode); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if err := os.Rename(temporaryPath, path); err != nil {
		return err
	}

	written, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !bytes.Equal(written, content) {
		return errors.New("saved file verification failed")
	}

	return nil
}

func snapshotFile(path string) (fileSnapshot, error) {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return fileSnapshot{}, nil
	}
	if err != nil {
		return fileSnapshot{}, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return fileSnapshot{}, err
	}

	return fileSnapshot{content: content, mode: info.Mode().Perm(), exists: true}, nil
}

func restoreFile(path string, snapshot fileSnapshot) error {
	if snapshot.exists {
		return writeFileSafely(path, snapshot.content, snapshot.mode)
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func filesystemError(action string, path string, err error) error {
	if errors.Is(err, fs.ErrPermission) {
		return fmt.Errorf("permission denied: cannot %s at %q: %w", action, path, err)
	}
	return fmt.Errorf("failed to %s at %q: %w", action, path, err)
}

func sanitizeThemeName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ""
	}

	b := strings.Builder{}
	b.Grow(len(trimmed))
	for _, r := range trimmed {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
			continue
		}

		switch r {
		case ' ', '-', '_', '.':
			b.WriteRune('-')
		}
	}

	sanitized := strings.Trim(b.String(), "-.")
	if sanitized == "" {
		return ""
	}

	sanitized = strings.ReplaceAll(sanitized, "--", "-")
	for strings.Contains(sanitized, "--") {
		sanitized = strings.ReplaceAll(sanitized, "--", "-")
	}

	if isWindowsReservedFileName(sanitized) {
		sanitized = "theme-" + sanitized
	}

	const maxRunes = 80
	runes := []rune(sanitized)
	if len(runes) > maxRunes {
		digest := sha256.Sum256([]byte(sanitized))
		suffix := fmt.Sprintf("-%x", digest[:5])
		sanitized = string(runes[:maxRunes-len(suffix)]) + suffix
	}

	return sanitized
}

func isWindowsReservedFileName(name string) bool {
	switch strings.ToUpper(name) {
	case "CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
		return true
	default:
		return false
	}
}
