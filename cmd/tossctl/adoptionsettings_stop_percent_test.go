package main

import (
	"context"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/console"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

func TestAdoptionSettingsRoundTripsStopWidth(t *testing.T) {
	cfgDir := t.TempDir()
	t.Setenv("TOSSOS_DATA_DIR", t.TempDir())
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(`{"schema_version":5}`), 0o600); err != nil {
		t.Fatal(err)
	}
	seam := newAdoptionSettingsSeam(&rootOptions{configDir: cfgDir})
	startVerify := func(
		context.Context,
		verifylive.BatchConfirmer,
		io.Writer,
		string,
		[]verifylive.StepID,
	) (verifylive.Summary, []verifylive.Entry, error) {
		return verifylive.Summary{}, nil, nil
	}
	app, err := console.New(console.Options{StartVerify: startVerify, Settings: seam})
	if err != nil {
		t.Fatalf("console.New: %v", err)
	}
	srv := httptest.NewServer(app.Handler())
	defer srv.Close()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Jar: jar}
	login, err := client.Get(srv.URL + "/?session=" + app.SessionToken())
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	login.Body.Close()

	// The adoption form is on the 상시 tab since change a055; /settings itself is
	// a redirect to the daily tab now.
	settings, err := client.Get(srv.URL + "/settings/standing")
	if err != nil {
		t.Fatalf("GET settings: %v", err)
	}
	body, err := io.ReadAll(settings.Body)
	settings.Body.Close()
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}
	match := regexp.MustCompile(`name="csrf" value="([^"]+)"`).FindSubmatch(body)
	if len(match) != 2 {
		t.Fatalf("settings page has no CSRF token")
	}
	response, err := client.PostForm(srv.URL+"/settings/save", url.Values{
		"csrf":                 {string(match[1])},
		"default_stop_percent": {"7.5"},
	})
	if err != nil {
		t.Fatalf("POST settings: %v", err)
	}
	response.Body.Close()

	block, verdict, err := seam.Load()
	if err != nil || verdict != "" || block.DefaultStopPct != 0.075 {
		t.Fatalf("round trip: block=%+v verdict=%q err=%v", block, verdict, err)
	}
}
