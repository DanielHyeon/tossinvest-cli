package main

// ledger.go 는 append-only JSONL 원장 — 행 종류 다섯(spec 「그림자 신호 원장은 모집단과 결손을 복원 가능하게 적는다」).

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	ledgerSchema   = "a140-signal-shadow/v1"
	ledgerFileName = "ledger.jsonl"

	kindUniverse = "universe" // 세션 시작에 고른 종목 집합(재시작 때 다시 쓰임)
	kindSignal   = "signal"   // 단계 · 거절이 바뀐 순간과 ARMED 시도마다
	kindLabel    = "label"    // 제안의 결과
	kindGap      = "gap"      // 읽기 포기 · 빈 분 · 세션 결손
	kindBudget   = "budget"   // 10 분마다 그룹별 GET 수 · 수집 시간
	kindSession  = "session"  // 세션 끝 요약

	doneProposed     = "proposed"
	doneArmedExpired = "armed_expired"
	doneInvalidated  = "invalidated"
	doneTimedOut     = "timed_out"
)

// ledgerRow 는 모든 행의 합집합(omitempty) — 행 종류마다 쓰는 칸이 다름.
type ledgerRow struct {
	Schema  string `json:"schema"`
	Kind    string `json:"kind"`
	At      string `json:"at"`
	Session string `json:"session"`
	Lot     string `json:"lot,omitempty"`
	Symbol  string `json:"symbol,omitempty"`

	// universe
	Symbols         []string         `json:"symbols,omitempty"`
	SessionSource   string           `json:"session_source,omitempty"`
	Params          *shadowParams    `json:"params,omitempty"`
	ParamsSource    string           `json:"params_source,omitempty"`
	Policy          *effectivePolicy `json:"effective_policy,omitempty"`
	PolicyDigest    string           `json:"policy_digest,omitempty"`
	SessionIdentity *sessionIdentity `json:"session_identity,omitempty"`

	// signal
	Phase          string   `json:"phase,omitempty"`
	Refusal        string   `json:"refusal,omitempty"`
	Done           string   `json:"done,omitempty"`
	Attempt        int      `json:"attempt,omitempty"`
	SetupID        string   `json:"setup_id,omitempty"`
	SnapshotDigest string   `json:"snapshot_digest,omitempty"`
	ConfigDigest   string   `json:"config_digest,omitempty"`
	MetricsDigest  string   `json:"metrics_digest,omitempty"`
	MetricsVersion string   `json:"metrics_version,omitempty"`
	Transitions    []string `json:"transitions,omitempty"`
	Bars           int      `json:"bars,omitempty"`
	Tick           uint64   `json:"tick,omitempty"`
	ATR            uint64   `json:"atr,omitempty"`
	Resistance     uint64   `json:"resistance,omitempty"`
	RangeLow       uint64   `json:"range_low,omitempty"`
	RVOL12         bool     `json:"rvol_1_2x,omitempty"`
	RVOL20         bool     `json:"rvol_2_0x,omitempty"`
	RVOL25         bool     `json:"rvol_2_5x,omitempty"`
	ArmedIndex     int      `json:"armed_index,omitempty"`
	Entry          uint64   `json:"entry,omitempty"`
	Ask            uint64   `json:"ask,omitempty"`
	Bid            uint64   `json:"bid,omitempty"`
	Worst          uint64   `json:"worst,omitempty"`
	Stop           uint64   `json:"stop,omitempty"`
	Target         uint64   `json:"target,omitempty"`
	RoundTrip      uint64   `json:"round_trip,omitempty"`
	Risk           uint64   `json:"risk,omitempty"`
	Quantity       uint64   `json:"quantity,omitempty"`
	SpreadPPM      uint64   `json:"spread_ppm,omitempty"`
	DriftPPM       uint64   `json:"drift_ppm,omitempty"`
	QuoteAgeMS     uint64   `json:"quote_age_ms,omitempty"`
	QuoteDigest    string   `json:"quote_digest,omitempty"`
	// LabelFromMS 는 라벨이 보기 시작하는 첫 봉의 여는 시각 — 제안 시각 뒤에 연 봉부터(진입 전 가격으로 결과를 매기지 않게).
	LabelFromMS uint64 `json:"label_from_ms,omitempty"`
	// ArmedSince 는 이 setup 이 ARMED 를 처음 본 시각(5 분 창의 시작 — 재시작 때 복원).
	ArmedSince string `json:"armed_since,omitempty"`

	// label
	Outcome    string `json:"outcome,omitempty"`
	ExitMinor  uint64 `json:"exit,omitempty"`
	ExitOpenAt string `json:"exit_bar_open_at,omitempty"`
	RPPM       int64  `json:"r_ppm,omitempty"`

	// gap
	Stage  string `json:"stage,omitempty"`
	Reason string `json:"reason,omitempty"`
	Minute string `json:"minute,omitempty"` // 결손이 난 분(그 분의 step 시각을 분으로 내림)

	// budget · session
	Reads          map[string]int `json:"reads,omitempty"`
	CollectSeconds float64        `json:"collect_seconds,omitempty"`
	Counts         map[string]int `json:"counts,omitempty"`
}

type ledger struct{ f *os.File }

// openLedger 는 O_APPEND 로만 엶 — 기존 행을 덮거나 자르는 경로가 코드에 없음.
func openLedger(dir string) (*ledger, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(dir, ledgerFileName), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}
	return &ledger{f: f}, nil
}

func (l *ledger) Close() error { return l.f.Close() }

// write 는 한 행을 한 줄로 쓰고 디스크까지 내림(장중 중단에도 앞 행이 남게).
func (l *ledger) write(row ledgerRow) error {
	row.Schema = ledgerSchema
	raw, err := json.Marshal(row)
	if err != nil {
		return err
	}
	if _, err := l.f.Write(append(raw, '\n')); err != nil {
		return err
	}
	return l.f.Sync()
}

// readLedger 는 원장 전체를 읽음. 없으면 빈 목록, 해석 못 하는 줄은 오류(조용히 건너뛰지 않음).
func readLedger(path string) ([]ledgerRow, error) {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var rows []ledgerRow
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	line := 0
	for scanner.Scan() {
		line++
		var row ledgerRow
		if err := json.Unmarshal(scanner.Bytes(), &row); err != nil {
			return nil, fmt.Errorf("%s line %d: %w", path, line, err)
		}
		if row.Schema != ledgerSchema {
			return nil, fmt.Errorf("%s line %d: schema %q is not %q", path, line, row.Schema, ledgerSchema)
		}
		rows = append(rows, row)
	}
	return rows, scanner.Err()
}
