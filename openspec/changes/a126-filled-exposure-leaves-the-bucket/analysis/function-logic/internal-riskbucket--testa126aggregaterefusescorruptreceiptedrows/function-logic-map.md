# Function Logic Map: `TestA126AggregateRefusesCorruptReceiptedRows`

- Source: `internal/riskbucket/a126_departed_rows_test.go` (`62`–`85`)
- Qualified: `TestA126AggregateRefusesCorruptReceiptedRows`
- AST evidence: `ast.json` (`source_sha256` 382cb136b66deb63…, revision current) — **편집 뒤**(시험 전용)
- Risk scan: `risk-pattern-report.md`
- AST branches 3 · return 7 · 호출 6

**역할.** 행 규칙 손상 거절 단위 시험. 비례 원칙(WORKFLOW 「비례 원칙」): 시험 전용 — 무거운 규율 대상 아님. base 재고정(review 2.1) 뒤 게이트가 수정 함수로 세므로 경량 번들을 둔다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 시험 임시 원장 · fixture | 시험 | 생산 작성자 경로 + 손상만 SQL | `t.Fatal`/`t.Errorf` |

## Branches and early returns

> AST 분기 열거(시험 코드 — 커버리지 블록 없음).

| Branch | 종류 | 조건 (원문) |
|---|---|---|
| B1 | range | `:76` `for name, corrupt := range cases {` |
| B2 | range | `:77` `for _, latched := range []int{0, 1} {` |
| B3 | if | `:80` `if _, err := aggregateProductionRiskUsage([]productionRiskUsageRow{row}); !errors.Is(err, ErrJournalUsageIn…` |

## Calls and live bindings

시험 fixture · 생산 journal/riskbucket API 호출. 브로커 · 실계좌 호출 없음.

## State mutations and fallbacks

시험 임시 원장만.

## Safety conclusion

- **Safe edit boundary**: 사례 하나 추가(HELD · held 0 — R4, 변이 X3).
- **High-risk impact**: no — 시험. 생산 동작 변화 0(`64ad3058` 비시험 `.go` 변경 0 — codex 재리뷰 확인).
