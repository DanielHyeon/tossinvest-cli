package console

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/html"
)

type stubMarketSchedule struct {
	reading MarketScheduleReading
	err     error
	calls   int
}

func (s *stubMarketSchedule) Read(context.Context) (MarketScheduleReading, error) {
	s.calls++
	return s.reading, s.err
}

func TestMarketScheduleShowsDesiredEffectiveAndCalendarProvenance(t *testing.T) {
	digest := "sha256:" + strings.Repeat("a", 64)
	reader := &stubMarketSchedule{reading: MarketScheduleReading{
		SchedulerDesired: false, AutoStartDesired: false,
		SchedulerEffective: "DISABLED", AutoStartEffective: false,
		Market: "none", Session: "regular", ApplyTiming: "다음 엔진 기동",
		CalendarSource: "official-openapi", CalendarVersion: digest,
		CalendarFetchedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		DecisionReason:    "NOT_ACTIVATED", NextTransition: time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC),
	}}
	h := newHarness(t, func(o *Options) { o.MarketSchedule = reader })
	h.authenticate(t)
	resp := h.get(t, "/strategy-runtime/market-schedule")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET = %d", resp.StatusCode)
	}
	page := body(t, resp)
	for _, want := range []string{
		"strategy-runtime", "시장·일정", "기본값", "Desired", "Effective",
		"DISABLED", "자동 시작", "선택 시장 없음", "정규장", "다음 엔진 기동",
		"official-openapi", digest, "NOT_ACTIVATED", "다음 전환",
		"exit·reconcile·fill detection은 계속된다",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("page missing %q", want)
		}
	}
	if reader.calls != 1 {
		t.Fatalf("reader calls = %d", reader.calls)
	}
}

func TestMarketScheduleDoesNotRenderArbitraryErrorsEnumsOrProvenance(t *testing.T) {
	t.Run("reader error", func(t *testing.T) {
		h := newHarness(t, func(o *Options) {
			o.MarketSchedule = &stubMarketSchedule{err: context.Canceled}
		})
		h.authenticate(t)
		page := body(t, h.get(t, "/strategy-runtime/market-schedule"))
		if strings.Contains(page, context.Canceled.Error()) {
			t.Fatal("raw reader error reached operator page")
		}
		if !strings.Contains(page, "상태를 읽지 못해 닫힌 기본값을 표시한다") {
			t.Fatal("generic fail-closed explanation missing")
		}
	})

	t.Run("untrusted values", func(t *testing.T) {
		reader := &stubMarketSchedule{reading: MarketScheduleReading{
			SchedulerEffective: "/private/config/path", Market: "FREE_SYMBOL", Session: "after-hours",
			ApplyTiming: "run arbitrary now", CalendarSource: "http://untrusted.invalid",
			CalendarVersion: "not-a-digest", CalendarFetchedAt: time.Now(),
			DecisionReason: "operator supplied reason", NextTransition: time.Now(),
		}}
		h := newHarness(t, func(o *Options) { o.MarketSchedule = reader })
		h.authenticate(t)
		page := body(t, h.get(t, "/strategy-runtime/market-schedule"))
		for _, forbidden := range []string{
			"/private/config/path", "FREE_SYMBOL", "after-hours", "run arbitrary now",
			"untrusted.invalid", "not-a-digest", "operator supplied reason",
		} {
			if strings.Contains(page, forbidden) {
				t.Errorf("untrusted value %q reached page", forbidden)
			}
		}
		for _, want := range []string{"DISABLED", "NOT_ACTIVATED", "선택 시장 없음", "정규장", "검증되지 않음"} {
			if !strings.Contains(page, want) {
				t.Errorf("safe fallback %q missing", want)
			}
		}
	})
}

func TestMarketScheduleIsAuthenticatedReadOnlyAndHasNoFreeFormControls(t *testing.T) {
	h := newHarness(t, func(o *Options) { o.MarketSchedule = &stubMarketSchedule{} })
	if got := h.get(t, "/strategy-runtime/market-schedule"); got.StatusCode == http.StatusOK {
		t.Fatal("unauthenticated request reached schedule")
	}
	h.authenticate(t)
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		req, err := http.NewRequest(method, h.srv.URL+"/strategy-runtime/market-schedule", nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := h.client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("%s = %d", method, resp.StatusCode)
		}
	}
	resp := h.get(t, "/strategy-runtime/market-schedule")
	page := strings.ToLower(body(t, resp))
	for _, forbidden := range []string{
		`type="text"`, `type="number"`, "<textarea", "contenteditable", "<form", "<select",
		"symbol", "holiday", "휴장일 입력", "운영 사유 입력",
	} {
		if strings.Contains(page, forbidden) {
			t.Errorf("free-form/control surface contains %q", forbidden)
		}
	}
	if !strings.Contains(page, `name="viewport"`) {
		t.Error("mobile viewport missing")
	}
	for _, responsive := range []string{"@media (max-width: 720px)", "flex-wrap: wrap", "min-width: 0", "overflow-wrap: anywhere"} {
		if !strings.Contains(page, responsive) {
			t.Errorf("360px responsive contract missing %q", responsive)
		}
	}
	if !strings.Contains(page, "a:focus-visible") {
		t.Error("keyboard focus indicator missing")
	}
	if !strings.Contains(page, `<main>`) || !strings.Contains(page, `<h1>`) || !strings.Contains(page, `<table>`) {
		t.Error("semantic landmarks missing")
	}
	assertMarketScheduleDOM(t, page)
	wantCSP := "default-src 'none'; style-src 'unsafe-inline'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'"
	if got := resp.Header.Get("Content-Security-Policy"); got != wantCSP {
		t.Errorf("CSP = %q, want %q", got, wantCSP)
	}
}

func assertMarketScheduleDOM(t *testing.T, page string) {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(page))
	if err != nil {
		t.Fatalf("parse DOM: %v", err)
	}
	ids := map[string]bool{}
	labels := []string{}
	h1, main := 0, 0
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			switch node.Data {
			case "form", "input", "textarea", "select", "button":
				t.Errorf("forbidden control <%s> in schedule DOM", node.Data)
			case "h1":
				h1++
			case "main":
				main++
			}
			for _, attr := range node.Attr {
				switch attr.Key {
				case "id":
					ids[attr.Val] = true
				case "aria-labelledby":
					labels = append(labels, attr.Val)
				case "contenteditable":
					t.Errorf("contenteditable reached schedule DOM")
				case "tabindex":
					if strings.HasPrefix(attr.Val, "-") {
						t.Errorf("keyboard target removed from tab order: %q", attr.Val)
					}
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	if h1 != 1 || main != 1 {
		t.Errorf("landmarks h1=%d main=%d", h1, main)
	}
	for _, label := range labels {
		if !ids[label] {
			t.Errorf("aria-labelledby target %q missing", label)
		}
	}
}

func TestMarketScheduleUnwiredStillShowsClosedDefaults(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)
	page := body(t, h.get(t, "/strategy-runtime/market-schedule"))
	for _, want := range []string{"OFF", "선택 시장 없음", "정규장", "DISABLED", "seam 미배선"} {
		if !strings.Contains(page, want) {
			t.Errorf("unwired page missing %q", want)
		}
	}
}
