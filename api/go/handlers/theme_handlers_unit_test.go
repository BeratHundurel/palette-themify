package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestNormalizeThemeSignature(t *testing.T) {
	t.Run("keeps short signatures", func(t *testing.T) {
		signature := "short-signature"
		normalized := normalizeThemeSignature(signature)
		if normalized != signature {
			t.Fatalf("expected signature to stay unchanged")
		}
	})

	t.Run("hashes long signatures", func(t *testing.T) {
		signature := strings.Repeat("sig-very-long-", 20)
		normalized := normalizeThemeSignature(signature)

		if len(normalized) != 64 {
			t.Fatalf("expected normalized signature length 64, got %d", len(normalized))
		}

		hash := sha256.Sum256([]byte(signature))
		expected := hex.EncodeToString(hash[:])
		if normalized != expected {
			t.Fatalf("expected normalized signature to be sha256 hex")
		}
	})
}

func TestParseThemePayload(t *testing.T) {
	t.Run("normalizes long signature", func(t *testing.T) {
		body := []byte(`{"name":"Theme","editorType":"vscode","signature":"` + strings.Repeat("sig-very-long-", 20) + `"}`)

		_, info, err := parseThemePayload(body)
		if err != nil {
			t.Fatalf("expected parse to succeed: %v", err)
		}

		if len(info.Signature) != 64 {
			t.Fatalf("expected normalized signature length 64, got %d", len(info.Signature))
		}
	})

	t.Run("rejects invalid json", func(t *testing.T) {
		_, _, err := parseThemePayload([]byte("not-json"))
		if err == nil || err.Error() != "Invalid theme payload" {
			t.Fatalf("expected invalid payload error, got %v", err)
		}
	})

	t.Run("requires fields", func(t *testing.T) {
		tests := []struct {
			name        string
			body        string
			expectedErr string
		}{
			{
				name:        "missing name",
				body:        `{"editorType":"vscode","signature":"sig"}`,
				expectedErr: "Theme name is required",
			},
			{
				name:        "missing editorType",
				body:        `{"name":"Theme","signature":"sig"}`,
				expectedErr: "Theme editor type is required",
			},
			{
				name:        "missing signature",
				body:        `{"name":"Theme","editorType":"vscode"}`,
				expectedErr: "Theme signature is required",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				_, _, err := parseThemePayload([]byte(tc.body))
				if err == nil || err.Error() != tc.expectedErr {
					t.Fatalf("expected %q, got %v", tc.expectedErr, err)
				}
			})
		}
	})
}
