package obs_test

// a131_redact_test.go — 엔진 로거(obs.NewLogger)의 모든 줄이 계좌 규칙을 지남. 센티널은 계좌 참조 · 감싼 오류 · With 속성 ·
// 기본 속성 · 포지션 id 에 심음(design D7-1).

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/redact"
)

const a131Account = "987-65-OBSSENTINEL"
const a131Digits = "98765432109"

func init() { redact.Register(a131Account, a131Digits) }

func a131Leaks(out string) []string {
	var found []string
	for _, n := range []string{a131Account, a131Digits, "*******2109", "pos-00112233445566778899aabb"} {
		if strings.Contains(out, n) {
			found = append(found, n)
		}
	}
	return found
}

func TestA131EngineLoggerDropsAccountAndScrubsValues(t *testing.T) {
	var buf bytes.Buffer
	log := obs.NewLogger(obs.LogOptions{Writer: &buf, JSON: true})
	// Runtime.log 꼴 — 모든 줄 앞에 계좌.
	log.With(obs.FieldAccount, a131Account).Event(obs.EventReconcileClean, "symbol", "005930")
	// logErr 꼴 — 계좌 필드 + 감싼 오류 원문.
	err := fmt.Errorf("exit: %w", fmt.Errorf("journal: no position for account %s (masked *******2109)", a131Digits))
	log.Error(obs.EventEngineLoopDegraded, err, obs.FieldAccount, a131Account, obs.FieldDetail, "pos-00112233445566778899aabb")
	// 알림 선점 줄 꼴 — 값 안의 계좌.
	log.Warn(obs.EventEngineLoopDegraded, "claimed_by", "engine:"+a131Digits+":42")
	out := buf.String()
	if l := a131Leaks(out); len(l) != 0 {
		t.Fatalf("engine log leaked %v:\n%s", l, out)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 3 {
		t.Fatalf("want 3 lines, got %d:\n%s", len(lines), out)
	}
	for _, line := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("not JSON: %v: %s", err, line)
		}
		if _, ok := m[obs.FieldAccount]; ok {
			t.Fatalf("account field survived: %s", line)
		}
		if m[obs.FieldEvent] == nil || m[obs.FieldSeverity] == nil {
			t.Fatalf("event/severity lost: %s", line)
		}
	}
	if !strings.Contains(lines[0], `"symbol":"005930"`) {
		t.Fatalf("non-account field changed: %s", lines[0])
	}
}

func TestA131BaseAttrsAreScrubbed(t *testing.T) {
	var buf bytes.Buffer
	log := obs.NewLogger(obs.LogOptions{Writer: &buf, JSON: true,
		Base: []slog.Attr{slog.String(obs.FieldAccount, a131Account), slog.String("holder", "x"+a131Digits)}})
	log.Event(obs.EventReconcileClean)
	if l := a131Leaks(buf.String()); len(l) != 0 {
		t.Fatalf("base attributes leaked %v: %s", l, buf.String())
	}
	if strings.Contains(buf.String(), `"account"`) {
		t.Fatalf("base account key survived: %s", buf.String())
	}
}
