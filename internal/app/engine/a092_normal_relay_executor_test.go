package engine

// a092 C8 · K13: 일반 등급 이관 실행자는 자기 정지 이벤트 타입을 가짐 — 배달 실행자의 EventAlertUndelivered 를 빌리지 않음.

import (
	"bytes"
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

func TestA092TheRelayStopIsNotAnUndeliveredCriticalAlert(t *testing.T) {
	buf := &bytes.Buffer{}
	r := &Runtime{opts: RuntimeOptions{Log: obs.NewLogger(obs.LogOptions{Writer: buf, JSON: true, Clock: clock.System()})}}
	aux := (&Context{}).NormalAlertRelayExecutor()
	aux.Run = func(context.Context) error { return errors.New("relay broke") }
	r.runAuxiliary(context.Background(), aux)
	log := buf.String()
	if !strings.Contains(log, string(obs.EventNormalAlertRelayStopped)) {
		t.Errorf("the relay stop was not recorded under its own type:\n%s", log)
	}
	if strings.Contains(log, string(obs.EventAlertUndelivered)) {
		t.Errorf("the relay stop borrowed the critical delivery event:\n%s", log)
	}
}

// 배달 실행자는 그대로(StopEvent 빈 값 → EventAlertUndelivered).
func TestA092TheDelivererStopKeepsItsEvent(t *testing.T) {
	buf := &bytes.Buffer{}
	r := &Runtime{opts: RuntimeOptions{Log: obs.NewLogger(obs.LogOptions{Writer: buf, JSON: true, Clock: clock.System()})}}
	r.runAuxiliary(context.Background(), AuxiliaryExecutor{Name: alertDeliveryName,
		Run: func(context.Context) error { return errors.New("broke") }})
	if !strings.Contains(buf.String(), string(obs.EventAlertUndelivered)) {
		t.Errorf("the delivery executor's stop lost its event:\n%s", buf.String())
	}
}

// Context 는 한 이관을 두 소비자(exit 관측기 · 런타임 실행자)에 같은 인스턴스로 줌.
func TestA092TheRelayIsOneInstancePerEngine(t *testing.T) {
	c := &Context{Notifier: &obs.Notifier{}}
	if c.NormalAlertRelay() == nil || c.NormalAlertRelay() != c.NormalAlertRelay() {
		t.Fatal("the relay is not a single instance")
	}
}

// 25라운드 보이스 A #2: 배달 실행자의 만료 임차 인수는 claim_stolen — claim_held(알림기에서는 정상 경합 INFO)와 이름을 가름.
func TestA092TheDelivererNamesATakeoverAsStolen(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "alertdelivery.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		c, ok := n.(*ast.CallExpr)
		if !ok || len(c.Args) < 3 {
			return true
		}
		lit, ok := c.Args[2].(*ast.BasicLit)
		if !ok || !strings.Contains(lit.Value, "expired alert lease was taken over") {
			return true
		}
		found = true
		if s, ok := c.Args[0].(*ast.SelectorExpr); !ok || s.Sel.Name != "EventAlertClaimStolen" {
			t.Errorf("the takeover line is logged as %s at %s, want EventAlertClaimStolen", a092Expr(c.Args[0]), fset.Position(c.Pos()))
		}
		return true
	})
	if !found {
		t.Fatal("control: the takeover log line was not found — the pin measures nothing")
	}
}

func a092Expr(e ast.Expr) string {
	if s, ok := e.(*ast.SelectorExpr); ok {
		return s.Sel.Name
	}
	return "?"
}
