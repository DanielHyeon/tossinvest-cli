# Function Logic Map: `AllReasonCodes`

> **편집 후 지도(post-edit) — 2단계 GREEN 뒤, 2026-11-24, base `603046a23`.** 분기 표는 AST 열거(`../../harness/build_bundles.py --post`). 편집 전 원본 AST 는 `../../pre-edit-ast/internal-execgw--allreasoncodes.ast.json`.

- Source: `internal/execgw/failclosed.go` (254-314)
- AST evidence: `ast.json` — branches **0** / returns **2** / calls **1**, source SHA-256 `fc611661e639d79b5c75bd90cbed98d13e468004d69a8d142ac3ab885eb39acb`(파일 단위).
  명령 `go run ./tools/logic-map --file internal/execgw/failclosed.go --func AllReasonCodes`.
- Risk scan: `risk-pattern-report.md`.
- 의도: 이 빌드가 낼 수 있는 사유 코드 전부(정렬) — golden 과 a098 길이 단언의 입력.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 없음 | — | — | 아래 분기 표 |

## Branches and early returns

| Branch | Condition (소스 그 줄) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (분기 없음) | 한 갈래(happy path 254-314) | — | — | 아래 BTM B1 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 1개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- a199 편집(적용됨): 새 사유 코드 한 줄 등록(D4). golden · a098 등록 변수 갱신.
- Safe edit boundary: 편집 전 AST(`../../pre-edit-ast/`)와 비교해 위 편집 칸 밖의 분기는 무편집. 행동 고정은 a199 시험 · 변이 원장(`../../mutation/`).
- High-risk impact: yes — 기동 인터록 · 주문 권한 · 손절 경로 가용성(안전 부트스트랩 5)
