# Function Logic Map: `ExitObserver.checkOutage`

- Source: `internal/app/engine/exitloop.go` (lines 746–783)
- AST evidence: `ast.json` (`source_sha256: b8788903ccb57d2f90722fe492290688f00531b34041a980a1c3f9033317b2b1`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **Normal** — 알림 문구와 표시 전용 배선. 주문·손절·사이징·원장 판정 경로를 바꾸지 않는다.

## What it does

알림 조립부. a085는 제목·본문을 한국어로 바꾸고 종목을 `이름(코드)`로 부른다. 이벤트 타입·Key·Fields(기계 판독 표면)와 발송 조건·latch는 그대로다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `Names` registry | nil 허용 | 대사 루프가 holdings 스냅샷에서 채운다 | nil이면 코드만 렌더 — a085 이전 동작 |
| 브로커 `name` | 빈 문자열 허용 | `GET /api/v1/holdings` 기존 필드 | 비면 코드만 렌더. 추정하지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (748) `if` — if since.IsZero() | 본문 참조 | — | 아래 Branch Test Map |
| B2 | (751) `if` — if o.clk.Now().Sub(since) < o.outageAfter() | 본문 참조 | — | 아래 Branch Test Map |
| B3 | (754) `if` — if o.outageRaised | 본문 참조 | — | 아래 Branch Test Map |
| B4 | (772) `if` — if o.opts.Escalate == nil || strings.TrimSpace(o.opts.AccountRef) == "" | 본문 참조 | — | 아래 Branch Test Map |
| B5 | (777) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `{'kind': 'call', 'at': {'line': 748, 'column': 5}, 'text': 'since.IsZero'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 751, 'column': 30}, 'text': 'o.outageAfter'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 751, 'column': 5}, 'text': 'Sub'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 751, 'column': 5}, 'text': 'o.clk.Now'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 759, 'column': 2}, 'text': 'o.alert'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 761, 'column': 10}, 'text': 'string'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 762, 'column': 32}, 'text': 'Round'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 762, 'column': 32}, 'text': 'String'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 762, 'column': 32}, 'text': 'Sub'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 762, 'column': 32}, 'text': 'o.clk.Now'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 763, 'column': 9}, 'text': 'Format'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 763, 'column': 9}, 'text': 'since.UTC'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 알림 제목·본문 문자열만. 이벤트 타입·Key·Fields·발송 조건·latch·원장은 바뀌지 않는다.

## Safety conclusion

- Safe edit boundary: 사람이 읽는 문자열과 표시 전용 필드. 기계 판독 표면(payload·로그 키·원장 cause)은 영문·원문 그대로다.
- High-risk impact: no — 판정도 주문도 이 경로에 없다. §0.4는 불변이다: 이름은 이미 대금을 지불한 응답에서 온다.
