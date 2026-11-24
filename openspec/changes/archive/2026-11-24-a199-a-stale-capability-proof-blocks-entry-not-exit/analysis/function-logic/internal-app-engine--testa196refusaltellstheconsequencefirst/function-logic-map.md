# Function Logic Map: `TestA196RefusalTellsTheConsequenceFirst`

> **편집 후 지도(post-edit) — 2단계 GREEN 뒤, 2026-11-24, base `603046a23`.** 분기 표는 AST 열거(`../../harness/build_bundles.py --post`). 편집 전 원본 AST 는 `../../pre-edit-ast/internal-app-engine--testa196refusaltellstheconsequencefirst.ast.json`.

- Source: `internal/app/engine/a196_attestation_deadline_internal_test.go` (439-487)
- AST evidence: `ast.json` — branches **9** / returns **0** / calls **23**, source SHA-256 `1a3a3148ee8fb55b56efc8657f28e703cc51912273747e3a92eae7919a519dd2`(파일 단위).
  명령 `go run ./tools/logic-map --file internal/app/engine/a196_attestation_deadline_internal_test.go --func TestA196RefusalTellsTheConsequenceFirst`.
- Risk scan: `risk-pattern-report.md`.
- 의도: 거절 단계 알림의 제목 · 본문 첫 문장 = 「기동이 거부된다」.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 사례 다섯 | scope_lost · expired · incomplete · account_mismatch · unreadable | — | 아래 분기 표 |

## Branches and early returns

| Branch | Condition (소스 그 줄) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (range@453) | `for name, tc := range cases {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@460) | `if len(got) != 1 \|\| got[0].Fields[obs.FieldReason] != name {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (if@466) | `if name == "scope_lost" \|\| name == "expired" {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B4 (if@469) | `if !strings.Contains(e.Title, consequence) \|\| strings.Contains(e.Title, "수 있다") {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B5 (if@472) | `if !strings.HasPrefix(e.Body, consequence) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B6 (if@475) | `if !strings.Contains(e.Body, "손절") \|\| !strings.Contains(e.Body, "할 일") {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B7 (if@478) | `if name == "scope_lost" {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B8 (range@479) | `for _, want := range []string{"POST /api/v1/orders", "tossctl verify run"} {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B9 (if@480) | `if !strings.Contains(e.Body, want) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 23개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- a199 편집(적용됨): scope_lost · expired 의 기대 문장을 청산 전용 귀결로, 나머지 셋 무변.
- Safe edit boundary: 편집 전 AST(`../../pre-edit-ast/`)와 비교해 위 편집 칸 밖의 분기는 무편집. 행동 고정은 a199 시험 · 변이 원장(`../../mutation/`).
- High-risk impact: no(시험 함수)
