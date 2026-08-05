# Function Logic Map: `TestAnExternalIncreaseAfterAdoptionIsReported`

- Source: `internal/app/engine/reconcileloop_test.go` (lines 799–839)
- AST evidence: `ast.json` (`source_sha256: d53e58de88a1444c84d97750f68023fe8878894fc04eb5bc2026ade205132eae`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **Normal** — 테스트 코드.

## What it does

외부 증가 알림 테스트. a085는 제목의 영어 단어 'grew'로 찾던 것을 이벤트 Key(`|grown|`)로 바꿨다 — 제목은 운영자 산문이고, 산문을 고정한 테스트는 그 테스트가 다루는 것을 전혀 바꾸지 않은 변경에 실패한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `Names` registry | nil 허용 | 대사 루프가 holdings 스냅샷에서 채운다 | nil이면 코드만 렌더 — a085 이전 동작 |
| 브로커 `name` | 빈 문자열 허용 | `GET /api/v1/holdings` 기존 필드 | 비면 코드만 렌더. 추정하지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (802) `if` — if cycle := h.cycle(); cycle.Adopted != 1 | 본문 참조 | — | 아래 Branch Test Map |
| B2 | (806) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B3 | (812) `if` — if cycle := h.cycle(); cycle.Err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B4 | (817) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B5 | (820) `if` — if after.EntryPrice != before.EntryPrice || after.InitialRisk != before.Initia | 본문 참조 | — | 아래 Branch Test Map |
| B6 | (825) `range` — for _, e := range h.alerts.events | 본문 참조 | — | 아래 Branch Test Map |
| B7 | (829) `if` — if e.Type == obs.EventExitPositionUnmanaged && strings.Contains(e.Key, "|grown | 본문 참조 | — | 아래 Branch Test Map |
| B8 | (831) `if` — if e.Fields["adopted_quantity"] != "10" | 본문 참조 | — | 아래 Branch Test Map |
| B9 | (836) `if` — if !found | 본문 참조 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `{'kind': 'call', 'at': {'line': 800, 'column': 7}, 'text': 'newDriverHarness'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 801, 'column': 2}, 'text': 'h.holds'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 802, 'column': 14}, 'text': 'h.cycle'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 803, 'column': 3}, 'text': 't.Fatalf'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 805, 'column': 17}, 'text': 'h.journal.ExitState'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 805, 'column': 37}, 'text': 'context.Background'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 805, 'column': 59}, 'text': 'h.position'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 807, 'column': 3}, 'text': 't.Fatal'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 812, 'column': 14}, 'text': 'h.cycle'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 813, 'column': 3}, 'text': 't.Fatalf'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 816, 'column': 16}, 'text': 'h.journal.ExitState'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 816, 'column': 36}, 'text': 'context.Background'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 알림 제목·본문 문자열만. 이벤트 타입·Key·Fields·발송 조건·latch·원장은 바뀌지 않는다.

## Safety conclusion

- Safe edit boundary: 사람이 읽는 문자열과 표시 전용 필드. 기계 판독 표면(payload·로그 키·원장 cause)은 영문·원문 그대로다.
- High-risk impact: no — 판정도 주문도 이 경로에 없다. §0.4는 불변이다: 이름은 이미 대금을 지불한 응답에서 온다.
