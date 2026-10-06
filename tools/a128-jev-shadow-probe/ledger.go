package main

// ledger.go 는 JSONL 원장(append-only)과 state 원문 보관을 담당함(design.md 「원장」).
//
// 행 종류:
//   - judgment: 판단 하나(질문 하나) — 시각·시장·심볼·질문 id/rev/digest·state digest·p·기준가·결손.
//   - label:    judgment 참조 + 실현가·실현 수익률, 또는 라벨 결손 사유.
//   - gap:      판단 행 자체가 생기지 못한 결손(랭킹·시세 읽기 포기, 판단 API 포기 등).
//     design.md 는 2종을 적었으나, 판단 행이 없으면 결손을 실을 행이 없어 침묵 생략이 됨 →
//     spec 「수집 결손은 행에 명시되어야 한다」를 지키려고 셋째 종류를 둠.
//
// state 원문은 원장에 넣지 않고 digest 만 넣음. 원문은 states/<digest>.json 으로 따로 보관.

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	ledgerFileName = "ledger.jsonl"
	statesDirName  = "states"
	ledgerSchema   = "a128-shadow-ledger/v1"

	kindJudgment = "judgment"
	kindLabel    = "label"
	kindGap      = "gap"
)

// judgmentRow — 판단 행. 모든 필드는 공개 시장 데이터 또는 질문·모델 메타뿐임.
type judgmentRow struct {
	Schema         string   `json:"schema"`
	Kind           string   `json:"kind"`
	ID             string   `json:"id"`
	At             string   `json:"at"`
	Market         string   `json:"market"`
	Symbol         string   `json:"symbol"`
	QuestionID     string   `json:"question_id"`
	QuestionRev    string   `json:"question_rev"`
	QuestionDigest string   `json:"question_digest"`
	ModelRequested string   `json:"model_requested"`
	ModelAnswered  string   `json:"model_answered"`
	StateDigest    string   `json:"state_digest"`
	P              float64  `json:"p"`
	BasePrice      float64  `json:"base_price"`
	Currency       string   `json:"currency"`
	HorizonMinutes int      `json:"horizon_minutes"`
	LabelDueAt     string   `json:"label_due_at"`
	InputTokens    int      `json:"input_tokens"`
	CollectionGaps []string `json:"collection_gaps"`
}

// labelRow — 라벨 행. 실현값이 없으면 LabelGap 이 비어 있지 않아야 함(둘 중 하나는 반드시).
type labelRow struct {
	Schema         string   `json:"schema"`
	Kind           string   `json:"kind"`
	JudgmentID     string   `json:"judgment_id"`
	At             string   `json:"at"`
	DueAt          string   `json:"due_at"`
	RealizedPrice  *float64 `json:"realized_price"`
	RealizedReturn *float64 `json:"realized_return"`
	LabelGap       string   `json:"label_gap,omitempty"`
}

// gapRow — 판단 행을 만들지 못한 결손.
type gapRow struct {
	Schema string `json:"schema"`
	Kind   string `json:"kind"`
	At     string `json:"at"`
	Market string `json:"market"`
	Stage  string `json:"stage"`
	Symbol string `json:"symbol,omitempty"`
	Reason string `json:"reason"`
}

// ledger 는 여러 고루틴(시장별 수집 + 라벨러)이 함께 쓰므로 한 줄 쓰기를 잠금으로 묶음.
type ledger struct {
	dir string
	mu  sync.Mutex
	f   *os.File
}

// openLedger 는 O_APPEND 로만 엶 — 기존 행을 덮거나 자르는 경로가 코드에 없음.
func openLedger(dir string) (*ledger, error) {
	if err := os.MkdirAll(filepath.Join(dir, statesDirName), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(dir, ledgerFileName), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}
	return &ledger{dir: dir, f: f}, nil
}

func (l *ledger) Close() error { return l.f.Close() }

func (l *ledger) appendRow(row any) error {
	line, err := json.Marshal(row)
	if err != nil {
		return err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, err := l.f.Write(append(line, '\n')); err != nil {
		return err
	}
	return l.f.Sync()
}

func (l *ledger) judgment(row judgmentRow) error {
	row.Schema, row.Kind = ledgerSchema, kindJudgment
	if row.CollectionGaps == nil {
		row.CollectionGaps = []string{}
	}
	return l.appendRow(row)
}

func (l *ledger) label(row labelRow) error {
	row.Schema, row.Kind = ledgerSchema, kindLabel
	if row.RealizedReturn == nil && row.LabelGap == "" {
		return errors.New("ledger: a label row needs a realized return or a label gap reason")
	}
	return l.appendRow(row)
}

func (l *ledger) gap(row gapRow) error {
	row.Schema, row.Kind = ledgerSchema, kindGap
	return l.appendRow(row)
}

// storeState 는 state 원문을 digest 이름 파일로 보관하고 digest 를 돌려줌.
// 같은 state 는 같은 파일 — 이미 있으면 다시 쓰지 않음.
func (l *ledger) storeState(state PublicState) (string, error) {
	raw, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	if err := assertPublicJSON(raw); err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	digest := "sha256:" + hex.EncodeToString(sum[:])
	path := filepath.Join(l.dir, statesDirName, hex.EncodeToString(sum[:])+".json")
	if _, err := os.Stat(path); err == nil {
		return digest, nil
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".state-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(append(raw, '\n')); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return "", err
	}
	return digest, nil
}

// ledgerContents 는 원장을 다시 읽은 결과(보고서·재시작 시 미라벨 복원용).
type ledgerContents struct {
	Judgments []judgmentRow
	Labels    map[string]labelRow
	Gaps      []gapRow
}

// readLedger 는 줄마다 kind 를 보고 나눔. 해석 못 하는 줄은 건너뛰지 않고 오류로 멈춤.
func readLedger(path string) (ledgerContents, error) {
	contents := ledgerContents{Labels: map[string]labelRow{}}
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return contents, nil
		}
		return contents, err
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	for lineNo := 1; ; lineNo++ {
		line, err := reader.ReadBytes('\n')
		if len(strings.TrimSpace(string(line))) > 0 {
			var head struct {
				Kind string `json:"kind"`
			}
			if jsonErr := json.Unmarshal(line, &head); jsonErr != nil {
				return contents, fmt.Errorf("%s line %d: %w", path, lineNo, jsonErr)
			}
			switch head.Kind {
			case kindJudgment:
				var row judgmentRow
				if jsonErr := json.Unmarshal(line, &row); jsonErr != nil {
					return contents, fmt.Errorf("%s line %d: %w", path, lineNo, jsonErr)
				}
				contents.Judgments = append(contents.Judgments, row)
			case kindLabel:
				var row labelRow
				if jsonErr := json.Unmarshal(line, &row); jsonErr != nil {
					return contents, fmt.Errorf("%s line %d: %w", path, lineNo, jsonErr)
				}
				contents.Labels[row.JudgmentID] = row
			case kindGap:
				var row gapRow
				if jsonErr := json.Unmarshal(line, &row); jsonErr != nil {
					return contents, fmt.Errorf("%s line %d: %w", path, lineNo, jsonErr)
				}
				contents.Gaps = append(contents.Gaps, row)
			default:
				return contents, fmt.Errorf("%s line %d: unknown row kind %q", path, lineNo, head.Kind)
			}
		}
		if errors.Is(err, io.EOF) {
			return contents, nil
		}
		if err != nil {
			return contents, err
		}
	}
}
