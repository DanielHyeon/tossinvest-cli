# Function Logic Map: `Context.ReconcileDriver`

- Source: `internal/app/engine/reconcileloop.go` (lines 348–374)
- AST evidence: `ast.json` (`source_sha256: 50a2c0f0b133fc0a6761fd8eee1950286c73fa6beca6d27da66dab16ecb42606`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **Normal** — 알림 문구와 표시 전용 배선. 주문·손절·사이징·원장 판정 경로를 바꾸지 않는다.

## What it does

조립 배선. a085는 알림 표면이 공유하는 InstrumentNames registry 하나를 만들어 넘긴다. 다른 배선은 그대로다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `Names` registry | nil 허용 | 대사 루프가 holdings 스냅샷에서 채운다 | nil이면 코드만 렌더 — a085 이전 동작 |
| 브로커 `name` | 빈 문자열 허용 | `GET /api/v1/holdings` 기존 필드 | 비면 코드만 렌더. 추정하지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (342) `if` — if c == nil | 본문 참조 | — | 아래 Branch Test Map |
| B2 | (345) `if` — if !c.Automation.Verified | 본문 참조 | — | 아래 Branch Test Map |
| B3 | (357) `if` — if opts.Prices == nil | 본문 참조 | — | 아래 Branch Test Map |
| B4 | (360) `if` — if opts.Alerts == nil && c.Notifier != nil | 본문 참조 | — | 아래 Branch Test Map |
| B5 | (363) `if` — if opts.Log == nil | 본문 참조 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `fmt.Errorf` | (353) return nil, fmt.Errorf("%w: the automation gate is not verified", ErrReconcileDriverUnavailable) | 호출부 계약 유지 | AST `calls` |
| `NewReconcileDriver` | (373) return NewReconcileDriver(opts) | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 알림 제목·본문 문자열만. 이벤트 타입·Key·Fields·발송 조건·latch·원장은 바뀌지 않는다.

## Safety conclusion

- Safe edit boundary: 사람이 읽는 문자열과 표시 전용 필드. 기계 판독 표면(payload·로그 키·원장 cause)은 영문·원문 그대로다.
- High-risk impact: no — 판정도 주문도 이 경로에 없다. §0.4는 불변이다: 이름은 이미 대금을 지불한 응답에서 온다.
