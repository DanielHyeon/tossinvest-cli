# Function Logic Map: `logGateDecision`

> **편집 후 지도(post-edit) — 2단계 GREEN 뒤, 2026-11-24, base `603046a23`.** 분기 표는 AST 열거(`../../harness/build_bundles.py --post`). 편집 전 원본 AST 는 `../../pre-edit-ast/internal-app-engine--loggatedecision.ast.json`.

- Source: `internal/app/engine/interlock.go` (472-512)
- AST evidence: `ast.json` — branches **3** / returns **2** / calls **16**, source SHA-256 `ed64648db576432a0100a9c8310b17ee18bd3f7f24067368acea4637759deb75`(파일 단위).
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
| B1 (if@496) | `if status.RunsExitSide() && status.Protection != ProtectionWired {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@501) | `if refusal != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (if@505) | `if status.ExitOnly {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 16개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 로그 한 줄.

## Safety conclusion

- a199 편집(적용됨): (조건부 — Q) B1@479 의 「보호는 프로세스 수명」 문장을 청산 전용 기동에도, 새 표지 필드 하나.
- Safe edit boundary: 편집 전 AST(`../../pre-edit-ast/`)와 비교해 위 편집 칸 밖의 분기는 무편집. 행동 고정은 a199 시험 · 변이 원장(`../../mutation/`).
- High-risk impact: yes — 기동 인터록 · 주문 권한 · 손절 경로 가용성(안전 부트스트랩 5)
