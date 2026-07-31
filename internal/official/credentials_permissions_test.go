package official_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

func TestSaveCredentialsTightensExistingFileTo0600(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "c.json")
	if err := os.WriteFile(file, []byte(`{"apiKey":"old","secretKey":"old"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := official.SaveCredentials(file, official.Credentials{
		APIKey: "replacement-key", SecretKey: "replacement-secret",
	}); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Lstat(file)
	if err != nil {
		t.Fatal(err)
	}
	if !fi.Mode().IsRegular() {
		t.Fatalf("credential target mode=%v, want regular file", fi.Mode())
	}
	if fi.Mode().Perm() != 0o600 {
		t.Fatalf("replacement kept permissive mode %v, want 0600", fi.Mode().Perm())
	}
}
