# Function Logic Map: `TestEachInterlockClauseHasALine`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-24, HEAD `e4165fa3d`(코드 = `9738a0c32`).** 분기 표는 AST 열거(`../../harness/build_bundles.py`), 2단계 GREEN 뒤 다시 만든다.

- Source: `internal/app/engine/runtime_wiring_test.go` (80-137)
- AST evidence: `ast.json` — branches **8** / returns **0** / calls **21**, source SHA-256 `cf812df46d4f90c47a84a73a0769d6982127aa323b45132b3b2ddad91d298681`(파일 단위).
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
| B1 (range@106) | `for name, tc := range cases {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@109) | `if tc.policy != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (else@111) | `} else {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B4 (if@119) | `if !tc.noGuardian {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B5 (if@123) | `if tc.noGuardian {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B6 (else@125) | `} else {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B7 (if@128) | `if err == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B8 (if@132) | `if !strings.Contains(joined, tc.wantHit) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 21개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 임시 디렉터리.

## Safety conclusion

- a199 편집(예정): 만료 사례를 만료 + 다른 계좌(D1a) 등 거절 유지 모양으로.
- Safe edit boundary: 위 편집 칸 밖의 분기는 무편집이 목표 — 2단계 post-edit 지도가 옛/새 AST 를 정렬해 확인.
- High-risk impact: no(시험 함수)
