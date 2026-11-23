# Function Logic Map: `TestGateOnRefusals`

> **편집 후 지도(post-edit) — 2단계 GREEN 뒤, 2026-11-24, base `603046a23`.** 분기 표는 AST 열거(`../../harness/build_bundles.py --post`). 편집 전 원본 AST 는 `../../pre-edit-ast/internal-app-engine--testgateonrefusals.ast.json`.

- Source: `internal/app/engine/interlock_test.go` (342-573)
- AST evidence: `ast.json` — branches **10** / returns **1** / calls **54**, source SHA-256 `d2c0c7aec58bfbb4138e89e776ec80beba0738d2e781dbcdcdc7e23ba2ee0fdb`(파일 단위).
  명령 `go run ./tools/logic-map --file internal/app/engine/interlock_test.go --func TestGateOnRefusals`.
- Risk scan: `risk-pattern-report.md`.
- 의도: 게이트 ON 거절 사례 표(조항별 sentinel · 문구).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 사례 표 | 시험 안 | 342-556 | 아래 분기 표 |

## Branches and early returns

| Branch | Condition (소스 그 줄) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (range@535) | `for _, tc := range cases {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@539) | `if tc.trading != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (if@544) | `if tc.writeAtt {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B4 (if@551) | `if tc.name == "no Guardian injected" {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B5 (else@553) | `} else {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B6 (if@556) | `if err == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B7 (if@559) | `if eng != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B8 (if@562) | `if !errors.Is(err, engine.ErrAutomationGateRefused) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B9 (if@565) | `if !errors.Is(err, tc.wantErr) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B10 (if@568) | `if tc.wantMessage != "" && !strings.Contains(err.Error(), tc.wantMessage) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 54개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 임시 디렉터리 · httptest.

## Safety conclusion

- a199 편집(적용됨): 낡음 사례 셋(「attestation expired」·「misses an endpoint the engine calls」·「predates the price read」)이 거절 기대에서 청산 전용 기대로.
- Safe edit boundary: 편집 전 AST(`../../pre-edit-ast/`)와 비교해 위 편집 칸 밖의 분기는 무편집. 행동 고정은 a199 시험 · 변이 원장(`../../mutation/`).
- High-risk impact: no(시험 함수)
