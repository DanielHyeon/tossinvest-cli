package main

// main_test.go 는 ① 생산 판정과의 표류 고정(질의 문자열 · 인자 · decisionVenue 본문) ② 임시 원장에서의 판정 ③ 읽기 전용을 잼.
// 운영 원장 · 실제 설정 디렉터리는 건드리지 않음 — 원장은 전부 t.TempDir() 에 생산 journal.Open 으로 만듦.

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

const productionSource = "../../internal/journal/open_positions.go"

// parseProduction 은 생산 open_positions.go 를 파싱함(시험의 작업 디렉터리는 이 패키지 디렉터리).
func parseProduction(t *testing.T) (*token.FileSet, *ast.File) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, productionSource, nil, 0)
	if err != nil {
		t.Fatalf("parsing %s: %v", productionSource, err)
	}
	return fset, f
}

// TestTheQueryIsProductionsHeldQueryByteForByte 는 생산 HELD 질의 문자열 · 인자가 바뀌면 실패함 — 그때 도구를 다시 맞출 것.
func TestTheQueryIsProductionsHeldQueryByteForByte(t *testing.T) {
	fset, f := parseProduction(t)
	var calls []*ast.CallExpr
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		for _, arg := range call.Args {
			if lit, ok := arg.(*ast.BasicLit); ok && lit.Kind == token.STRING &&
				strings.Contains(lit.Value, "FROM risk_reservations r JOIN decisions d") {
				calls = append(calls, call)
			}
		}
		return true
	})
	// 정확히 하나여야 함 — 둘이면 어느 것이 개수 판정인지 이 시험이 고를 수 없음.
	if len(calls) != 1 {
		t.Fatalf("found %d calls carrying the held reservation query in %s, want exactly 1", len(calls), productionSource)
	}
	call := calls[0]
	// 모양: tx.QueryContext(ctx, <질의>, accountRef, <인자 다섯>).
	if len(call.Args) != 8 {
		t.Fatalf("the production held query call has %d arguments, want 8 (ctx, query, account, five constants)", len(call.Args))
	}
	sqlText, err := strconv.Unquote(call.Args[1].(*ast.BasicLit).Value)
	if err != nil {
		t.Fatal(err)
	}
	if sqlText != heldEntryQuery {
		t.Fatalf("the production held query drifted from the tool's copy:\nproduction: %q\ntool:       %q", sqlText, heldEntryQuery)
	}
	var printed []string
	for _, arg := range call.Args[2:] {
		var b bytes.Buffer
		if err := printer.Fprint(&b, fset, arg); err != nil {
			t.Fatal(err)
		}
		printed = append(printed, b.String())
	}
	wantProduction := []string{"accountRef", "ReservationKindOpenExposure", "ReservationKindDailyLoss", "ReservationKindCash",
		"ReservationHeld", "string(SafetyClassExposureRaising)"}
	if !reflect.DeepEqual(printed, wantProduction) {
		t.Fatalf("the production held query arguments drifted: got %v, want %v", printed, wantProduction)
	}
	wantTool := []any{journal.ReservationKindOpenExposure, journal.ReservationKindDailyLoss, journal.ReservationKindCash,
		journal.ReservationHeld, string(journal.SafetyClassExposureRaising)}
	if !reflect.DeepEqual(heldEntryArgs, wantTool) {
		t.Fatalf("the tool's query arguments are %v, want %v", heldEntryArgs, wantTool)
	}
}

// productionDecisionVenue 는 이 도구의 venueError 가 옮겨 적은 생산 decisionVenue 의 본문임(주석 제외 인쇄형).
const productionDecisionVenue = `func decisionVenue(id, kind, preimage string) (positionVenue, error) {
	parsed, err := ParsePreimage(kind, preimage)
	if err != nil {
		return positionVenue{}, fmt.Errorf("journal: decision %s preimage: %w", id, err)
	}
	_, market, symbol, _ := PreimageVenue(parsed)
	if strings.TrimSpace(market) == "" || strings.TrimSpace(symbol) == "" {
		return positionVenue{}, fmt.Errorf("journal: decision %s names no market or symbol", id)
	}
	return venueOf(market, symbol), nil
}`

// TestTheJudgementIsProductionsDecisionVenue 는 생산 decisionVenue 가 한 글자라도 바뀌면 실패함 — 옮겨 적은 판정이 낡았다는 신호.
func TestTheJudgementIsProductionsDecisionVenue(t *testing.T) {
	fset, f := parseProduction(t)
	var found []*ast.FuncDecl
	for _, d := range f.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == "decisionVenue" {
			found = append(found, fn)
		}
	}
	if len(found) != 1 {
		t.Fatalf("found %d decisionVenue declarations, want 1", len(found))
	}
	var b bytes.Buffer
	if err := printer.Fprint(&b, fset, found[0]); err != nil {
		t.Fatal(err)
	}
	if b.String() != productionDecisionVenue {
		t.Fatalf("production decisionVenue changed — re-check venueError against it:\n%s", b.String())
	}
}

// ---- 임시 원장 --------------------------------------------------------------

// tempLedger 는 생산 journal.Open 으로 스키마를 만든 뒤 닫은 임시 원장 경로임.
func tempLedger(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "journal.db")
	j, err := journal.Open(context.Background(), journal.Options{
		Path:     path,
		Clock:    clock.NewFake(time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC)),
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

// validPreimage 는 생산 RecordDecision 이 쓰는 정규형 RISK_INTENT preimage 임.
func validPreimage(t *testing.T, account, symbol string) string {
	t.Helper()
	p, err := journal.RiskIntent{AccountRef: account, Market: "kr", Symbol: symbol, Side: "BUY", Quantity: "1",
		EntryPrice: "100", StopPrice: "90", PolicyVersion: "test-1"}.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// seed 는 결정 하나와 그 예약 하나를 원시 INSERT 로 넣음 — a138 원장 시험의 seedExistingRiskReservation 과 같은 열 · 값 모양
// (해석 불가 preimage 는 공개 API 로 만들 수 없으므로 원시 INSERT 를 씀). 쓰기 연결은 이 시험의 시드에만 쓰고 닫음.
func seed(t *testing.T, path, id, account, safetyClass, preimage, state string) {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`INSERT INTO decisions(id,account_ref,generation,safety_class,preimage_kind,risk_preimage,risk_hash,nonce,issued_at,expires_at)
		VALUES(?,?,0,?,'RISK_INTENT',?,?,?,?,?)`, id, account, safetyClass, preimage, "hash-"+id, "nonce-"+id,
		"2026-10-19T00:29:00Z", "2026-10-19T01:30:00Z"); err != nil {
		t.Fatal(err)
	}
	var releasedAt, reason any
	if state == journal.ReservationReleased {
		releasedAt, reason = "2026-10-19T00:40:00Z", "OPERATOR"
	}
	if _, err := db.Exec(`INSERT INTO risk_reservations(id,decision_id,account_ref,kind,amount,currency,snapshot_as_of,state,released_at,release_reason)
		VALUES(?,?,?,'OPEN_EXPOSURE','1','KRW','2026-10-19T00:29:00Z',?,?,?)`, "res-"+id, id, account, state, releasedAt, reason); err != nil {
		t.Fatal(err)
	}
}

func runTool(t *testing.T, path string) (int, string) {
	t.Helper()
	var out bytes.Buffer
	code, err := run(context.Background(), []string{"--journal", path}, &out)
	if err != nil {
		t.Fatalf("tool error: %v\n%s", err, out.String())
	}
	return code, out.String()
}

func TestAValidHeldEntryIsClean(t *testing.T) {
	path := tempLedger(t)
	seed(t, path, "dec-ok", "acct-secret-1", journal.SafetyClassExposureRaising, validPreimage(t, "acct-secret-1", "005930"), journal.ReservationHeld)
	code, out := runTool(t, path)
	if code != exitClean || !strings.Contains(out, "CLEAN — no held entry reservation has an uninterpretable preimage") {
		t.Fatalf("code=%d, want CLEAN:\n%s", code, out)
	}
	if !strings.Contains(out, "account #1: 1 held entry decision(s), 0 uninterpretable") {
		t.Fatalf("the held decision was not counted:\n%s", out)
	}
	if strings.Contains(out, "acct-secret-1") || strings.Contains(out, "005930") {
		t.Fatalf("the output leaks an account reference or a preimage value:\n%s", out)
	}
}

func TestAMalformedHeldEntryIsBlockingAndNamed(t *testing.T) {
	path := tempLedger(t)
	seed(t, path, "dec-ok", "acct-1", journal.SafetyClassExposureRaising, validPreimage(t, "acct-1", "005930"), journal.ReservationHeld)
	seed(t, path, "dec-bad", "acct-1", journal.SafetyClassExposureRaising, `{}`, journal.ReservationHeld)
	// 두 번째 계좌의 정규형이 아닌(값은 맞지만 철자가 다른) preimage — 생산은 정규형 왕복이 깨져도 거부함.
	seed(t, path, "dec-noncanon", "acct-2", journal.SafetyClassExposureRaising,
		strings.Replace(validPreimage(t, "acct-2", "000660"), `"quantity":"1"`, `"quantity":"1.0"`, 1), journal.ReservationHeld)
	// 숫자가 아닌 수량 — 생산 오류 문구가 그 값을 %q 로 싣지만 출력에는 가려져야 함.
	seed(t, path, "dec-amount", "acct-2", journal.SafetyClassExposureRaising,
		strings.Replace(validPreimage(t, "acct-2", "035720"), `"quantity":"1"`, `"quantity":"98,765"`, 1), journal.ReservationHeld)
	code, out := runTool(t, path)
	if code != exitBlocking {
		t.Fatalf("code=%d, want %d (BLOCKING):\n%s", code, exitBlocking, out)
	}
	for _, want := range []string{
		"account #1: 2 held entry decision(s), 1 uninterpretable",
		"decision dec-bad: preimage: ",
		"account #2: 2 held entry decision(s), 2 uninterpretable",
		"decision dec-noncanon: preimage: ",
		"decision dec-amount: preimage: ",
		"BLOCKING — 3 held entry decision(s) would make every new entry fail closed after deploy",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "decision dec-ok") || strings.Contains(out, "acct-") || strings.Contains(out, "98,765") {
		t.Fatalf("the output names a good decision or an account reference:\n%s", out)
	}
}

func TestReleasedAndNonEntryReservationsAreNotCounted(t *testing.T) {
	path := tempLedger(t)
	// 해제된 예약 · 위험 축소 결정은 생산 질의가 고르지 않음 — preimage 가 망가져도 개수 판정을 막지 않음.
	seed(t, path, "dec-released", "acct-1", journal.SafetyClassExposureRaising, `{}`, journal.ReservationReleased)
	seed(t, path, "dec-reducing", "acct-1", journal.SafetyClassRiskReducing, `{}`, journal.ReservationHeld)
	code, out := runTool(t, path)
	if code != exitClean || !strings.Contains(out, "CLEAN — ") {
		t.Fatalf("code=%d, want CLEAN:\n%s", code, out)
	}
	// 위험 축소 결정의 HELD 예약 때문에 계좌는 목록에 오르지만 진입 결정은 0 건임.
	if !strings.Contains(out, "account #1: 0 held entry decision(s), 0 uninterpretable") {
		t.Fatalf("unexpected per-account line:\n%s", out)
	}
}

func TestAnEmptyLedgerIsClean(t *testing.T) {
	code, out := runTool(t, tempLedger(t))
	if code != exitClean || !strings.Contains(out, "accounts with held entry reservations: 0") {
		t.Fatalf("code=%d:\n%s", code, out)
	}
}

func TestToolErrorsAreNotClean(t *testing.T) {
	notSQLite := filepath.Join(t.TempDir(), "junk.db")
	if err := os.WriteFile(notSQLite, []byte("not a database, just text that is long enough to have a header"), 0o600); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(t.TempDir(), "missing.db")
	for _, args := range [][]string{
		{},
		{"--journal", missing},
		{"--journal", t.TempDir()},
		{"--journal", notSQLite},
		{"--journal", tempLedger(t), "extra"},
	} {
		var out bytes.Buffer
		code, err := run(context.Background(), args, &out)
		if code != exitError || err == nil {
			t.Errorf("args %v: code=%d err=%v, want a tool error (exit 1)", args, code, err)
		}
		if strings.Contains(out.String(), "CLEAN") {
			t.Errorf("args %v: a tool error printed CLEAN:\n%s", args, out.String())
		}
	}
	// 없는 경로에 원장을 만들지 않음.
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Fatalf("the tool created %s", missing)
	}
}

func fileSum(t *testing.T, path string) [32]byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return sha256.Sum256(b)
}

// TestTheToolDoesNotWriteTheLedger 는 원장 본 파일의 바이트가 그대로이고 WAL 에 프레임이 생기지 않음을 잼.
func TestTheToolDoesNotWriteTheLedger(t *testing.T) {
	path := tempLedger(t)
	seed(t, path, "dec-ok", "acct-1", journal.SafetyClassExposureRaising, validPreimage(t, "acct-1", "005930"), journal.ReservationHeld)
	seed(t, path, "dec-bad", "acct-1", journal.SafetyClassExposureRaising, `{}`, journal.ReservationHeld)
	before := fileSum(t, path)
	for i := 0; i < 2; i++ {
		if code, out := runTool(t, path); code != exitBlocking {
			t.Fatalf("run %d: code=%d:\n%s", i, code, out)
		}
	}
	if fileSum(t, path) != before {
		t.Fatal("the ledger file changed although the tool is read-only")
	}
	if info, err := os.Stat(path + "-wal"); err == nil && info.Size() > 0 {
		t.Fatalf("the tool left %d bytes of WAL frames", info.Size())
	}
}
