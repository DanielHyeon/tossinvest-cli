# Function Logic Map: `ReconcileDriver.adoptOne`

- Source: `internal/app/engine/adoption.go` (lines 318–380)
- AST evidence: `ast.json` (`source_sha256: f121aba90cd05c31de5eeb9acad14497db9150c61314039419cbd5d9e0734edd`)
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
| B1 | (320) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B2 | (339) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B3 | (344) `if` — if _, err := d.opts.Journal.OpenAdoptedExitState(ctx, c.position.ID); err != n | 본문 참조 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `exitpolicy.SyntheticStop` | (319) stop, err := exitpolicy.SyntheticStop(observed, d.opts.Adoption.DefaultStopPct) | 호출부 계약 유지 | AST `calls` |
| `d.logDeferred` | (321) d.logDeferred(c.position, "the synthetic stop could not be derived: "+err.Error()) | 호출부 계약 유지 | AST `calls` |
| `err.Error` | (321) d.logDeferred(c.position, "the synthetic stop could not be derived: "+err.Error()) | 호출부 계약 유지 | AST `calls` |
| `d.opts.Journal.AdoptPosition` | (325) adoption, err := d.opts.Journal.AdoptPosition(ctx, journal.AdoptionRequest{ | 호출부 계약 유지 | AST `calls` |
| `journal.RFC3339` | (336) ObservedAt:    journal.RFC3339(d.clk.Now()), | 호출부 계약 유지 | AST `calls` |
| `d.clk.Now` | (336) ObservedAt:    journal.RFC3339(d.clk.Now()), | 호출부 계약 유지 | AST `calls` |
| `d.opts.Journal.OpenAdoptedExitState` | (344) if _, err := d.opts.Journal.OpenAdoptedExitState(ctx, c.position.ID); err != nil { | 호출부 계약 유지 | AST `calls` |
| `d.alert` | (353) d.alert(ctx, obs.Event{ | 호출부 계약 유지 | AST `calls` |
| `d.label` | (356) Title: d.label(c.position.Symbol) + " 엔진 관리로 편입됐다", | 호출부 계약 유지 | AST `calls` |
| `delete` | (378) delete(d.unmanaged, c.position.ID) | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 알림 제목·본문 문자열만. 이벤트 타입·Key·Fields·발송 조건·latch·원장은 바뀌지 않는다.

## Safety conclusion

- Safe edit boundary: 사람이 읽는 문자열과 표시 전용 필드. 기계 판독 표면(payload·로그 키·원장 cause)은 영문·원문 그대로다.
- High-risk impact: no — 판정도 주문도 이 경로에 없다. §0.4는 불변이다: 이름은 이미 대금을 지불한 응답에서 온다.
