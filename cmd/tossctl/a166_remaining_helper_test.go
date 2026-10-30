package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/session"
)

func TestA166RemainingOfficialFactory(t *testing.T) {
	ctx := context.Background()
	client, err := newRootOfficialClient(ctx, &rootOptions{}, official.Credentials{}, "")
	if err != nil || client == nil {
		t.Fatalf("OFF: client=%v err=%v", client, err)
	}
	bad := newRootBudgetOwner(func(key string) (string, bool) {
		if key == "TOSSOS_ACCOUNT_BUDGET_MODE" {
			return "required", true
		}
		return "", false
	}, nil)
	client, err = newRootOfficialClient(ctx, &rootOptions{budgetOwner: bad}, official.Credentials{}, "")
	if err == nil || client != nil {
		t.Fatalf("required missing manifest: client=%v err=%v", client, err)
	}
}

func TestA166MCPRequiredFailureDoesNotFallback(t *testing.T) {
	dir := isolate(t)
	cfg := config.DefaultFile()
	cfg.OpenAPI = config.OpenAPI{Enabled: true, Prefer: "auto"}
	writeConfigFile(t, dir, cfg)
	writeCredentials(t, dir)
	owner := newRootBudgetOwner(func(key string) (string, bool) {
		if key == "TOSSOS_ACCOUNT_BUDGET_MODE" {
			return "required", true
		}
		return "", false
	}, nil)
	oldTransport := http.DefaultTransport
	http.DefaultTransport = a166NoNetwork{}
	t.Cleanup(func() { http.DefaultTransport = oldTransport })
	cmd := newMCPCmd(&rootOptions{configDir: dir, budgetOwner: owner})
	cmd.SetIn(strings.NewReader(""))
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	if err := cmd.Execute(); !errors.Is(err, accountbudget.ErrUnavailable) {
		t.Fatalf("expected budget error, got %v", err)
	}
}

type a166NoNetwork struct{}

func (a166NoNetwork) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("test blocks network")
}

func TestA166MCPWTSDoesNotOpenBudget(t *testing.T) {
	dir := isolate(t)
	cfg := config.DefaultFile()
	cfg.OpenAPI = config.OpenAPI{Enabled: true, Prefer: "auto"}
	writeConfigFile(t, dir, cfg)
	writeCredentials(t, dir)
	if err := session.WriteFile(filepath.Join(dir, "session.json"), &session.Session{Cookies: map[string]string{"SESSION": "fixture"}}); err != nil {
		t.Fatal(err)
	}
	owner := newRootBudgetOwner(func(key string) (string, bool) {
		if key == "TOSSOS_ACCOUNT_BUDGET_MODE" {
			return "required", true
		}
		return "", false
	}, nil)
	oldTransport := http.DefaultTransport
	http.DefaultTransport = a166NoNetwork{}
	t.Cleanup(func() { http.DefaultTransport = oldTransport })
	cmd := newMCPCmd(&rootOptions{configDir: dir, backend: "wts", budgetOwner: owner})
	cmd.SetIn(strings.NewReader(""))
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if owner.factory != nil {
		t.Fatal("WTS unexpectedly opened budget")
	}
}
