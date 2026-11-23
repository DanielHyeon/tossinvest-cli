# Function Logic Map: `NewContext`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-24, HEAD `e4165fa3d`(코드 = `9738a0c32`).** 분기 표는 AST 열거(`../../harness/build_bundles.py`), 2단계 GREEN 뒤 다시 만든다.

- Source: `internal/app/engine/engine.go` (451-631)
- AST evidence: `ast.json` — branches **17** / returns **12** / calls **38**, source SHA-256 `b6347e9a3cf527dd06dc2bb8c0b2dec3d799c29648849c4b187c779c253a2fca`(파일 단위).
  명령 `go run ./tools/logic-map --file internal/app/engine/engine.go --func NewContext`.
- Risk scan: `risk-pattern-report.md`.
- 의도: 엔진 프로필 조립: 계좌 → journal → 게이트웨이 → Guardian → 인터록 → Context.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `opts` | assembleEngine · 시험 | Options | 아래 분기 표 |
| 불변식 | 인터록이 거절하면 Context 없음(B15@588) | engine-safety 「엔진 프로필은 다음 순서로 구성한다」 | 아래 분기 표 |
| a166 census | 본문 전체가 `shapes.json` 키로 동결 | internal/accountbudgetruntime/testdata/a166-wiring | 아래 분기 표 |

## Branches and early returns

| Branch | Condition (소스 그 줄) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@453) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@460) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (if@464) | `if clk == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B4 (if@478) | `if opts.Publisher != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B5 (if@486) | `if err := recordGateSettings(auditLog, gate, cfg.Engine.Adoption, notifications,` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B6 (if@501) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B7 (if@508) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B8 (if@517) | `if err := bindApplyHooks(jrn, tradeOutcomeRefusalSink(opts.Logger)); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B9 (if@524) | `if opts.BudgetFactory != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B10 (if@526) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B11 (if@545) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B12 (if@560) | `if gate.Enabled && guardian == nil && !opts.disableProductionGuardian {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B13 (if@562) | `if factory == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B14 (if@568) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B15 (if@588) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B16 (if@592) | `if !automation.Verified {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B17 (if@597) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 38개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- journal 열기 · 감사 · 게이트웨이 조립(기동 복원 래치) · Guardian 생성.

## Safety conclusion

- a199 편집(예정): B16@592 Guardian 유지 조건을 청산 쪽 표지로(D3). 래치를 runInterlock 에 두면 `gateFacts` 에 진입 게이트 한 칸. census 키 1 갱신.
- Safe edit boundary: 위 편집 칸 밖의 분기는 무편집이 목표 — 2단계 post-edit 지도가 옛/새 AST 를 정렬해 확인.
- High-risk impact: yes — 기동 인터록 · 주문 권한 · 손절 경로 가용성(안전 부트스트랩 5)
