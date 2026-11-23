# Function Logic Map: `TestA196RefusalTellsTheConsequenceFirst`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-24, HEAD `e4165fa3d`(코드 = `9738a0c32`).** 분기 표는 AST 열거(`../../harness/build_bundles.py`), 2단계 GREEN 뒤 다시 만든다.

- Source: `internal/app/engine/a196_attestation_deadline_internal_test.go` (439-482)
- AST evidence: `ast.json` — branches **8** / returns **0** / calls **23**, source SHA-256 `1cd44b304b287c353b6ec200163e414bc025ff1f0f33c57bccab35fb6d9210af`(파일 단위).
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
| B3 (if@464) | `if !strings.Contains(e.Title, a196Refused) \|\| strings.Contains(e.Title, "수 있다") {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B4 (if@467) | `if !strings.HasPrefix(e.Body, a196Refused) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B5 (if@470) | `if !strings.Contains(e.Body, "손절") \|\| !strings.Contains(e.Body, "할 일") {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B6 (if@473) | `if name == "scope_lost" {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B7 (range@474) | `for _, want := range []string{"POST /api/v1/orders", "tossctl verify run"} {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B8 (if@475) | `if !strings.Contains(e.Body, want) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 23개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- a199 편집(예정): scope_lost · expired 의 기대 문장을 청산 전용 귀결로, 나머지 셋 무변.
- Safe edit boundary: 위 편집 칸 밖의 분기는 무편집이 목표 — 2단계 post-edit 지도가 옛/새 AST 를 정렬해 확인.
- High-risk impact: no(시험 함수)
