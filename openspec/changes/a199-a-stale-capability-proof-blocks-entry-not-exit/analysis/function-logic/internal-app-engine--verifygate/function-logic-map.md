# Function Logic Map: `verifyGate`

> **편집 후 지도(post-edit) — 2단계 GREEN 뒤, 2026-11-24, base `603046a23`.** 분기 표는 AST 열거(`../../harness/build_bundles.py --post`). 편집 전 원본 AST 는 `../../pre-edit-ast/internal-app-engine--verifygate.ast.json`.

- Source: `internal/app/engine/interlock.go` (519-610)
- AST evidence: `ast.json` — branches **12** / returns **12** / calls **24**, source SHA-256 `ed64648db576432a0100a9c8310b17ee18bd3f7f24067368acea4637759deb75`(파일 단위).
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
| B1 (if@521) | `if facts.guardian == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@529) | `if err := status.Limits.Validate(); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (if@536) | `if err := checkTradingPolicy(facts.trading); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B4 (if@542) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B5 (if@554) | `if accountRef == "" {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B6 (if@562) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B7 (if@568) | `if err := checkGuardianLimits(facts.guardian, status.Limits); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B8 (if@572) | `if err := checkGuardianDailyBuyLimit(facts.guardian, *status); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B9 (if@576) | `if err := checkGuardianOpenPositionLimit(facts.guardian, *status); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B10 (if@581) | `if err := checkGateway(facts.gateway); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B11 (if@584) | `if err := checkKeyCapableTransport(facts.transport); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B12 (if@603) | `if stale != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 24개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- `status.AttestationExpiresAt`(524) 한 칸만 씀.

## Safety conclusion

- a199 편집(적용됨): B6@536 의 낡음(만료 · endpoint 누락)을 반환하지 않고 표지로 남긴 채 B7~B11 을 끝까지 검사(D2), 만료일 때 계좌를 따로 확인(D1a). 첫 실패 반환 구조를 바꾸는 편집 — claim-check C2.
- Safe edit boundary: 편집 전 AST(`../../pre-edit-ast/`)와 비교해 위 편집 칸 밖의 분기는 무편집. 행동 고정은 a199 시험 · 변이 원장(`../../mutation/`).
- High-risk impact: yes — 기동 인터록 · 주문 권한 · 손절 경로 가용성(안전 부트스트랩 5)
