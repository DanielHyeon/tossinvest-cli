package engine

// a130_trade_outcome_refusal_log_test.go 는 a130 design D3 의 엔진 쪽 시험임 — 성과 동결 거절이 engine.log 에 계좌 없는 줄
// 하나로 남고, 그 쓰기가 체결 트랜잭션(호출자)을 붙잡지 않으며, 생산 기동이 엔진 로거로 싱크를 묶음.

import (
	"bytes"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// a130Buffer 는 동시 쓰기에 안전한 로그 대상임.
type a130Buffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *a130Buffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *a130Buffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// a130WaitLine 은 줄 하나가 쓰일 때까지 기다림(쓰기는 별도 goroutine).
func a130WaitLine(t *testing.T, b *a130Buffer) string {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if s := b.String(); strings.Contains(s, "\n") {
			return s
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("no refusal line was written within 5s")
	return ""
}

func TestA130NoLoggerMeansNoRefusalSink(t *testing.T) {
	if sink := tradeOutcomeRefusalSink(nil); sink != nil {
		t.Error("a nil logger must bind no sink, so the fill path is exactly as before a130")
	}
}

// TestA130TheRefusalLineCarriesOnlyItsFields: 줄의 키 = obs 공통 키 + {symbol, market, instance_seq, reason}.
func TestA130TheRefusalLineCarriesOnlyItsFields(t *testing.T) {
	out := &a130Buffer{}
	sink := tradeOutcomeRefusalSink(obs.NewLogger(obs.LogOptions{Writer: out, JSON: true}))
	if sink == nil {
		t.Fatal("a logger must bind a sink")
	}
	sink(journal.TradeOutcomeRefusal{Symbol: "010170", Market: "kr", InstanceSeq: 3, Reason: "sell_quantity_mismatch"})

	var line map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(a130WaitLine(t, out))), &line); err != nil {
		t.Fatalf("the line is not one JSON object: %v", err)
	}
	if line["event"] != "trade_outcome.not_frozen" || line["level"] != "WARN" {
		t.Errorf("event/level = %v/%v, want trade_outcome.not_frozen/WARN", line["event"], line["level"])
	}
	var keys []string
	for k := range line {
		switch k {
		case "time", "level", "msg", "event", "subject", "severity":
		default:
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	if want := []string{"instance_seq", "market", "reason", "symbol"}; strings.Join(keys, ",") != strings.Join(want, ",") {
		t.Errorf("line keys = %v, want %v", keys, want)
	}
	if line["symbol"] != "010170" || line["market"] != "kr" || line["reason"] != "sell_quantity_mismatch" ||
		line["instance_seq"] != float64(3) {
		t.Errorf("line = %v", line)
	}
}

// a130BlockingWriter 는 release 가 닫힐 때까지 쓰기를 붙잡음 — 막힌 디스크.
type a130BlockingWriter struct{ release chan struct{} }

func (w a130BlockingWriter) Write(p []byte) (int, error) {
	<-w.release
	return len(p), nil
}

// TestA130ABlockedLogWriterDoesNotHoldTheFillTransaction: 싱크는 체결 트랜잭션 안에서 불림 — 쓰기가 막혀도 호출은 곧바로 돌아와야 함.
func TestA130ABlockedLogWriterDoesNotHoldTheFillTransaction(t *testing.T) {
	w := a130BlockingWriter{release: make(chan struct{})}
	defer close(w.release)
	sink := tradeOutcomeRefusalSink(obs.NewLogger(obs.LogOptions{Writer: w, JSON: true}))

	done := make(chan struct{})
	go func() {
		sink(journal.TradeOutcomeRefusal{Symbol: "005930", Market: "kr", InstanceSeq: 1, Reason: "insert_failed"})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("the sink blocked on the log writer; it would hold the SQLite write lock the stop-loss path waits on")
	}
}

// TestA130TheLoggedBindingCarriesTheSink: 싱크를 넘기면 묶이고, 인자 없는 bindApplyHooks(시험 14곳)는 a130 이전과 같은 바인딩임.
func TestA130TheLoggedBindingCarriesTheSink(t *testing.T) {
	logged := openTestJournal(t)
	if err := bindApplyHooks(logged, tradeOutcomeRefusalSink(obs.NewLogger(obs.LogOptions{Writer: &a130Buffer{}}))); err != nil {
		t.Fatal(err)
	}
	if !logged.OutcomeRefusalBound() {
		t.Error("bindApplyHooks with the engine sink must bind the refusal sink")
	}
	plain := openTestJournal(t)
	if err := bindApplyHooks(plain); err != nil {
		t.Fatal(err)
	}
	if plain.OutcomeRefusalBound() {
		t.Error("bindApplyHooks (no logger) must bind no sink")
	}
	if !plain.ExitApplierBound() || !plain.ProjectionBound() {
		t.Error("bindApplyHooks must still bind the projection and the exit applier")
	}
}

// TestA130NewContextBindsTheHooksWithTheEngineLogger 는 구조 핀임: NewContext 의 훅 바인딩은
// bindApplyHooks(jrn, tradeOutcomeRefusalSink(opts.Logger)) 하나이고, 싱크 없는 호출은 없음.
func TestA130NewContextBindsTheHooksWithTheEngineLogger(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "engine.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var body *ast.BlockStmt
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil && fn.Name.Name == "NewContext" {
			body = fn.Body
		}
	}
	if body == nil {
		t.Fatal("NewContext not found in engine.go")
	}
	var logged, plain int
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if ident, ok := call.Fun.(*ast.Ident); !ok || ident.Name != "bindApplyHooks" {
			return true
		}
		if len(call.Args) != 2 {
			plain++
			return true
		}
		sink, ok := call.Args[1].(*ast.CallExpr)
		if !ok || len(sink.Args) != 1 {
			plain++
			return true
		}
		fn, _ := sink.Fun.(*ast.Ident)
		sel, _ := sink.Args[0].(*ast.SelectorExpr)
		if fn != nil && fn.Name == "tradeOutcomeRefusalSink" && sel != nil && sel.Sel.Name == "Logger" {
			if x, ok := sel.X.(*ast.Ident); ok && x.Name == "opts" {
				logged++
				return true
			}
		}
		plain++
		return true
	})
	if logged != 1 || plain != 0 {
		t.Errorf("NewContext binds with the engine logger's sink %d time(s) and otherwise %d time(s), want 1 and 0", logged, plain)
	}
}
