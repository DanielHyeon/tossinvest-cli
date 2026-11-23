# Function Logic Map: `AllReasonCodes`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-24, HEAD `e4165fa3d`(코드 = `9738a0c32`).** 분기 표는 AST 열거(`../../harness/build_bundles.py`), 2단계 GREEN 뒤 다시 만든다.

- Source: `internal/execgw/failclosed.go` (254-311)
- AST evidence: `ast.json` — branches **0** / returns **2** / calls **1**, source SHA-256 `fbe97b79db74c5e7801fd3e1c6301d13f261d0d6c5e880308b36323ca328ebd4`(파일 단위).
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
| (분기 없음) | 한 갈래(happy path 254-311) | — | — | 아래 BTM B1 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 1개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- a199 편집(예정): 새 사유 코드 한 줄 등록(D4). golden · a098 등록 변수 갱신.
- Safe edit boundary: 위 편집 칸 밖의 분기는 무편집이 목표 — 2단계 post-edit 지도가 옛/새 AST 를 정렬해 확인.
- High-risk impact: yes — 기동 인터록 · 주문 권한 · 손절 경로 가용성(안전 부트스트랩 5)
