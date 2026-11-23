# Function Logic Map: `TestGateOnRefusals`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-24, HEAD `e4165fa3d`(코드 = `9738a0c32`).** 분기 표는 AST 열거(`../../harness/build_bundles.py`), 2단계 GREEN 뒤 다시 만든다.

- Source: `internal/app/engine/interlock_test.go` (342-596)
- AST evidence: `ast.json` — branches **10** / returns **1** / calls **59**, source SHA-256 `23dc2525546a616531f0e55d28ff2d05fc352ad351efff108d47fb150c92cf47`(파일 단위).
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
| B1 (range@558) | `for _, tc := range cases {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@562) | `if tc.trading != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (if@567) | `if tc.writeAtt {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B4 (if@574) | `if tc.name == "no Guardian injected" {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B5 (else@576) | `} else {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B6 (if@579) | `if err == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B7 (if@582) | `if eng != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B8 (if@585) | `if !errors.Is(err, engine.ErrAutomationGateRefused) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B9 (if@588) | `if !errors.Is(err, tc.wantErr) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B10 (if@591) | `if tc.wantMessage != "" && !strings.Contains(err.Error(), tc.wantMessage) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 59개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 임시 디렉터리 · httptest.

## Safety conclusion

- a199 편집(예정): 낡음 사례 셋(「attestation expired」·「misses an endpoint the engine calls」·「predates the price read」)이 거절 기대에서 청산 전용 기대로.
- Safe edit boundary: 위 편집 칸 밖의 분기는 무편집이 목표 — 2단계 post-edit 지도가 옛/새 AST 를 정렬해 확인.
- High-risk impact: no(시험 함수)
