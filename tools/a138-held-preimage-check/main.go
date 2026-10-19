// Command a138-held-preimage-check 는 엔진 이미지 교체 전 사람이 원장 **사본**에 돌리는 읽기 전용 확인 도구임.
//
// 왜 필요한가: a138(openspec/changes/archive/2026-10-15-a138-max-open-positions)은 예약 트랜잭션이 최대 보유 포지션 수를 셀 때
// HELD 진입 예약을 가진 결정마다 저장된 preimage 에서 (시장, 종목)을 읽게 했음. 그 preimage 하나라도 해석이 안 되면 개수를 셀 수
// 없으므로 **모든 새 진입이 fail-closed 로 거부됨**(design.md:84, :126 · review.md P3-6). 교체 직후 진입이 전부 막히는 일을
// 미리 알기 위해, 생산과 같은 판정으로 그런 결정이 원장에 있는지 셈.
//
// 판정은 재구현이 아니라 생산의 것임:
//   - 질의는 internal/journal/open_positions.go heldVenues 의 HELD 질의 문자열과 바이트 단위로 같음(main_test.go 가 소스를 파싱해 고정).
//   - 행마다 decisionVenue 와 같은 순서로 journal.ParsePreimage → journal.PreimageVenue → 빈 시장 · 종목 검사를 함
//     (decisionVenue 는 비공개라 본문 세 줄을 옮겨 적었고, 그 본문도 main_test.go 가 소스에서 고정함).
//
// 읽기 전용 보장: `file:<절대경로>?mode=ro` + `query_only(true)` 연결 하나(journal.OpenReadOnly 와 같은 계약). journal_mode 는 설정하지
// 않음(설정 자체가 쓰기). 원장을 만들거나 이관하지 않음 — 없는 경로는 도구 오류.
//
// 사본 만들기(엔진이 돌고 있어도 일관된 스냅숏 — sqlite3 CLI 는 이 호스트에 없음): 파이썬 표준 sqlite3 의 backup API 를 읽기 전용
// 원본 연결로 씀. 파일 복사(cp)는 WAL 의 커밋되지 않은 · 체크포인트 안 된 내용을 놓치거나 찢긴 사본을 만들 수 있어 쓰지 않음.
//
//	umask 077   # 사본에는 계좌 데이터가 있음 — 본인만 읽게 만들고 확인 뒤 지움
//	python3 -I -c 'import sqlite3,sys; s=sqlite3.connect("file:"+sys.argv[1]+"?mode=ro",uri=True); d=sqlite3.connect(sys.argv[2]); s.backup(d); d.close(); s.close()' \
//	    ~/.config/tossctl/journal.db /tmp/a138-journal-copy.db
//	go build -o /tmp/a138-held-preimage-check ./tools/a138-held-preimage-check
//	/tmp/a138-held-preimage-check --journal /tmp/a138-journal-copy.db; echo "exit=$?"
//	rm -f /tmp/a138-journal-copy.db /tmp/a138-journal-copy.db-wal /tmp/a138-journal-copy.db-shm
//
// `go run` 은 도구의 종료 코드를 1 로 뭉개므로(3 도 1 로 보임) 판정을 종료 코드로 읽으려면 빌드한 바이너리를 씀. 읽기 전용 연결도
// WAL 모드 사본 옆에 -shm · 빈 -wal 을 만들 수 있음(사본 본 파일의 바이트는 그대로).
//
// 출력에는 계좌 참조 · 금액 · preimage 내용이 없음 — 계좌는 #1..n 순번, 결정은 id 와 오류 문구만(오류 문구 속 따옴표 값은 가림).
//
// 종료 코드: 0 = CLEAN, 3 = BLOCKING(교체하면 모든 새 진입이 막힘), 1 = 도구 오류 — **도구 오류는 CLEAN 이 아님**(판정을 못 했다는 뜻).
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	_ "modernc.org/sqlite"
)

// 종료 코드 — 도구 오류(1)와 BLOCKING(3)을 갈라 사람이 「판정 실패」를 「깨끗함」으로 읽지 않게 함.
const (
	exitClean    = 0
	exitError    = 1
	exitBlocking = 3
)

// heldEntryQuery 는 생산 heldVenues 의 HELD 진입 예약 질의를 그대로 옮긴 것임 — 공백까지 같아야 함(main_test.go 가 소스와 대조).
const heldEntryQuery = `SELECT DISTINCT d.id, d.preimage_kind, d.risk_preimage
		   FROM risk_reservations r JOIN decisions d ON d.id = r.decision_id
		  WHERE r.account_ref = ? AND r.kind IN (?, ?, ?) AND r.state = ? AND d.safety_class = ?`

// heldEntryArgs 는 계좌 참조 뒤에 붙는 질의 인자임 — 생산과 같은 공개 상수를 같은 순서로 씀.
var heldEntryArgs = []any{
	journal.ReservationKindOpenExposure, journal.ReservationKindDailyLoss, journal.ReservationKindCash,
	journal.ReservationHeld, string(journal.SafetyClassExposureRaising),
}

// heldAccountsQuery 는 HELD 진입 예약 종류(같은 세 종류)를 가진 계좌 전부를 고름 — 생산은 예약하는 계좌 하나만 보지만 교체 뒤
// 어느 계좌가 진입하든 같은 판정을 거치므로 전부 셈. 정렬은 #순번을 실행마다 같게 하려는 것임.
const heldAccountsQuery = `SELECT DISTINCT account_ref FROM risk_reservations
		  WHERE kind IN (?, ?, ?) AND state = ? ORDER BY account_ref`

func main() {
	code, err := run(context.Background(), os.Args[1:], os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, "a138 held preimage check: TOOL ERROR (this is NOT a clean result): "+err.Error())
	}
	os.Exit(code)
}

// accountResult 는 계좌 하나의 HELD 진입 결정 수와 해석 불가 결정 목록임(계좌 참조 자체는 담지 않음).
type accountResult struct {
	held          int
	uninterpreted []string // "결정 id: 오류 문구"
}

func run(ctx context.Context, args []string, stdout io.Writer) (int, error) {
	flags := flag.NewFlagSet("a138-held-preimage-check", flag.ContinueOnError)
	journalPath := flags.String("journal", "", "required: a COPY of the ledger (production lives at ~/.config/tossctl/journal.db)")
	if err := flags.Parse(args); err != nil {
		return exitError, err
	}
	if flags.NArg() != 0 {
		return exitError, fmt.Errorf("unexpected arguments %v", flags.Args())
	}
	if *journalPath == "" {
		return exitError, errors.New("--journal <path> is required")
	}
	abs, err := filepath.Abs(*journalPath)
	if err != nil {
		return exitError, fmt.Errorf("resolving %s: %w", *journalPath, err)
	}
	// 드라이버에 묻기 전에 확인함 — mode=ro 는 파일을 만들지 않지만, 「원장이 없음」을 드라이버 문구가 아닌 분명한 오류로 냄.
	if info, err := os.Stat(abs); err != nil || info.IsDir() {
		return exitError, fmt.Errorf("no journal file at %s", abs)
	}
	fmt.Fprintln(stdout, "a138 held entry preimage check (read-only)")
	fmt.Fprintln(stdout, "reminder: run this on a COPY of the ledger made with the sqlite backup API, not on the live file")

	results, version, err := scan(ctx, abs)
	if err != nil {
		return exitError, err
	}
	fmt.Fprintf(stdout, "ledger schema version: %d\n", version)
	fmt.Fprintf(stdout, "accounts with held entry reservations: %d\n", len(results))
	blocking := 0
	for i, r := range results {
		fmt.Fprintf(stdout, "account #%d: %d held entry decision(s), %d uninterpretable\n", i+1, r.held, len(r.uninterpreted))
		for _, line := range r.uninterpreted {
			fmt.Fprintf(stdout, "  %s\n", line)
		}
		blocking += len(r.uninterpreted)
	}
	if blocking > 0 {
		fmt.Fprintf(stdout, "BLOCKING — %d held entry decision(s) would make every new entry fail closed after deploy\n", blocking)
		return exitBlocking, nil
	}
	fmt.Fprintln(stdout, "CLEAN — no held entry reservation has an uninterpretable preimage")
	return exitClean, nil
}

// scan 은 읽기 전용 연결 하나 · 읽기 트랜잭션 하나에서 계좌 목록과 계좌별 HELD 진입 결정을 읽고 판정함(한 스냅숏 안의 값끼리 비교).
func scan(ctx context.Context, abs string) ([]accountResult, int, error) {
	// 계약은 journal.readOnlyDSN 과 같음: mode=ro(파일 생성 · 쓰기 불가) + query_only(한 층 위의 같은 거부) + 체크포인트 대기 상한.
	db, err := sql.Open("sqlite", "file:"+abs+"?mode=ro&_pragma=query_only(true)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, 0, fmt.Errorf("opening %s read-only: %w", abs, err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	// query_only 가 실제로 켜졌는지 읽어 확인함 — DSN 이 조용히 무시되면 읽기 전용이라는 주장이 거짓이 되므로 도구 오류로 멈춤.
	var queryOnly int
	if err := db.QueryRowContext(ctx, "PRAGMA query_only").Scan(&queryOnly); err != nil {
		return nil, 0, fmt.Errorf("reading %s: %w", abs, err)
	}
	if queryOnly != 1 {
		return nil, 0, errors.New("the connection is not query_only; refusing to continue")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("starting a read transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // 읽기만 했으므로 되돌릴 것이 없음.

	var version int
	if err := tx.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		return nil, 0, fmt.Errorf("reading the schema version: %w", err)
	}
	accounts, err := heldAccounts(ctx, tx)
	if err != nil {
		return nil, 0, err
	}
	results := make([]accountResult, 0, len(accounts))
	for _, account := range accounts {
		r, err := judgeAccount(ctx, tx, account)
		if err != nil {
			return nil, 0, err
		}
		results = append(results, r)
	}
	return results, version, nil
}

func heldAccounts(ctx context.Context, tx *sql.Tx) ([]string, error) {
	rows, err := tx.QueryContext(ctx, heldAccountsQuery, heldEntryArgs[0], heldEntryArgs[1], heldEntryArgs[2], heldEntryArgs[3])
	if err != nil {
		return nil, fmt.Errorf("reading accounts with held reservations: %w", err)
	}
	defer rows.Close()
	var accounts []string
	for rows.Next() {
		var a string
		if err := rows.Scan(&a); err != nil {
			return nil, fmt.Errorf("reading accounts with held reservations: %w", err)
		}
		accounts = append(accounts, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading accounts with held reservations: %w", err)
	}
	return accounts, nil
}

// judgeAccount 는 생산 질의로 그 계좌의 HELD 진입 결정을 읽고, 결정마다 생산 decisionVenue 와 같은 판정을 함.
// 생산은 첫 실패에서 개수 판정 전체를 거부하지만, 이 도구는 사람이 고칠 대상을 다 보도록 실패를 모두 모음.
func judgeAccount(ctx context.Context, tx *sql.Tx, account string) (accountResult, error) {
	rows, err := tx.QueryContext(ctx, heldEntryQuery, append([]any{account}, heldEntryArgs...)...)
	if err != nil {
		return accountResult{}, fmt.Errorf("reading held entry reservations: %w", err)
	}
	defer rows.Close()
	var r accountResult
	for rows.Next() {
		var id, kind, preimage string
		if err := rows.Scan(&id, &kind, &preimage); err != nil {
			return accountResult{}, fmt.Errorf("reading held entry reservations: %w", err)
		}
		r.held++
		if err := venueError(kind, preimage); err != nil {
			r.uninterpreted = append(r.uninterpreted, fmt.Sprintf("decision %s: %s", id, redactQuoted(err.Error())))
		}
	}
	if err := rows.Err(); err != nil {
		return accountResult{}, fmt.Errorf("reading held entry reservations: %w", err)
	}
	return r, nil
}

// venueError 는 생산 decisionVenue 의 판정 부분임 — 해석 실패 또는 시장 · 종목이 빈 preimage 면 오류(생산은 이때 개수 판정을 거부).
func venueError(kind, preimage string) error {
	parsed, err := journal.ParsePreimage(kind, preimage)
	if err != nil {
		return fmt.Errorf("preimage: %w", err)
	}
	_, market, symbol, _ := journal.PreimageVenue(parsed)
	if strings.TrimSpace(market) == "" || strings.TrimSpace(symbol) == "" {
		return errors.New("names no market or symbol")
	}
	return nil
}

// quoted 는 오류 문구 속 큰따옴표 값임 — ParsePreimage 오류는 잘못된 수량 · 가격 · 종류를 %q 로 실으므로 가림(금액 · preimage 내용 비출력).
var quoted = regexp.MustCompile(`"[^"]*"`)

func redactQuoted(s string) string { return quoted.ReplaceAllString(s, `"…"`) }
