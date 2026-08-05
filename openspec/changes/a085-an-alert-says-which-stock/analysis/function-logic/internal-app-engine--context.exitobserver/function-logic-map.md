# Function Logic Map: `Context.ExitObserver`

- Source: `internal/app/engine/exitwiring.go` (lines 319–348)
- AST evidence: `ast.json` (`source_sha256: 43dcca8f37b748962f5e066706a887c40f694d189fc81f7b7697949e084cb020`)
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
| B1 | (320) `if` — if c == nil | 본문 참조 | — | 아래 Branch Test Map |
| B2 | (323) `if` — if !c.Automation.Verified | 본문 참조 | — | 아래 Branch Test Map |
| B3 | (327) `if` — if !ok | 본문 참조 | — | 아래 Branch Test Map |
| B4 | (338) `if` — if opts.Names == nil | 본문 참조 | — | 아래 Branch Test Map |
| B5 | (341) `if` — if opts.Alerts == nil && c.Notifier != nil | 본문 참조 | — | 아래 Branch Test Map |
| B6 | (344) `if` — if opts.Floor == nil | 본문 참조 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `{'kind': 'call', 'at': {'line': 324, 'column': 15}, 'text': 'fmt.Errorf'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 328, 'column': 15}, 'text': 'fmt.Errorf'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 347, 'column': 9}, 'text': 'NewExitObserver'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 알림 제목·본문 문자열만. 이벤트 타입·Key·Fields·발송 조건·latch·원장은 바뀌지 않는다.

## Safety conclusion

- Safe edit boundary: 사람이 읽는 문자열과 표시 전용 필드. 기계 판독 표면(payload·로그 키·원장 cause)은 영문·원문 그대로다.
- High-risk impact: no — 판정도 주문도 이 경로에 없다. §0.4는 불변이다: 이름은 이미 대금을 지불한 응답에서 온다.
