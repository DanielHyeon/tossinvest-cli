# Function Logic Map: `TestA196ScopeLossWordingFollowsTheMissingKind`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-24, HEAD `e4165fa3d`(코드 = `9738a0c32`).** 분기 표는 AST 열거(`../../harness/build_bundles.py`), 2단계 GREEN 뒤 다시 만든다.

- Source: `internal/app/engine/a196_attestation_deadline_internal_test.go` (542-598)
- AST evidence: `ast.json` — branches **10** / returns **0** / calls **26**, source SHA-256 `1cd44b304b287c353b6ec200163e414bc025ff1f0f33c57bccab35fb6d9210af`(파일 단위).
  명령 `go run ./tools/logic-map --file internal/app/engine/a196_attestation_deadline_internal_test.go --func TestA196ScopeLossWordingFollowsTheMissingKind`.
- Risk scan: `risk-pattern-report.md`.
- 의도: 범위 상실 문장의 원인 · 할 일이 누락 종류를 따름.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 사례 넷 | POST 둘 · POST 하나 · GET 하나 · POST+GET | — | 아래 분기 표 |

## Branches and early returns

| Branch | Condition (소스 그 줄) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (range@561) | `for _, tc := range cases {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (range@564) | `for _, e := range tc.drop {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (if@571) | `if len(got) != 1 {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B4 (if@575) | `if !strings.Contains(e.Title, a196Refused) \|\| !strings.Contains(e.Title, tc.title) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B5 (if@578) | `if !strings.HasPrefix(e.Body, a196Refused+".") {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B6 (range@581) | `for _, d := range tc.drop {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B7 (if@582) | `if !strings.Contains(e.Body, d) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B8 (range@586) | `for _, want := range []string{"자동 손절이 멈춘다", "지금 실행 중인 엔진은 계속 돈다"} {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B9 (if@587) | `if !strings.Contains(e.Body, want) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B10 (if@591) | `if strings.Contains(e.Body, cause) != tc.cause \|\| strings.Contains(e.Body, "tossctl verify run") != tc.ve…` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 26개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- a199 편집(예정): 접두 단언(「기동이 거부된다」)을 청산 전용 귀결로.
- Safe edit boundary: 위 편집 칸 밖의 분기는 무편집이 목표 — 2단계 post-edit 지도가 옛/새 AST 를 정렬해 확인.
- High-risk impact: no(시험 함수)
