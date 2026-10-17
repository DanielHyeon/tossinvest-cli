package main

// manual_session_test.go 는 --session 수동 세션창 우회를 고정함(달력이 틀렸을 때 사람이
// 명시적으로 세션창을 공급하는 수단 — 2026-10-12 KR 달력 오응답 대응).

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestManualSessionBypassesCalendarKR 는 핵심 계약 셋을 봄:
//  1. --session 지정 시 TypedMarketCalendar 가 한 번도 안 불림(mock 호출 수 0).
//  2. 지정 창의 Open/Close 가 KR 현지 오늘 날짜 + 시각(Asia/Seoul)으로 맞음.
//  3. 우회가 원장(gap stage "session")과 state(session_source="manual-override")에 남음.
func TestManualSessionBypassesCalendarKR(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	rig.probe.cfg.ManualWindow = &manualWindow{StartHour: 9, StartMin: 0, EndHour: 15, EndMin: 30, Raw: "09:00-15:30"}
	// 시계는 at(10,0) = 2026-10-07 10:00 KST → KR 현지 날짜 = 2026-10-07.
	sess, err := rig.probe.resolveSession(context.Background(), "KR")
	if err != nil {
		t.Fatal(err)
	}
	if got := rig.official.callCount("/api/v1/market-calendar/KR"); got != 0 {
		t.Fatalf("manual --session must not call the market calendar; calls = %d", got)
	}
	if sess.Source != sessionSourceManual {
		t.Fatalf("session source = %q, want %q", sess.Source, sessionSourceManual)
	}
	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		t.Fatal(err)
	}
	wantOpen := time.Date(2026, 10, 7, 9, 0, 0, 0, loc)
	wantClose := time.Date(2026, 10, 7, 15, 30, 0, 0, loc)
	if !sess.Open.Equal(wantOpen) || !sess.Close.Equal(wantClose) {
		t.Fatalf("manual window = %s..%s, want %s..%s", sess.Open, sess.Close, wantOpen, wantClose)
	}

	gaps := rig.contents(t).Gaps
	found := 0
	for _, g := range gaps {
		if g.Market == "KR" && g.Stage == "session" {
			found++
			if !strings.Contains(g.Reason, "calendar bypassed by --session, window=09:00-15:30, reason=operator") {
				t.Fatalf("override gap reason = %q", g.Reason)
			}
		}
	}
	if found != 1 {
		t.Fatalf("want exactly one session-bypass gap row, got %d: %+v", found, gaps)
	}

	state := assembleState(marketInputs{Market: "KR", Session: sess})
	if state.SessionSource != sessionSourceManual {
		t.Fatalf("state.session_source = %q, want %q", state.SessionSource, sessionSourceManual)
	}
}

// TestManualSessionUsesEasternDateUS 는 US 창이 America/New_York 현지 날짜 + 시각으로 맞고
// DST(EDT = -04:00)를 반영하는지 봄. 현지 날짜는 "지금"의 시장 현지 날짜임을 명시적 instant 로 고정.
func TestManualSessionUsesEasternDateUS(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	rig.probe.cfg.ManualWindow = &manualWindow{StartHour: 9, StartMin: 30, EndHour: 16, EndMin: 0, Raw: "09:30-16:00"}
	rig.clock.Set(time.Date(2026, 10, 7, 15, 0, 0, 0, time.UTC)) // = 2026-10-07 11:00 EDT
	sess, err := rig.probe.resolveSession(context.Background(), "US")
	if err != nil {
		t.Fatal(err)
	}
	if got := rig.official.callCount("/api/v1/market-calendar/US"); got != 0 {
		t.Fatalf("manual --session must not call the US calendar; calls = %d", got)
	}
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	wantOpen := time.Date(2026, 10, 7, 9, 30, 0, 0, loc)
	wantClose := time.Date(2026, 10, 7, 16, 0, 0, 0, loc)
	if !sess.Open.Equal(wantOpen) || !sess.Close.Equal(wantClose) {
		t.Fatalf("US manual window = %s..%s, want %s..%s", sess.Open.UTC(), sess.Close.UTC(), wantOpen.UTC(), wantClose.UTC())
	}
	// DST 못 박기: 09:30 ET(EDT -04:00) = 13:30 UTC. tz 가 EST(-05:00)로 틀리면 14:30Z 가 됨.
	if !sess.Open.Equal(time.Date(2026, 10, 7, 13, 30, 0, 0, time.UTC)) {
		t.Fatalf("US open in UTC = %s, want 13:30Z (EDT -04:00)", sess.Open.UTC())
	}
}

// TestCalendarPathUnchangedWithoutOverride 는 --session 미지정 시 종전 달력 경로가 불변임을 봄:
// 달력을 실제로 조회하고(양성 대조군), source="calendar", session-bypass gap 은 없음.
func TestCalendarPathUnchangedWithoutOverride(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t) // ManualWindow nil
	sess, err := rig.probe.resolveSession(context.Background(), "KR")
	if err != nil {
		t.Fatal(err)
	}
	if sess.Source != sessionSourceCalendar {
		t.Fatalf("default session source = %q, want %q", sess.Source, sessionSourceCalendar)
	}
	if rig.official.callCount("/api/v1/market-calendar/KR") == 0 {
		t.Fatal("the default path must consult the market calendar")
	}
	for _, g := range rig.contents(t).Gaps {
		if g.Stage == "session" {
			t.Fatalf("calendar path must not write a session-bypass gap: %+v", g)
		}
	}
}

// TestParseManualWindowRejects 는 파싱 실패·역전창이 시작 전에 거부됨을 봄.
func TestParseManualWindowRejects(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{
		"",            // 빈 값(구분자 없음)
		"0900-1530",   // 콜론 없음
		"09:00",       // 구분자 없음
		"24:00-25:00", // 시 범위 초과
		"09:61-10:00", // 분 범위 초과
		"15:30-09:00", // 역전창
		"09:00-09:00", // start == end
	} {
		if w, err := parseManualWindow(raw); err == nil {
			t.Fatalf("parseManualWindow(%q) = %+v, want error", raw, w)
		}
	}
}

// TestParseManualWindowAccepts 는 올바른 창이 그대로 파싱됨을 봄.
func TestParseManualWindowAccepts(t *testing.T) {
	t.Parallel()
	w, err := parseManualWindow("09:00-15:30")
	if err != nil {
		t.Fatal(err)
	}
	if w.StartHour != 9 || w.StartMin != 0 || w.EndHour != 15 || w.EndMin != 30 || w.Raw != "09:00-15:30" {
		t.Fatalf("parsed = %+v", w)
	}
}
