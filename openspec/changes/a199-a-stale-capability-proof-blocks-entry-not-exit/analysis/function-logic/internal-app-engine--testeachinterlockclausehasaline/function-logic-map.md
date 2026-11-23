# Function Logic Map: `TestEachInterlockClauseHasALine`

> **편집 후 지도(post-edit) — 2단계 GREEN 뒤, 2026-11-24, base `603046a23`.** 분기 표는 AST 열거(`../../harness/build_bundles.py --post`). 편집 전 원본 AST 는 `../../pre-edit-ast/internal-app-engine--testeachinterlockclausehasaline.ast.json`.

- Source: `internal/app/engine/runtime_wiring_test.go` (80-141)
- AST evidence: `ast.json` — branches **8** / returns **0** / calls **21**, source SHA-256 `87fdcdbf6c662efee6d0fe72537a749a2a3b32c156df934378047a224d56c976`(파일 단위).
  명령 `go run ./tools/logic-map --file internal/app/engine/runtime_wiring_test.go --func TestEachInterlockClauseHasALine`.
- Risk scan: `risk-pattern-report.md`.
- 의도: 거절 조항마다 열거 한 줄.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 「an expired attestation」 사례 | 95 | — | 아래 분기 표 |

## Branches and early returns

| Branch | Condition (소스 그 줄) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (range@110) | `for name, tc := range cases {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@113) | `if tc.policy != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (else@115) | `} else {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B4 (if@123) | `if !tc.noGuardian {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B5 (if@127) | `if tc.noGuardian {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B6 (else@129) | `} else {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B7 (if@132) | `if err == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B8 (if@136) | `if !strings.Contains(joined, tc.wantHit) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 21개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 임시 디렉터리.

## Safety conclusion

- a199 편집(적용됨): 만료 사례를 만료 + 다른 계좌(D1a) 등 거절 유지 모양으로.
- Safe edit boundary: 편집 전 AST(`../../pre-edit-ast/`)와 비교해 위 편집 칸 밖의 분기는 무편집. 행동 고정은 a199 시험 · 변이 원장(`../../mutation/`).
- High-risk impact: no(시험 함수)
