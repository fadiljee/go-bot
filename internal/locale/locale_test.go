package locale

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocale(t *testing.T) {
	// Create temporary translation file
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "test_locale.json")
	content := `{
		"greeting": "Halo {name}, selamat datang di {place}!",
		"simple": "Halo Dunia",
		"formatted": "Angka: %d, String: %s"
	}`
	if err := os.WriteFile(jsonPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	if err := Init(jsonPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Test 1: Simple key
	if res := T("simple"); res != "Halo Dunia" {
		t.Errorf("expected 'Halo Dunia', got '%s'", res)
	}

	// Test 2: Named placeholder
	if res := T("greeting", "name", "Budi", "place", "Bangka Barat"); res != "Halo Budi, selamat datang di Bangka Barat!" {
		t.Errorf("named placeholder mismatch: got '%s'", res)
	}

	// Test 3: Sprintf formatting
	if res := T("formatted", 42, "BPS"); res != "Angka: 42, String: BPS" {
		t.Errorf("sprintf formatting mismatch: got '%s'", res)
	}

	// Test 4: Missing key fallback
	if res := T("non_existent_key"); res != "non_existent_key" {
		t.Errorf("expected key fallback, got '%s'", res)
	}
}
