package engine

// a146 — 진입 건너뜀 타입(strategyEntrySkip)의 언급 census(범위 거절 census 와 같은 규칙, a112 J4 ①).
//
// 전달 루프는 이 타입으로 「그 범위를 건너뛰고 **오류 없이** 다음으로 간다」를 고른다. 범위 거절보다 조용하다 — 주기 오류에도 남지 않는다.
// 그러므로 이 타입을 **만드는** 자리가 늘면 진짜 결함이 잠금 · 삼킴 없이 사라진다. 만드는 모양은 합성 리터럴만이 아니므로 식별자 언급 전부를
// 파일 · 함수 단위로 세고, 새 언급은 무엇이든 이 표를 바꿔야 한다(표를 바꾸는 편집은 리뷰에 보인다).

import (
	"go/ast"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

var strategyEntrySkipMentions = map[string]int{
	// 선언과 Error · Unwrap 메서드의 수신자.
	"a146_entry_skip.go:(decl)": 1,
	"a146_entry_skip.go:Error":  1,
	"a146_entry_skip.go:Unwrap": 1,
	// 기록자 운반(ctx 값의 함수 타입 · 꺼내기 · 감독자 기록) — 만들지 않음.
	"a146_entry_skip.go:withStrategyEntrySkipRecorder": 1,
	"a146_entry_skip.go:recordStrategyEntrySkip":       2,
	"a146_entry_skip.go:recordGuardianCapZeroSkip":     1,
	// **만드는 자리는 여기 하나뿐**: q_final 거절 타입 + 코드 EXISTING_GUARDIAN_CAP + 원인 sentinel risk.ErrStrategyQuantityZero(design D1).
	"strategy_first_leg_admission.go:admit": 1,
	// 운반: 전달 몸통의 errors.As 대상, 감독자 주기 ctx 의 기록자 클로저.
	"strategy_market_handoff_delivery.go:deliverEachStrategyHandoff": 1,
	"strategy_entry_supervisor.go:runMarket":                         1,
}

func TestTheEntrySkipTypeIsMadeOnlyWhereTheCensusSaysItIs(t *testing.T) {
	found := map[string]int{}
	for _, path := range engineProductionFiles(t) {
		file := parseEngineFile(t, path)
		base := filepath.Base(path)
		count := func(node ast.Node, label string) {
			ast.Inspect(node, func(inner ast.Node) bool {
				if ident, ok := inner.(*ast.Ident); ok && ident.Name == "strategyEntrySkip" {
					found[base+":"+label]++
				}
				return true
			})
		}
		for _, decl := range file.Decls {
			if function, ok := decl.(*ast.FuncDecl); ok {
				count(function, function.Name.Name)
				continue
			}
			count(decl, "(decl)")
		}
	}
	var drift []string
	for site, got := range found {
		if strategyEntrySkipMentions[site] != got {
			drift = append(drift, site)
		}
	}
	for site := range strategyEntrySkipMentions {
		if _, ok := found[site]; !ok {
			drift = append(drift, site)
		}
	}
	if len(drift) != 0 {
		sort.Strings(drift)
		t.Fatalf("entry-skip type mentions moved at %s: found %v, census %v — a new maker silences what the order path skips",
			strings.Join(drift, ", "), found, strategyEntrySkipMentions)
	}
}
