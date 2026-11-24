package reconcile

// a201_fill_credit.go 는 「체결이 조정을 대신했다」는 credit(a201 design v3)을 담음.
//
// 대사의 로컬 읽기 뒤 · 수렴의 조정 tx 전에 그 종목의 체결이 커밋되면 수렴은 그 종목을 조정하지 못하고(거절 · stale ·
// 판 중단) 차단만 섬 — 자동 해제 근거(조정 credit)가 없어 고착함. 그 주기의 Observe 는 지금과 똑같이 돌게 두고, 끝난 뒤
// 그 종목에 조정 credit 과 같은 종류의 credit 을 줌. credit 은 Tracker.adjusted 에 같은 값(비교 as-of)으로 들어가므로
// 쓰기 · 반박 · 소멸은 조정 credit 의 기존 분기가 그대로 맡음(새 판정 0). 해제 원인 이름만 곁 맵 filled 로 가름.

import (
	"context"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// FillMark 는 로컬 읽기 직전에 읽은 전역 최대 체결 이벤트 id 임. known 이 거짓이면 「모름」 — 그 주기는 체결 credit 이 없음.
type FillMark struct {
	id    int64
	known bool
}

// MarkBeforeLocalRead 는 로컬 읽기 **전에** 전역 최대 체결 id 를 읽음(V2 (b)).
//
// 뒤에 읽으면 그 사이 커밋된 체결을 놓침 — id 순서가 커밋 순서(단일 쓰기 연결)이므로 이 표식보다 큰 종목 워터마크는
// 로컬 읽기와 겹치거나 그 뒤에 커밋된 체결임. 읽기 실패는 주기를 실패시키지 않고 「모름」으로 돌려줌(로컬 읽기 오류는
// 바로 뒤의 LocalStateFromJournal 이 따로 처분함).
func MarkBeforeLocalRead(ctx context.Context, j *journal.Journal) FillMark {
	if j == nil {
		return FillMark{}
	}
	id, err := j.MaxFillEventID(ctx)
	if err != nil {
		return FillMark{}
	}
	return FillMark{id: id, known: true}
}

// FillOutranSymbols 는 체결 credit 을 줄 종목을 셈 — V2 의 닫힌 조건 (a)(b)(c), (d)는 부여 메서드가 봄.
//
//	(a) 그 비교의 수량 불일치 종목(빈 종목 · missing-order 만 · 외부 포지션만은 아님)
//	(b) 그 종목 워터마크가 로컬 읽기 전 표식보다 큼
//	(c) 이번 수렴이 그 종목을 조정하지 않음(거절 · stale · 판 중단 모두 — 조정이 있었으면 조정 credit 이 이미 있음)
//
// 표식이 「모름」이거나 종목 워터마크를 하나라도 못 읽으면 아무 종목도 주지 않음 — credit 은 해제 쪽 근거라 모르면
// 안 주는 쪽이 보수임.
func FillOutranSymbols(ctx context.Context, j *journal.Journal, mark FillMark, diff Diff, converged ConvergeReport) []string {
	if j == nil || !mark.known {
		return nil
	}
	adjusted := make(map[string]bool, len(converged.Converged))
	for _, c := range converged.Converged {
		adjusted[strings.ToUpper(strings.TrimSpace(c.Symbol))] = true
	}
	var out []string
	for _, mismatch := range diff.Quantities {
		symbol := strings.ToUpper(strings.TrimSpace(mismatch.Symbol))
		if symbol == "" || adjusted[symbol] {
			continue
		}
		watermark, err := j.FillWatermark(ctx, symbol)
		if err != nil {
			return nil
		}
		if watermark > mark.id {
			out = append(out, symbol)
		}
	}
	return out
}

// FillsOutranAdjustment 는 그 종목들에 체결 credit 을 줌(V3 — RunOnce 가 Observe 성공 뒤에 부름).
//
// credit 은 조정 credit 과 같은 맵 · 같은 값(그 주기 비교 as-of)이라 그 주기에는 쓰이지도 반박되지도 않고(엄격히 나중),
// 그 뒤 비교에서 조정 credit 의 규칙 그대로 쓰이거나 반박 · 소멸함. 그 종목에 자동 해제 대상 수량 불일치 블록이 없으면
// 주지 않음(V2 (d) — credit 의 수명 규칙과 같은 조건). 곁 맵은 지금 credit 과 값이 다른 항목을 먼저 정리해 낡은 표지가
// 남지 않게 함(V5b).
func (t *Tracker) FillsOutranAdjustment(comparison string, symbols ...string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for symbol, stamp := range t.filled {
		if t.adjusted[symbol] != stamp {
			delete(t.filled, symbol)
		}
	}
	stamp := strings.TrimSpace(comparison)
	for _, s := range symbols {
		symbol := strings.ToUpper(strings.TrimSpace(s))
		if symbol == "" || !hasBlockFor(t.blocks, symbol) {
			continue
		}
		if t.adjusted == nil {
			t.adjusted = map[string]string{}
		}
		if t.filled == nil {
			t.filled = map[string]string{}
		}
		t.adjusted[symbol] = stamp
		t.filled[symbol] = stamp
	}
}
