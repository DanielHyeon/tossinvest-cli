# Function Logic Map: `TestRefusedStartIsAudited`

> **편집 후 지도(post-edit) — 2단계 GREEN 뒤, 2026-11-24, base `603046a23`.** 분기 표는 AST 열거(`../../harness/build_bundles.py --post`). 편집 전 원본 AST 는 `../../pre-edit-ast/internal-app-engine--testrefusedstartisaudited.ast.json`.

- Source: `internal/app/engine/interlock_test.go` (626-647)
- AST evidence: `ast.json` — branches **3** / returns **0** / calls **17**, source SHA-256 `b564dfcd90d2a19953783d191ce6793b97258eace99b95c36ac2d4805f12881f`(파일 단위).
  명령 `go run ./tools/logic-map --file internal/app/engine/interlock_test.go --func TestRefusedStartIsAudited`.
- Risk scan: `risk-pattern-report.md`.
- 의도: 거절된 기동의 감사 1줄.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 픽스처 | 만료 attestation(653) | — | 아래 분기 표 |

## Branches and early returns

| Branch | Condition (소스 그 줄) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@634) | `if _, err := openGateEngine(t, dir, srv, matchedGuardian()); err == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@640) | `if refused == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (if@644) | `if !strings.Contains(refused.Detail, attest.ErrAccountMismatch.Error()) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 17개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 임시 디렉터리.

## Safety conclusion

- a199 편집(적용됨): 픽스처를 거절이 유지되는 결함으로 교체(만료는 청산 전용이 됨).
- Safe edit boundary: 편집 전 AST(`../../pre-edit-ast/`)와 비교해 위 편집 칸 밖의 분기는 무편집. 행동 고정은 a199 시험 · 변이 원장(`../../mutation/`).
- High-risk impact: no(시험 함수)
