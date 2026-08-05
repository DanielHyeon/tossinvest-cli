# Function Logic Map: `TestAnExternalIncreaseAfterAdoptionIsReported`

- Source: `internal/app/engine/reconcileloop_test.go` (lines 806–846)
- AST evidence: `ast.json` (`source_sha256: 2456d7022ac6ce4a6de30e65d7c57d0b10b7a67ded5c6c404c38faa40e550106`)
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
| `newDriverHarness` | (807) h := newDriverHarness(t, nil) | 호출부 계약 유지 | AST `calls` |
| `h.holds` | (808) h.holds("005930", "10", "55000", 70000) | 호출부 계약 유지 | AST `calls` |
| `h.cycle` | (809) if cycle := h.cycle(); cycle.Adopted != 1 { | 호출부 계약 유지 | AST `calls` |
| `t.Fatalf` | (810) t.Fatalf("adopted = %d (%v)", cycle.Adopted, cycle.Err) | 호출부 계약 유지 | AST `calls` |
| `h.journal.ExitState` | (812) before, err := h.journal.ExitState(context.Background(), h.position("005930").ID) | 호출부 계약 유지 | AST `calls` |
| `context.Background` | (812) before, err := h.journal.ExitState(context.Background(), h.position("005930").ID) | 호출부 계약 유지 | AST `calls` |
| `h.position` | (812) before, err := h.journal.ExitState(context.Background(), h.position("005930").ID) | 호출부 계약 유지 | AST `calls` |
| `t.Fatal` | (814) t.Fatal(err) | 호출부 계약 유지 | AST `calls` |
| `t.Errorf` | (828) t.Errorf("t0 moved from %s/%s to %s/%s; the frozen denominator must not be recomputed", | 호출부 계약 유지 | AST `calls` |
| `strings.Contains` | (836) if e.Type == obs.EventExitPositionUnmanaged && strings.Contains(e.Key, "\|grown\|") { | 호출부 계약 유지 | AST `calls` |
| `eventTypes` | (844) t.Errorf("no alert about the external increase; events = %v", eventTypes(h.alerts.events)) | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 알림 제목·본문 문자열만. 이벤트 타입·Key·Fields·발송 조건·latch·원장은 바뀌지 않는다.

## Safety conclusion

- Safe edit boundary: 사람이 읽는 문자열과 표시 전용 필드. 기계 판독 표면(payload·로그 키·원장 cause)은 영문·원문 그대로다.
- High-risk impact: no — 판정도 주문도 이 경로에 없다. §0.4는 불변이다: 이름은 이미 대금을 지불한 응답에서 온다.
