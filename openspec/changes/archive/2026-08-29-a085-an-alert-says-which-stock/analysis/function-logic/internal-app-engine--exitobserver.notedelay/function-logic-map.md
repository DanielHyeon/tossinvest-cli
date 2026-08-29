# Function Logic Map: `ExitObserver.noteDelay`

- Source: `internal/app/engine/exitloop.go` (lines 1569–1593)
- AST evidence: `ast.json` (`source_sha256: 6625c92061d5b05f566ecb0913f5c5f74a7fdde4cc4b5d8e7dfe8e75dd71de00`)
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
| B1 | (1483) `if` — if !running | 본문 참조 | — | 아래 Branch Test Map |
| B2 | (1487) `if` — if now.Sub(since) < o.delayBound() || o.delayAlerted[m.position.ID] | 본문 참조 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `o.clk.Now` | (1570) now := o.clk.Now() | 호출부 계약 유지 | AST `calls` |
| `now.Sub` | (1576) if now.Sub(since) < o.delayBound() \|\| o.delayAlerted[m.position.ID] { | 호출부 계약 유지 | AST `calls` |
| `o.delayBound` | (1576) if now.Sub(since) < o.delayBound() \|\| o.delayAlerted[m.position.ID] { | 호출부 계약 유지 | AST `calls` |
| `o.alert` | (1580) o.alert(ctx, obs.Event{ | 호출부 계약 유지 | AST `calls` |
| `o.label` | (1583) Title: o.label(m.position.Symbol) + " 청산이 허용 지연을 넘겼다", | 호출부 계약 유지 | AST `calls` |
| `fmt.Sprintf` | (1584) Body: fmt.Sprintf("%s · 경과 %s. 브로커에 상주하는 손절이 생기기 전까지 이 지연은 "+ | 호출부 계약 유지 | AST `calls` |
| `Round` | (1585) "보호되지 않는 노출이다.", why, now.Sub(since).Round(time.Second)), | 호출부 계약 유지 | AST `calls` |
| `int` | (1589) "delay_seconds": int(now.Sub(since).Seconds()), | 호출부 계약 유지 | AST `calls` |
| `Seconds` | (1589) "delay_seconds": int(now.Sub(since).Seconds()), | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 알림 제목·본문 문자열만. 이벤트 타입·Key·Fields·발송 조건·latch·원장은 바뀌지 않는다.

## Safety conclusion

- Safe edit boundary: 사람이 읽는 문자열과 표시 전용 필드. 기계 판독 표면(payload·로그 키·원장 cause)은 영문·원문 그대로다.
- High-risk impact: no — 판정도 주문도 이 경로에 없다. §0.4는 불변이다: 이름은 이미 대금을 지불한 응답에서 온다.
