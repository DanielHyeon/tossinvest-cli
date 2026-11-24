# Function Logic Map: `a135Watch`

> **편집 후 지도(post-edit) — 2단계 GREEN 뒤, 2026-11-24, base `603046a23`.** 분기 표는 AST 열거(`../../harness/build_bundles.py --post`). 편집 전 원본 AST 는 `../../pre-edit-ast/internal-app-engine--a135watch.ast.json`.

- Source: `internal/app/engine/a135_attestation_watch_internal_test.go` (170-188)
- AST evidence: `ast.json` — branches **1** / returns **4** / calls **1**, source SHA-256 `56db9d94fc20556db8bbae659296b2c6c2e505c4bae178b95404e6cd93513afe`(파일 단위).
  명령 `go run ./tools/logic-map --file internal/app/engine/a135_attestation_watch_internal_test.go --func a135Watch`.
- Risk scan: `risk-pattern-report.md`.
- 의도: a135 · a196 · a199 감시 시험의 감시 조립(가짜 시계 · 기록 알림기 · 바꿔 끼우는 파일).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `att` | nil 이면 ErrMissing | 시험 | 아래 분기 표 |
| 불변식 | 생산 감시와 같은 칸 집합 | attestationWatch | 아래 분기 표 |

## Branches and early returns

| Branch | Condition (소스 그 줄) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@180) | `if att == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 1개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 없음(감시 값 하나를 돌려줌).

## Safety conclusion

- a199 편집(적용됨): G2: `proof` 칸에 「이 계좌로 엔진 주문이 확정된 기록 있음」을 넣음 — 기존 시험의 귀결 문장 무변, 기록 없는 설치는 a199 시험이 바꿔 끼움.
- Safe edit boundary: 편집 전 AST(`../../pre-edit-ast/`)와 비교해 위 편집 칸 밖의 분기는 무편집. 행동 고정은 a199 시험 · 변이 원장(`../../mutation/`).
- High-risk impact: no(시험 함수)
