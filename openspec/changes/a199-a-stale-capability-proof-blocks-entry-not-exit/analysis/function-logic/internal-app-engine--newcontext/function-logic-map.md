# Function Logic Map: `NewContext`

> **편집 후 지도(post-edit) — 2단계 GREEN 뒤, 2026-11-24, base `603046a23`.** 분기 표는 AST 열거(`../../harness/build_bundles.py --post`). 편집 전 원본 AST 는 `../../pre-edit-ast/internal-app-engine--newcontext.ast.json`.

- Source: `internal/app/engine/engine.go` (451-632)
- AST evidence: `ast.json` — branches **17** / returns **12** / calls **39**, source SHA-256 `6db4b2eb9434d9f52d31b9ac03c80e5484e4e24fafaea178a75d2c418d17178b`(파일 단위).
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
| B15 (if@589) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B16 (if@593) | `if !automation.RunsExitSide() { // a199: 청산 전용 기동도 Guardian 을 둠 — exit 관측이 그 위험 감소 발급을 씀` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B17 (if@598) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 39개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- journal 열기 · 감사 · 게이트웨이 조립(기동 복원 래치) · Guardian 생성.

## Safety conclusion

- a199 편집(적용됨): B16@592 Guardian 유지 조건을 청산 쪽 표지로(D3). 래치를 runInterlock 에 두면 `gateFacts` 에 진입 게이트 한 칸. census 키 1 갱신.
- Safe edit boundary: 편집 전 AST(`../../pre-edit-ast/`)와 비교해 위 편집 칸 밖의 분기는 무편집. 행동 고정은 a199 시험 · 변이 원장(`../../mutation/`).
- High-risk impact: yes — 기동 인터록 · 주문 권한 · 손절 경로 가용성(안전 부트스트랩 5)
