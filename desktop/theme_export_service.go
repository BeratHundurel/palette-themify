package main

import (
	"encoding/json"
	"errors"
	"fmt"
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

func (s *ThemeExportService) SaveThemeToEditorTarget(editorType string, themeName string, themeJSON string) (string, error) {
	if strings.TrimSpace(themeJSON) == "" {
		return "", errors.New("theme JSON cannot be empty")
	}

	switch strings.ToLower(strings.TrimSpace(editorType)) {
	case "vscode":
		return saveThemeToVSCode(themeName, themeJSON)
	case "zed":
		return saveThemeToZed(themeName, themeJSON)
	default:
		return "", fmt.Errorf("unsupported editor type: %s", editorType)
	}
}

func saveThemeToVSCode(themeName string, themeJSON string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not resolve user home directory: %w", err)
	}

	resolvedThemeName := sanitizeThemeName(themeName)
	if resolvedThemeName == "" {
		resolvedThemeName = "generated-theme"
	}

	extensionDirectory := filepath.Join(home, ".vscode", "extensions", "themesmith-local")
	themesDirectory := filepath.Join(extensionDirectory, "themes")
	if err := os.MkdirAll(themesDirectory, 0o755); err != nil {
		return "", fmt.Errorf("failed to create VS Code theme directory: %w", err)
	}

	themeFileName := resolvedThemeName + ".json"
	themeFilePath := filepath.Join(themesDirectory, themeFileName)
	if err := os.WriteFile(themeFilePath, prettyJSON(themeJSON, "    "), 0o644); err != nil {
		return "", fmt.Errorf("failed to write VS Code theme file: %w", err)
	}

	if err := updateVSCodePackageJSON(extensionDirectory, themeName, themeFileName, inferVSCodeUITheme(themeJSON)); err != nil {
		return "", err
	}

	return extensionDirectory, nil
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
		return "", fmt.Errorf("failed to create Zed theme directory: %w", err)
	}

	filePath := filepath.Join(targetDirectory, resolvedThemeName+".json")
	if err := os.WriteFile(filePath, prettyJSON(themeJSON, "    "), 0o644); err != nil {
		return "", fmt.Errorf("failed to write Zed theme file: %w", err)
	}

	return filePath, nil
}

func resolveZedThemeDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not resolve user home directory: %w", err)
	}

	switch runtime.GOOS {
	case "windows":
		if appData, ok := os.LookupEnv("APPDATA"); ok && strings.TrimSpace(appData) != "" {
			return filepath.Join(appData, "Zed", "themes"), nil
		}
		return filepath.Join(home, "AppData", "Roaming", "Zed", "themes"), nil
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Zed", "themes"), nil
	case "linux":
		if xdgConfigHome, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok && strings.TrimSpace(xdgConfigHome) != "" {
			return filepath.Join(xdgConfigHome, "zed", "themes"), nil
		}
		return filepath.Join(home, ".config", "zed", "themes"), nil
	default:
		return "", fmt.Errorf("unsupported OS for Zed target: %s", runtime.GOOS)
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
		if strings.EqualFold(strings.TrimSpace(entry.Label), strings.TrimSpace(themeLabel)) {
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
	if err := os.WriteFile(packagePath, encoded, 0o644); err != nil {
		return fmt.Errorf("failed to write VS Code package.json: %w", err)
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

	return sanitized
}
