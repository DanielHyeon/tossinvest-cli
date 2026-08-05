# Function Logic Map: `ReconcileDriver.stabilise`

- Source: `internal/app/engine/reconcileloop.go` (lines 485–523)
- AST evidence: `ast.json` (`source_sha256: 50a2c0f0b133fc0a6761fd8eee1950286c73fa6beca6d27da66dab16ecb42606`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **Normal**

## What it does

안정화 간격을 두고 두 번 수집해 같은 스냅샷이면 채택한다. a085가 더한 것은
`d.learnNames(first)` 한 줄이다 — 개정 2 이전에는 두 번째(안정화된) 스냅샷에서만 이름을
배웠고, 첫 수집만 성공하고 두 번째가 실패한 사이클은 이미 대금을 지불한 응답의 이름을 버렸다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `d.opts.StabilisationInterval` | > 0, 아니면 기본값 | 런타임 config | <= 0이면 `reconcile.DefaultStabilisationInterval` |
| `Collector.Collect` 결과 | 계좌 스냅샷 | 공식 Open API 조회 | 오류면 `cycle.Err`에 담고 false — 이번 사이클 종료 |
| `d.stabiliser` | 연속 동일 스냅샷 누적기 | 드라이버 소유 | 불안정이면 streak을 유지한 채 false |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Required test |
|---|---|---|---|
| B1 | (487) `if` — if interval <= 0 { | 본문 참조 | 아래 Branch Test Map |
| B2 | (492) `if` — if err != nil { | 본문 참조 | 아래 Branch Test Map |
| B3 | (500) `if` — if err := d.clk.Sleep(ctx, interval); err != nil { | 본문 참조 | 아래 Branch Test Map |
| B4 | (506) `if` — if err != nil { | 본문 참조 | 아래 Branch Test Map |
| B5 | (513) `if` — if !result.Stable { | 본문 참조 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `d.opts.Collector.Collect` | 계좌 스냅샷 수집 | 오류를 `cycle.Err`로 옮기고 즉시 종료 | AST `calls` L491·L505 |
| `d.learnNames` | 이미 받은 응답에서 종목명을 기억 | 무오류. 빈 이름은 기존 이름을 지우지 않는다 | AST `calls` L497·L511 |
| `d.clk.Sleep` | 안정화 간격 | ctx 취소면 종료 | AST `calls` L500 |
| `d.stabiliser.Offer` / `Reset` | 연속 동일 판정 | 순수 상태 누적 | AST `calls` L498·L512·L521 |

## State mutations and fallbacks

- `cycle.Collected++` (수집 2회), `cycle.Err`.
- 이름 registry — 표시 전용. 판정·주문·원장에 들어가지 않는다.
- `d.stabiliser` streak — 불안정 시 의도적으로 유지한다.

## Safety conclusion

- Safe edit boundary: 이름 학습 호출 지점. 수집 횟수·간격·안정화 판정은 그대로다.
- High-risk impact: no — §0.4 불변. 이름은 이미 지불한 응답에서 오며 추가 요청이 없다.
