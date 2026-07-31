package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

func TestConsoleOpenAPISetupUsesIsolatedTokenThenSavesInvalidatesAndAudits(t *testing.T) {
	dir := t.TempDir()
	root := &rootOptions{configDir: dir}
	normalToken := filepath.Join(dir, "openapi-token.json")
	if err := os.WriteFile(normalToken, []byte(`{"access_token":"old"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	var order []string
	var validationToken string
	deps := defaultConsoleOpenAPIDeps()
	deps.validate = func(_ context.Context, creds official.Credentials, tokenFile string) (probeResult, error) {
		order = append(order, "validate")
		validationToken = tokenFile
		if tokenFile == normalToken {
			t.Fatal("replacement credentials used the normal token cache")
		}
		if err := os.WriteFile(tokenFile, []byte(`{"access_token":"temporary"}`), 0o600); err != nil {
			t.Fatal(err)
		}
		return probeResult{OK: true, Message: "ok"}, nil
	}
	deps.save = func(path, key, secret string) error {
		order = append(order, "save")
		return saveOpenAPICredentials(path, key, secret)
	}
	deps.remove = func(path string) error {
		order = append(order, "invalidate")
		return os.Remove(path)
	}
	deps.audit = func() error {
		order = append(order, "audit")
		return nil
	}

	seam, err := newConsoleOpenAPISeam(root, deps)
	if err != nil {
		t.Fatal(err)
	}
	got := seam.Save(context.Background(), "new-key", "new-secret")
	if got.State != consoleOpenAPIReady {
		t.Fatalf("state=%q message=%q", got.State, got.Message)
	}
	if strings.Join(order, ",") != "validate,save,invalidate,audit" {
		t.Fatalf("order=%q", strings.Join(order, ","))
	}
	if validationToken == "" || validationToken == normalToken {
		t.Fatalf("isolated token path=%q", validationToken)
	}
	if _, err := os.Stat(validationToken); !os.IsNotExist(err) {
		t.Fatalf("temporary token cache survived: %v", err)
	}
	if _, err := os.Stat(normalToken); !os.IsNotExist(err) {
		t.Fatalf("normal token cache survived: %v", err)
	}
	creds, err := official.LoadCredentials(func(string) string { return "" },
		filepath.Join(dir, "openapi-credentials.json"))
	if err != nil || creds == nil || creds.APIKey != "new-key" || creds.SecretKey != "new-secret" {
		t.Fatalf("saved credentials did not round-trip: %v", err)
	}
}

func TestConsoleOpenAPIRejectsEnvironmentCredentialWebReplacement(t *testing.T) {
	dir := t.TempDir()
	deps := defaultConsoleOpenAPIDeps()
	deps.getenv = func(name string) string {
		switch name {
		case "TOSSCTL_OPENAPI_KEY":
			return "environment-key"
		case "TOSSCTL_OPENAPI_SECRET":
			return "environment-secret"
		default:
			return ""
		}
	}
	deps.validate = func(context.Context, official.Credentials, string) (probeResult, error) {
		return probeResult{ErrorKind: "auth", Message: "raw auth failure"}, nil
	}
	deps.save = func(string, string, string) error {
		t.Fatal("environment-managed credentials were replaced with a file")
		return nil
	}
	seam, err := newConsoleOpenAPISeam(&rootOptions{configDir: dir}, deps)
	if err != nil {
		t.Fatal(err)
	}

	check := seam.Check(context.Background())
	if check.State != consoleOpenAPIEnvironmentRejected {
		t.Fatalf("check=%+v", check)
	}
	if strings.Contains(check.Message, "environment-key") || strings.Contains(check.Message, "environment-secret") {
		t.Fatal("environment credential leaked")
	}
	for _, name := range []string{"TOSSCTL_OPENAPI_KEY", "TOSSCTL_OPENAPI_SECRET"} {
		if !strings.Contains(check.Message, name) {
			t.Fatalf("guidance missing %s: %q", name, check.Message)
		}
	}

	save := seam.Save(context.Background(), "web-key", "web-secret")
	if save.State != consoleOpenAPIEnvironmentRejected {
		t.Fatalf("save=%+v", save)
	}
}

func TestConsoleOpenAPIWhitespaceEnvironmentUsesTheSameSourceAsTheChild(t *testing.T) {
	dir := t.TempDir()
	deps := defaultConsoleOpenAPIDeps()
	deps.getenv = func(name string) string {
		switch name {
		case "TOSSCTL_OPENAPI_KEY", "TOSSCTL_OPENAPI_SECRET":
			return " "
		default:
			return ""
		}
	}
	deps.validate = func(_ context.Context, creds official.Credentials, _ string) (probeResult, error) {
		if creds.APIKey != " " || creds.SecretKey != " " {
			t.Fatalf("effective environment credentials changed: %+v", creds)
		}
		return probeResult{ErrorKind: "auth"}, nil
	}
	deps.save = func(string, string, string) error {
		t.Fatal("environment-managed credentials were replaced with a file")
		return nil
	}
	seam, err := newConsoleOpenAPISeam(&rootOptions{configDir: dir}, deps)
	if err != nil {
		t.Fatal(err)
	}

	if got := seam.Check(context.Background()); got.State != consoleOpenAPIEnvironmentRejected {
		t.Fatalf("check=%+v", got)
	}
	if got := seam.Save(context.Background(), "web-key", "web-secret"); got.State != consoleOpenAPIEnvironmentRejected {
		t.Fatalf("save=%+v", got)
	}
}

func TestConsoleOpenAPIEnvironmentGenerationLeavesFileMarkerDormant(t *testing.T) {
	dir := t.TempDir()
	pending := filepath.Join(dir, openAPIOnboardingPendingName)
	if err := writeOpenAPIOnboardingPending(pending); err != nil {
		t.Fatal(err)
	}
	deps := defaultConsoleOpenAPIDeps()
	deps.getenv = func(name string) string {
		switch name {
		case "TOSSCTL_OPENAPI_KEY":
			return "environment-key"
		case "TOSSCTL_OPENAPI_SECRET":
			return "environment-secret"
		default:
			return ""
		}
	}
	deps.validate = func(context.Context, official.Credentials, string) (probeResult, error) {
		return probeResult{OK: true}, nil
	}
	seam, err := newConsoleOpenAPISeam(&rootOptions{configDir: dir}, deps)
	if err != nil {
		t.Fatal(err)
	}
	if got := seam.Check(context.Background()); got.State != consoleOpenAPIReady {
		t.Fatalf("environment generation=%+v", got)
	}
	if _, err := os.Stat(pending); err != nil {
		t.Fatalf("environment generation changed dormant marker: %v", err)
	}

	deps.getenv = func(string) string { return "" }
	seam, err = newConsoleOpenAPISeam(&rootOptions{configDir: dir}, deps)
	if err != nil {
		t.Fatal(err)
	}
	if got := seam.Check(context.Background()); got.State != consoleOpenAPISavedNotStarted {
		t.Fatalf("file mode bypassed dormant marker: %+v", got)
	}
}

func TestConsoleOpenAPIEnvironmentPreflightUsesFreshTokenAndInvalidatesNormalCache(t *testing.T) {
	dir := t.TempDir()
	normalToken := filepath.Join(dir, "openapi-token.json")
	if err := os.WriteFile(normalToken, []byte(`{"access_token":"old"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	deps := defaultConsoleOpenAPIDeps()
	deps.getenv = func(name string) string {
		switch name {
		case "TOSSCTL_OPENAPI_KEY":
			return "replacement-key"
		case "TOSSCTL_OPENAPI_SECRET":
			return "replacement-secret"
		default:
			return ""
		}
	}
	var validationToken string
	deps.validate = func(_ context.Context, _ official.Credentials, tokenFile string) (probeResult, error) {
		validationToken = tokenFile
		if tokenFile == normalToken {
			t.Fatal("environment replacement was validated through the old normal token cache")
		}
		if err := os.WriteFile(tokenFile, []byte(`{"access_token":"fresh"}`), 0o600); err != nil {
			t.Fatal(err)
		}
		return probeResult{OK: true}, nil
	}
	seam, err := newConsoleOpenAPISeam(&rootOptions{configDir: dir}, deps)
	if err != nil {
		t.Fatal(err)
	}

	if got := seam.Check(context.Background()); got.State != consoleOpenAPIReady {
		t.Fatalf("check=%+v", got)
	}
	if validationToken == "" || validationToken == normalToken {
		t.Fatalf("validation token=%q", validationToken)
	}
	if _, err := os.Stat(validationToken); !os.IsNotExist(err) {
		t.Fatalf("isolated token survived environment preflight: %v", err)
	}
	if _, err := os.Stat(normalToken); !os.IsNotExist(err) {
		t.Fatalf("old normal token survived environment replacement: %v", err)
	}
}

func TestConsoleOpenAPIEnvironmentPreflightFailuresRemainStopped(t *testing.T) {
	for _, tc := range []struct {
		name string
		fail func(*consoleOpenAPIDeps)
	}{
		{
			name: "isolated token allocation",
			fail: func(deps *consoleOpenAPIDeps) {
				deps.temp = func(string) (string, func() error, error) {
					return "", nil, errors.New("temp unavailable")
				}
				deps.validate = func(context.Context, official.Credentials, string) (probeResult, error) {
					t.Fatal("validation ran without an isolated token path")
					return probeResult{}, nil
				}
			},
		},
		{
			name: "isolated token cleanup",
			fail: func(deps *consoleOpenAPIDeps) {
				deps.temp = func(string) (string, func() error, error) {
					return filepath.Join(t.TempDir(), "isolated-token"),
						func() error { return errors.New("cleanup failed") }, nil
				}
			},
		},
		{
			name: "normal token invalidation",
			fail: func(deps *consoleOpenAPIDeps) {
				deps.remove = func(string) error { return errors.New("remove failed") }
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			deps := defaultConsoleOpenAPIDeps()
			deps.getenv = func(name string) string {
				switch name {
				case "TOSSCTL_OPENAPI_KEY":
					return "environment-key"
				case "TOSSCTL_OPENAPI_SECRET":
					return "environment-secret"
				default:
					return ""
				}
			}
			deps.validate = func(context.Context, official.Credentials, string) (probeResult, error) {
				return probeResult{OK: true}, nil
			}
			tc.fail(&deps)
			seam, err := newConsoleOpenAPISeam(&rootOptions{configDir: t.TempDir()}, deps)
			if err != nil {
				t.Fatal(err)
			}
			if got := seam.Check(context.Background()); got.State != consoleOpenAPIUnavailable {
				t.Fatalf("check=%+v", got)
			}
		})
	}
}

func TestWriteOpenAPIOnboardingPendingDoesNotTouchExistingMarker(t *testing.T) {
	path := filepath.Join(t.TempDir(), openAPIOnboardingPendingName)
	if err := writeOpenAPIOnboardingPending(path); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Millisecond)

	if err := writeOpenAPIOnboardingPending(path); err != nil {
		t.Fatal(err)
	}
	after, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !after.ModTime().Equal(before.ModTime()) {
		t.Fatalf("existing fail-closed marker was rewritten: before=%v after=%v",
			before.ModTime(), after.ModTime())
	}
}

func TestConsoleOpenAPIPendingGenerationBlocksPreflightUntilSetupCompletes(t *testing.T) {
	for _, tc := range []struct {
		name string
		fail func(*consoleOpenAPIDeps)
	}{
		{
			name: "token invalidation",
			fail: func(deps *consoleOpenAPIDeps) {
				deps.remove = func(string) error { return errors.New("disk busy") }
			},
		},
		{
			name: "audit",
			fail: func(deps *consoleOpenAPIDeps) {
				deps.audit = func() error { return errors.New("audit unavailable") }
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			deps := defaultConsoleOpenAPIDeps()
			deps.getenv = func(string) string { return "" }
			deps.validate = func(context.Context, official.Credentials, string) (probeResult, error) {
				return probeResult{OK: true}, nil
			}
			deps.audit = func() error { return nil }
			tc.fail(&deps)
			seam, err := newConsoleOpenAPISeam(&rootOptions{configDir: dir}, deps)
			if err != nil {
				t.Fatal(err)
			}

			if got := seam.Save(context.Background(), "key", "secret"); got.State != consoleOpenAPISavedNotStarted {
				t.Fatalf("failed setup=%+v", got)
			}
			info, err := os.Stat(seam.pendingFile)
			if err != nil {
				t.Fatalf("pending marker missing: %v", err)
			}
			if got := info.Mode().Perm(); got != 0o600 {
				t.Fatalf("pending marker mode=%#o, want 0600", got)
			}
			if got := seam.Check(context.Background()); got.State != consoleOpenAPISavedNotStarted {
				t.Fatalf("ordinary preflight bypassed pending generation: %+v", got)
			}

			deps.remove = func(path string) error {
				err := os.Remove(path)
				if errors.Is(err, os.ErrNotExist) {
					return nil
				}
				return err
			}
			deps.audit = func() error { return nil }
			seam, err = newConsoleOpenAPISeam(&rootOptions{configDir: dir}, deps)
			if err != nil {
				t.Fatal(err)
			}
			if got := seam.Save(context.Background(), "key", "secret"); got.State != consoleOpenAPIReady {
				t.Fatalf("retry setup=%+v", got)
			}
			if got := seam.Check(context.Background()); got.State != consoleOpenAPIReady {
				t.Fatalf("completed setup did not release preflight: %+v", got)
			}
		})
	}
}

func TestConsoleOpenAPIMarkerFailuresRemainFailClosed(t *testing.T) {
	readyDeps := func() consoleOpenAPIDeps {
		deps := defaultConsoleOpenAPIDeps()
		deps.getenv = func(string) string { return "" }
		deps.validate = func(context.Context, official.Credentials, string) (probeResult, error) {
			return probeResult{OK: true}, nil
		}
		deps.save = func(string, string, string) error { return nil }
		deps.remove = func(string) error { return nil }
		deps.audit = func() error { return nil }
		return deps
	}

	t.Run("marker read", func(t *testing.T) {
		deps := readyDeps()
		deps.pending = func(string) (bool, error) {
			return false, errors.New("read failed")
		}
		seam, err := newConsoleOpenAPISeam(&rootOptions{configDir: t.TempDir()}, deps)
		if err != nil {
			t.Fatal(err)
		}
		if got := seam.Check(context.Background()); got.State != consoleOpenAPIUnavailable {
			t.Fatalf("check=%+v", got)
		}
	})

	t.Run("marker create", func(t *testing.T) {
		deps := readyDeps()
		deps.mark = func(string) error { return errors.New("write failed") }
		deps.save = func(string, string, string) error {
			t.Fatal("credential save ran without a pending marker")
			return nil
		}
		seam, err := newConsoleOpenAPISeam(&rootOptions{configDir: t.TempDir()}, deps)
		if err != nil {
			t.Fatal(err)
		}
		if got := seam.Save(context.Background(), "key", "secret"); got.State != consoleOpenAPIUnavailable {
			t.Fatalf("save=%+v", got)
		}
	})

	t.Run("save error retains marker", func(t *testing.T) {
		deps := readyDeps()
		deps.mark = func(string) error { return nil }
		deps.save = func(string, string, string) error { return errors.New("save failed") }
		deps.clear = func(string) error {
			t.Fatal("an attempted credential save error cleared its fail-closed marker")
			return nil
		}
		seam, err := newConsoleOpenAPISeam(&rootOptions{configDir: t.TempDir()}, deps)
		if err != nil {
			t.Fatal(err)
		}
		if got := seam.Save(context.Background(), "key", "secret"); got.State != consoleOpenAPISavedNotStarted {
			t.Fatalf("save=%+v", got)
		}
	})

	t.Run("final marker clear", func(t *testing.T) {
		deps := readyDeps()
		deps.mark = func(string) error { return nil }
		deps.clear = func(string) error { return errors.New("clear failed") }
		seam, err := newConsoleOpenAPISeam(&rootOptions{configDir: t.TempDir()}, deps)
		if err != nil {
			t.Fatal(err)
		}
		if got := seam.Save(context.Background(), "key", "secret"); got.State != consoleOpenAPISavedNotStarted {
			t.Fatalf("save=%+v", got)
		}
	})
}

func TestConsoleOpenAPIPreflightClassifiesMissingReadyAndTransient(t *testing.T) {
	dir := t.TempDir()
	deps := defaultConsoleOpenAPIDeps()
	deps.getenv = func(string) string { return "" }
	seam, err := newConsoleOpenAPISeam(&rootOptions{configDir: dir}, deps)
	if err != nil {
		t.Fatal(err)
	}
	if got := seam.Check(context.Background()); got.State != consoleOpenAPIMissing {
		t.Fatalf("missing check=%+v", got)
	}

	if err := saveOpenAPICredentials(filepath.Join(dir, "openapi-credentials.json"), "k", "s"); err != nil {
		t.Fatal(err)
	}
	deps.validate = func(context.Context, official.Credentials, string) (probeResult, error) {
		return probeResult{OK: true, Message: "ok"}, nil
	}
	seam, _ = newConsoleOpenAPISeam(&rootOptions{configDir: dir}, deps)
	if got := seam.Check(context.Background()); got.State != consoleOpenAPIReady {
		t.Fatalf("ready check=%+v", got)
	}

	deps.validate = func(context.Context, official.Credentials, string) (probeResult, error) {
		return probeResult{ErrorKind: "rate_limited", Message: "raw broker text"}, nil
	}
	seam, _ = newConsoleOpenAPISeam(&rootOptions{configDir: dir}, deps)
	got := seam.Check(context.Background())
	if got.State != consoleOpenAPIUnavailable || strings.Contains(got.Message, "raw broker text") {
		t.Fatalf("transient check=%+v", got)
	}
}
