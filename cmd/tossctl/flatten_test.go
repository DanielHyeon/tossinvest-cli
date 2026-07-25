package main

// flatten_test.go covers the CLI surface of `tossctl flatten-all` (task 4.5).
//
// The saga's behaviour is tested in internal/flatten. What is tested here is what
// only the command can get wrong: the flags it offers, the flags it must never
// offer, and its refusal to run without official credentials.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/spf13/cobra"
)

func findCommand(root *cobra.Command, name string) *cobra.Command {
	for _, c := range root.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

// TestFlattenOffersNoAutomationFlag is the spec requirement, as a test:
// "확인 문자열은 … TTY 직접 입력만 허용하며 자동화 플래그는 금지된다".
//
// The list is of names an author might reach for while making the command
// scriptable. Adding any of them would let an agent or a cron job sell an entire
// account, which is exactly what the confirmation exists to prevent.
func TestFlattenOffersNoAutomationFlag(t *testing.T) {
	cmd := findCommand(newRootCmd(), "flatten-all")
	if cmd == nil {
		t.Fatal("flatten-all is not registered on the root command")
	}

	banned := []string{
		"yes", "y", "force", "f", "confirm", "assume-yes", "no-confirm",
		"non-interactive", "batch", "auto-approve", "skip-confirmation",
	}
	for _, name := range banned {
		if flag := cmd.Flags().Lookup(name); flag != nil {
			t.Errorf("flatten-all offers --%s; the liquidation confirmation must be typed by a person", name)
		}
		if len(name) == 1 {
			if flag := cmd.Flags().ShorthandLookup(name); flag != nil {
				t.Errorf("flatten-all offers -%s; the liquidation confirmation must be typed by a person", name)
			}
		}
	}
}

// TestFlattenOffersDryRun: the spec requires a mode that lists the plan and
// mutates nothing.
func TestFlattenOffersDryRun(t *testing.T) {
	cmd := findCommand(newRootCmd(), "flatten-all")
	if cmd == nil {
		t.Fatal("flatten-all is not registered")
	}
	flag := cmd.Flags().Lookup("dry-run")
	if flag == nil {
		t.Fatal("flatten-all has no --dry-run")
	}
	if flag.DefValue != "false" {
		t.Errorf("--dry-run defaults to %q; the safe default is not the issue here, but it must be opt-in",
			flag.DefValue)
	}
}

// TestFlattenIsAnnotatedAsAnOfficialMutation. The annotations drive the help
// convention checks and tell a reader which backend the command reaches.
func TestFlattenIsAnnotatedAsAnOfficialMutation(t *testing.T) {
	cmd := findCommand(newRootCmd(), "flatten-all")
	if cmd == nil {
		t.Fatal("flatten-all is not registered")
	}
	if got := cmd.Annotations["source"]; got != "official" {
		t.Errorf("source = %q, want official — the engine profile has no web-session path", got)
	}
	if got := cmd.Annotations["mutating"]; got != "true" {
		t.Errorf("mutating = %q, want true", got)
	}
}

// TestFlattenRefusesWithoutOfficialCredentials.
//
// A flatten that ran through a scraped web session would leave no durable record
// of what it did, so the engine profile refuses to start without official
// credentials and this command inherits that.
func TestFlattenRefusesWithoutOfficialCredentials(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(root, "cache"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(root, "data"))
	t.Setenv("TOSSOS_DATA_DIR", filepath.Join(root, "tossos"))
	t.Setenv("TOSSCTL_OPENAPI_KEY", "")
	t.Setenv("TOSSCTL_OPENAPI_SECRET", "")

	configDir := filepath.Join(root, "cfg")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	cmd := newRootCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"--config-dir", configDir, "flatten-all", "--dry-run"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("flatten-all must refuse to run without official credentials")
	}
	if !strings.Contains(err.Error(), "credentials") {
		t.Errorf("err = %v, want it to name the missing credentials", err)
	}
	if err.Error() != engine.ErrOfficialCredentialsRequired.Error() {
		t.Logf("note: the error is %v", err)
	}
}
