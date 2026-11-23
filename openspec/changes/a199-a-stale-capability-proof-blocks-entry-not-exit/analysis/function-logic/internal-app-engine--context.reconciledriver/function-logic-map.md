# Function Logic Map: `Context.ReconcileDriver`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-24, HEAD `e4165fa3d`(코드 = `9738a0c32`).** 분기 표는 AST 열거(`../../harness/build_bundles.py`), 2단계 GREEN 뒤 다시 만든다.

- Source: `internal/app/engine/reconcileloop.go` (364-395)
- AST evidence: `ast.json` — branches **6** / returns **4** / calls **4**, source SHA-256 `4e21e03de5619d5493799201640fb333e9c2b21f55255ecdfcdeafd14a19c7d3`(파일 단위).
  명령 `go run ./tools/logic-map --file internal/app/engine/reconcileloop.go --func Context.ReconcileDriver`.
- Risk scan: `risk-pattern-report.md`.
- 의도: 대사 + 편입 루프 조립 — 미검증 게이트면 거절.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.Automation.Verified` | NewContext | B2@368 | 아래 분기 표 |
| `opts.Prices` | 호출자 또는 c.Official | B5@385 poll 결속 | 아래 분기 표 |

## Branches and early returns

| Branch | Condition (소스 그 줄) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@365) | `if c == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@368) | `if !c.Automation.Verified {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (if@382) | `if opts.Prices == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B4 (if@385) | `if !opts.Retrier.EnginePollReaderAllowed(opts.Prices) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B5 (if@388) | `if opts.Alerts == nil && c.Notifier != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B6 (if@391) | `if opts.Log == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 4개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- opts 덮어쓰기.

## Safety conclusion

- a199 편집(예정): B2@368 를 청산 쪽 표지로(D3).
- Safe edit boundary: 위 편집 칸 밖의 분기는 무편집이 목표 — 2단계 post-edit 지도가 옛/새 AST 를 정렬해 확인.
- High-risk impact: yes — 기동 인터록 · 주문 권한 · 손절 경로 가용성(안전 부트스트랩 5)
