# Function Logic Map: `newSoakAttestCmd`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-24, HEAD `e4165fa3d`(코드 = `9738a0c32`).** 분기 표는 AST 열거(`../../harness/build_bundles.py`), 2단계 GREEN 뒤 다시 만든다.

- Source: `cmd/tossctl/soak.go` (199-241)
- AST evidence: `ast.json` — branches **0** / returns **2** / calls **18**, source SHA-256 `3c1263d5d690d192265d0b255e91840620dae6c25e1d1b7fa8c3080b6e50162c`(파일 단위).
  명령 `go run ./tools/logic-map --file cmd/tossctl/soak.go --func newSoakAttestCmd`.
- Risk scan: `risk-pattern-report.md`.
- 의도: `soak attest` 명령 정의 — Long 도움말.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 도움말 문자열 | 상수 | 220 | 아래 분기 표 |

## Branches and early returns

| Branch | Condition (소스 그 줄) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (분기 없음) | 한 갈래(happy path 199-241) | — | — | 아래 BTM B1 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 18개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- a199 편집(예정): 결정 11: 도움말 「the gate will still refuse」 문구를 새 귀결로.
- Safe edit boundary: 위 편집 칸 밖의 분기는 무편집이 목표 — 2단계 post-edit 지도가 옛/새 AST 를 정렬해 확인.
- High-risk impact: yes — 기동 인터록 · 주문 권한 · 손절 경로 가용성(안전 부트스트랩 5)
