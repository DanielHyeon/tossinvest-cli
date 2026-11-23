# Function Logic Map: `TestA196ScopeLossWordingFollowsTheMissingKind`

> **편집 후 지도(post-edit) — 2단계 GREEN 뒤, 2026-11-24, base `603046a23`.** 분기 표는 AST 열거(`../../harness/build_bundles.py --post`). 편집 전 원본 AST 는 `../../pre-edit-ast/internal-app-engine--testa196scopelosswordingfollowsthemissingkind.ast.json`.

- Source: `internal/app/engine/a196_attestation_deadline_internal_test.go` (547-605)
- AST evidence: `ast.json` — branches **10** / returns **0** / calls **26**, source SHA-256 `1a3a3148ee8fb55b56efc8657f28e703cc51912273747e3a92eae7919a519dd2`(파일 단위).
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
| B1 (range@566) | `for _, tc := range cases {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (range@569) | `for _, e := range tc.drop {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (if@576) | `if len(got) != 1 {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B4 (if@582) | `if !strings.Contains(e.Title, exitOnly) \|\| !strings.Contains(e.Title, tc.title) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B5 (if@585) | `if !strings.HasPrefix(e.Body, exitOnly+".") {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B6 (range@588) | `for _, d := range tc.drop {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B7 (if@589) | `if !strings.Contains(e.Body, d) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B8 (range@593) | `for _, want := range []string{"새 진입만 막힌다", "지금 실행 중인 엔진은 계속 돈다"} {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B9 (if@594) | `if !strings.Contains(e.Body, want) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B10 (if@598) | `if strings.Contains(e.Body, cause) != tc.cause \|\| strings.Contains(e.Body, "tossctl verify run") != tc.ve…` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 26개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- a199 편집(적용됨): 접두 단언(「기동이 거부된다」)을 청산 전용 귀결로.
- Safe edit boundary: 편집 전 AST(`../../pre-edit-ast/`)와 비교해 위 편집 칸 밖의 분기는 무편집. 행동 고정은 a199 시험 · 변이 원장(`../../mutation/`).
- High-risk impact: no(시험 함수)
