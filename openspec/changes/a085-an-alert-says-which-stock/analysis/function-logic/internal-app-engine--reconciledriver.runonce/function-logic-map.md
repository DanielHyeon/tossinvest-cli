# Function Logic Map: `ReconcileDriver.RunOnce`

- Source: `internal/app/engine/reconcileloop.go` (lines 397–457)
- AST evidence: `ast.json` (`source_sha256: 821dc7a7c1b58e8756f7cc423f527d01c76593a623c2b86db0e3fc08a8a01364`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **Normal** — 알림 문구와 표시 전용 배선. 주문·손절·사이징·원장 판정 경로를 바꾸지 않는다.

## What it does

대사 한 주기. a085는 안정화된 스냅샷의 holdings에서 심볼→이름을 학습하는 루프 하나를 더한다. 추가 요청 없음(§0.4) — 같은 응답의 기존 필드다. 비교·수렴·관측·편입 판정은 그대로다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `Names` registry | nil 허용 | 대사 루프가 holdings 스냅샷에서 채운다 | nil이면 코드만 렌더 — a085 이전 동작 |
| 브로커 `name` | 빈 문자열 허용 | `GET /api/v1/holdings` 기존 필드 | 비면 코드만 렌더. 추정하지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (401) `if` — if !ok | 본문 참조 | — | 아래 Branch Test Map |
| B2 | (409) `if` — if d.opts.Names != nil | 본문 참조 | — | 아래 Branch Test Map |
| B3 | (410) `range` — for _, holding := range snapshot.Holdings | 본문 참조 | — | 아래 Branch Test Map |
| B4 | (416) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B5 | (427) `if` — if err != nil && cycle.Err == nil | 본문 참조 | — | 아래 Branch Test Map |
| B6 | (433) `if` — if err != nil && cycle.Err == nil | 본문 참조 | — | 아래 Branch Test Map |
| B7 | (439) `if` — if err := d.opts.Tracker.Refresh(ctx); err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B8 | (440) `if` — if cycle.Err == nil | 본문 참조 | — | 아래 Branch Test Map |
| B9 | (448) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B10 | (449) `if` — if cycle.Err == nil | 본문 참조 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `{'kind': 'call', 'at': {'line': 398, 'column': 17}, 'text': 'd.note'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 398, 'column': 8}}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 400, 'column': 18}, 'text': 'd.stabilise'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 411, 'column': 4}, 'text': 'd.opts.Names.Learn'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 415, 'column': 16}, 'text': 'reconcile.LocalStateFromJournal'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 417, 'column': 15}, 'text': 'fmt.Errorf'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 420, 'column': 10}, 'text': 'Compare'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 426, 'column': 17}, 'text': 'd.ingest.IngestExternalPositions'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 430, 'column': 17}, 'text': 'len'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 432, 'column': 20}, 'text': 'd.opts.Converge.ConvergeQuantities'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 436, 'column': 20}, 'text': 'len'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 439, 'column': 12}, 'text': 'd.opts.Tracker.Refresh'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 알림 제목·본문 문자열만. 이벤트 타입·Key·Fields·발송 조건·latch·원장은 바뀌지 않는다.

## Safety conclusion

- Safe edit boundary: 사람이 읽는 문자열과 표시 전용 필드. 기계 판독 표면(payload·로그 키·원장 cause)은 영문·원문 그대로다.
- High-risk impact: no — 판정도 주문도 이 경로에 없다. §0.4는 불변이다: 이름은 이미 대금을 지불한 응답에서 온다.
