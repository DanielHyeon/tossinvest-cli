# Function Logic Map: `Context.ExitObserver`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-24, HEAD `e4165fa3d`(코드 = `9738a0c32`).** 분기 표는 AST 열거(`../../harness/build_bundles.py`), 2단계 GREEN 뒤 다시 만든다.

- Source: `internal/app/engine/exitwiring.go` (319-365)
- AST evidence: `ast.json` — branches **6** / returns **4** / calls **6**, source SHA-256 `e8f131a312ece6fbaadecd1c71131d7db892c9ab64ac2f9254d56ec1936c8c30`(파일 단위).
  명령 `go run ./tools/logic-map --file internal/app/engine/exitwiring.go --func Context.ExitObserver`.
- Risk scan: `risk-pattern-report.md`.
- 의도: exit 관측 루프 조립 — 미검증 게이트면 거절, Guardian 이 `ReductionIssuer` 여야 함.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.Automation.Verified` | NewContext | B2@323 | 아래 분기 표 |
| `c.Guardian` | NewContext B16 | B3@327 | 아래 분기 표 |

## Branches and early returns

| Branch | Condition (소스 그 줄) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@320) | `if c == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@323) | `if !c.Automation.Verified {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (if@327) | `if !ok {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B4 (if@344) | `if opts.Names == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B5 (if@349) | `if c.Notifier != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B6 (if@361) | `if opts.Floor == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 6개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- opts 덮어쓰기(Journal · Prices · Retrier · Issuer · Submit …).

## Safety conclusion

- a199 편집(예정): B2@323 를 청산 쪽 표지로(D3). a198 이 같은 함수를 편집(Guardian 을 CancelIssuer 로 단언) — a198 착지 뒤 순차.
- Safe edit boundary: 위 편집 칸 밖의 분기는 무편집이 목표 — 2단계 post-edit 지도가 옛/새 AST 를 정렬해 확인.
- High-risk impact: yes — 기동 인터록 · 주문 권한 · 손절 경로 가용성(안전 부트스트랩 5)
