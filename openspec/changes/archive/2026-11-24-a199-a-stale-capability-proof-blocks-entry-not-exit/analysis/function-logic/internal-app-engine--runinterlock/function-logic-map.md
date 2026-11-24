# Function Logic Map: `runInterlock`

> **편집 후 지도(post-edit) — 2단계 GREEN 뒤, 2026-11-24, base `603046a23`.** 분기 표는 AST 열거(`../../harness/build_bundles.py --post`). 편집 전 원본 AST 는 `../../pre-edit-ast/internal-app-engine--runinterlock.ast.json`.

- Source: `internal/app/engine/interlock.go` (405-440)
- AST evidence: `ast.json` — branches **4** / returns **5** / calls **9**, source SHA-256 `7d2ce7348681c518284279c217092c1f7ad4cc9662ceadf37df3221341497099`(파일 단위).
  명령 `go run ./tools/logic-map --file internal/app/engine/interlock.go --func runInterlock`.
- Risk scan: `risk-pattern-report.md`.
- 의도: 게이트 OFF 면 즉시 반환, ON 이면 verifyGate → 거절(감사 · 로그) 또는 수락(Verified · EntryPermitted · 감사 · 로그).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `gate.Enabled` | 설정 | config | 아래 분기 표 |
| `log`·`logger` | NewContext | 감사 · 구조 로그 | 아래 분기 표 |
| 불변식 | 수락 감사 기록 실패면 기동 거절(B3) | §0.5 | 아래 분기 표 |

## Branches and early returns

| Branch | Condition (소스 그 줄) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@408) | `if !gate.Enabled {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@415) | `if err := verifyGate(&status, facts); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (if@420) | `if status.ExitOnly {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B4 (if@429) | `if err := log.Record(audit.Entry{` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 9개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- `status.Verified`(406) · `status.EntryPermitted`(409) 씀, 감사 1줄.

## Safety conclusion

- a199 편집(적용됨): 세 번째 결과(청산 전용): Verified=false 유지 · 새 표지 · EntryPermitted=false · 새 감사 action · 진입 래치(D3·D4·D5·D8).
- Safe edit boundary: 편집 전 AST(`../../pre-edit-ast/`)와 비교해 위 편집 칸 밖의 분기는 무편집. 행동 고정은 a199 시험 · 변이 원장(`../../mutation/`).
- High-risk impact: yes — 기동 인터록 · 주문 권한 · 손절 경로 가용성(안전 부트스트랩 5)
