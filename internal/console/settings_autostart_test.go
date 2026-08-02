package console

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
)

type fakeAutostart struct {
	mu      sync.Mutex
	enabled bool
	loadErr error
	saveErr error
	saved   []bool
}

func (f *fakeAutostart) Load() (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.enabled, f.loadErr
}

func (f *fakeAutostart) Save(on bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.saveErr != nil {
		return f.saveErr
	}
	f.enabled = on
	f.saved = append(f.saved, on)
	return nil
}

func TestAutostartScreenRendersStateAndMeaning(t *testing.T) {
	a := &fakeAutostart{enabled: true}
	h := newDashboardHarness(t, func(o *Options) { o.EngineBoot = a })
	h.authenticate(t)

	page := body(t, h.get(t, pathSettingsStanding))
	for _, want := range []string{
		`action="/settings/autostart"`,
		`name="enabled" checked`,
		"엔진 자동 시작",
		"automation gate",
		"Guardian",
		"OFF로 저장해도 현재 엔진은 정지하지 않는다",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("autostart section does not say %q:\n%s", want, page)
		}
	}
}

func TestAutostartOnSavesThenStartsExactlyOnce(t *testing.T) {
	a := &fakeAutostart{}
	starts := 0
	stops := 0
	h := newDashboardHarness(t, func(o *Options) {
		o.EngineBoot = a
		o.StartEngine = func() (string, error) {
			starts++
			return "engine child accepted", nil
		}
		o.StopEngine = func() (string, error) {
			stops++
			return "", nil
		}
	})
	h.authenticate(t)

	page := body(t, h.post(t, "/settings/autostart", url.Values{
		"csrf": {h.csrf}, "enabled": {"on"},
	}))
	if len(a.saved) != 1 || !a.saved[0] {
		t.Fatalf("saved = %v, want [true]", a.saved)
	}
	if starts != 1 {
		t.Fatalf("starts = %d, want 1", starts)
	}
	if stops != 0 {
		t.Fatalf("stops = %d, want 0", stops)
	}
	if !strings.Contains(page, "engine child accepted") {
		t.Fatalf("start result is absent from settings page:\n%s", page)
	}
}

func TestAutostartOnPreservesTheEnginesOwnRefusal(t *testing.T) {
	a := &fakeAutostart{}
	starts := 0
	h := newDashboardHarness(t, func(o *Options) {
		o.EngineBoot = a
		o.StartEngine = func() (string, error) {
			starts++
			return "interlock clauses: capability attestation expired", errors.New("startup refused")
		}
	})
	h.authenticate(t)

	page := body(t, h.post(t, "/settings/autostart", url.Values{
		"csrf": {h.csrf}, "enabled": {"on"},
	}))
	if len(a.saved) != 1 || !a.saved[0] {
		t.Fatalf("the persisted approval was not recorded: %v", a.saved)
	}
	if starts != 1 {
		t.Fatalf("starts = %d, want exactly one", starts)
	}
	for _, want := range []string{"startup refused", "capability attestation expired"} {
		if !strings.Contains(page, want) {
			t.Errorf("engine refusal does not include %q:\n%s", want, page)
		}
	}
}

func TestAutostartOffDoesNotStopOrStartTheCurrentEngine(t *testing.T) {
	a := &fakeAutostart{enabled: true}
	starts := 0
	stops := 0
	h := newDashboardHarness(t, func(o *Options) {
		o.EngineBoot = a
		o.StartEngine = func() (string, error) { starts++; return "", nil }
		o.StopEngine = func() (string, error) { stops++; return "", nil }
	})
	h.authenticate(t)

	page := body(t, h.post(t, "/settings/autostart", url.Values{"csrf": {h.csrf}}))
	if len(a.saved) != 1 || a.saved[0] {
		t.Fatalf("saved = %v, want [false]", a.saved)
	}
	if starts != 0 || stops != 0 {
		t.Fatalf("starts=%d stops=%d, OFF is not a process action", starts, stops)
	}
	if !strings.Contains(page, "현재 엔진은 [엔진 정지]") {
		t.Fatalf("OFF notice does not explain current process state:\n%s", page)
	}
}

func TestAutostartSaveFailureNeverStarts(t *testing.T) {
	a := &fakeAutostart{saveErr: errors.New("disk full")}
	starts := 0
	h := newDashboardHarness(t, func(o *Options) {
		o.EngineBoot = a
		o.StartEngine = func() (string, error) { starts++; return "", nil }
	})
	h.authenticate(t)

	page := body(t, h.post(t, "/settings/autostart", url.Values{
		"csrf": {h.csrf}, "enabled": {"on"},
	}))
	if starts != 0 {
		t.Fatalf("engine started after failed save: %d", starts)
	}
	if !strings.Contains(page, "disk full") {
		t.Fatalf("save error is absent:\n%s", page)
	}
}

func TestAutostartRequiresCSRF(t *testing.T) {
	a := &fakeAutostart{}
	starts := 0
	h := newDashboardHarness(t, func(o *Options) {
		o.EngineBoot = a
		o.StartEngine = func() (string, error) { starts++; return "", nil }
	})
	h.authenticate(t)

	resp := h.post(t, "/settings/autostart", url.Values{
		"csrf": {"wrong"}, "enabled": {"on"},
	})
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", resp.StatusCode)
	}
	if len(a.saved) != 0 || starts != 0 {
		t.Fatalf("CSRF refusal changed state: saved=%v starts=%d", a.saved, starts)
	}
}

func TestAutostartUnwiredAndLoadErrorAreExplicit(t *testing.T) {
	t.Run("unwired", func(t *testing.T) {
		h := newDashboardHarness(t)
		h.authenticate(t)
		page := body(t, h.get(t, pathSettingsStanding))
		if !strings.Contains(page, "엔진 자동 시작 저장이 배선되지 않았다") {
			t.Fatalf("unwired section disappeared:\n%s", page)
		}
		if strings.Contains(page, `action="/settings/autostart"`) {
			t.Fatal("autostart form exists with no seam")
		}
	})
	t.Run("load error", func(t *testing.T) {
		h := newDashboardHarness(t, func(o *Options) {
			o.EngineBoot = &fakeAutostart{loadErr: errors.New("bad config")}
		})
		h.authenticate(t)
		page := body(t, h.get(t, pathSettingsStanding))
		if !strings.Contains(page, "bad config") {
			t.Fatalf("load error is absent:\n%s", page)
		}
	})
}

func TestInitialEngineNoteIsRendered(t *testing.T) {
	h := newDashboardHarness(t, func(o *Options) {
		o.EngineBootNote = "자동 시작 거부: gate is off"
	})
	h.authenticate(t)
	page := body(t, h.get(t, "/"))
	if !strings.Contains(page, "자동 시작 거부: gate is off") {
		t.Fatalf("initial engine note is absent:\n%s", page)
	}
}
