# Function Logic Map: `logGateDecision`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-24, HEAD `e4165fa3d`(코드 = `9738a0c32`).** 분기 표는 AST 열거(`../../harness/build_bundles.py`), 2단계 GREEN 뒤 다시 만든다.

- Source: `internal/app/engine/interlock.go` (456-489)
- AST evidence: `ast.json` — branches **2** / returns **1** / calls **12**, source SHA-256 `d018e5cf62544eef99ad68fadebdbe0de63f4677d0bea51d39a33ccc18f037e1`(파일 단위).
  명령 `go run ./tools/logic-map --file internal/app/engine/interlock.go --func logGateDecision`.
- Risk scan: `risk-pattern-report.md`.
- 의도: 기동마다 구조 로그 한 줄(`engine.operating_mode`) — 거절이면 Warn + reason.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `status` | runInterlock | AutomationStatus | 아래 분기 표 |
| `refusal` | verifyGate 오류 또는 nil | — | 아래 분기 표 |

## Branches and early returns

| Branch | Condition (소스 그 줄) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@479) | `if status.Verified && status.Protection != ProtectionWired {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@484) | `if refusal != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 12개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 로그 한 줄.

## Safety conclusion

- a199 편집(예정): (조건부 — Q) B1@479 의 「보호는 프로세스 수명」 문장을 청산 전용 기동에도, 새 표지 필드 하나.
- Safe edit boundary: 위 편집 칸 밖의 분기는 무편집이 목표 — 2단계 post-edit 지도가 옛/새 AST 를 정렬해 확인.
- High-risk impact: yes — 기동 인터록 · 주문 권한 · 손절 경로 가용성(안전 부트스트랩 5)
