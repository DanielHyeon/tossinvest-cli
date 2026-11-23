# Function Logic Map: `verifyGate`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-24, HEAD `e4165fa3d`(코드 = `9738a0c32`).** 분기 표는 AST 열거(`../../harness/build_bundles.py`), 2단계 GREEN 뒤 다시 만든다.

- Source: `internal/app/engine/interlock.go` (496-578)
- AST evidence: `ast.json` — branches **11** / returns **12** / calls **22**, source SHA-256 `d018e5cf62544eef99ad68fadebdbe0de63f4677d0bea51d39a33ccc18f037e1`(파일 단위).
  명령 `go run ./tools/logic-map --file internal/app/engine/interlock.go --func verifyGate`.
- Risk scan: `risk-pattern-report.md`.
- 의도: 게이트 ON 기동의 조항 1~8 검사. 첫 실패에서 반환(AST 11 갈래가 전부 오류 반환).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `status.Limits`·`AttestationFile`·`AccountRef` | newAutomationStatus + NewContext 가 채움 | NewContext | 아래 분기 표 |
| `facts`(guardian · trading · gateway · transport · now) | NewContext 가 조립 | gateFacts | 아래 분기 표 |
| 불변식 | 하나라도 실패하면 `ErrAutomationGateRefused` 로 감싼 오류(조항 sentinel `%w`) | engine-safety 「자동화 게이트 기동 인터록」 | 아래 분기 표 |

## Branches and early returns

| Branch | Condition (소스 그 줄) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@498) | `if facts.guardian == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@506) | `if err := status.Limits.Validate(); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (if@513) | `if err := checkTradingPolicy(facts.trading); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B4 (if@519) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B5 (if@531) | `if accountRef == "" {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B6 (if@536) | `if err := att.Verify(facts.now, accountRef, RequiredEndpoints()); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B7 (if@542) | `if err := checkGuardianLimits(facts.guardian, status.Limits); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B8 (if@546) | `if err := checkGuardianDailyBuyLimit(facts.guardian, *status); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B9 (if@550) | `if err := checkGuardianOpenPositionLimit(facts.guardian, *status); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B10 (if@555) | `if err := checkGateway(facts.gateway); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B11 (if@558) | `if err := checkKeyCapableTransport(facts.transport); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 22개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- `status.AttestationExpiresAt`(524) 한 칸만 씀.

## Safety conclusion

- a199 편집(예정): B6@536 의 낡음(만료 · endpoint 누락)을 반환하지 않고 표지로 남긴 채 B7~B11 을 끝까지 검사(D2), 만료일 때 계좌를 따로 확인(D1a). 첫 실패 반환 구조를 바꾸는 편집 — claim-check C2.
- Safe edit boundary: 위 편집 칸 밖의 분기는 무편집이 목표 — 2단계 post-edit 지도가 옛/새 AST 를 정렬해 확인.
- High-risk impact: yes — 기동 인터록 · 주문 권한 · 손절 경로 가용성(안전 부트스트랩 5)
