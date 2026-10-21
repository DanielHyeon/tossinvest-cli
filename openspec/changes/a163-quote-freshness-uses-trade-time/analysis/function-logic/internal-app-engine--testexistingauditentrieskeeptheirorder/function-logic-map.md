# Function Logic Map: `TestExistingAuditEntriesKeepTheirOrder`

- Source: `internal/app/engine/a074_notification_audit_test.go` (119-154)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- 기준: 편집 전(base `f7d80c348`) — 시험 함수
- AST 분기 5 · 반환 0 · defer 0 · go 0
- 위험 등급: 시험 전용

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| audit 고정물 | 임시 경로 | `auditFixture` | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 124 `if err := recordGateSettings(log, config.AutomationGate{LimitCurrency: "KRW"},` | — | `t.Fatalf` | 자기 자신 |
| B2 | if 130 `if err != nil {` | — | `t.Fatalf` | 자기 자신 |
| B3 | if 146 `if len(entries) < len(want) {` | — | `t.Fatalf`(앞부분 행 수) | 자기 자신 |
| B4 | range 149 `for i, setting := range want {` | 순서 순회 | — | 자기 자신 |
| B5 | if 150 `if entries[i].Setting != setting {` | — | `t.Fatalf` | 자기 자신 |

반환 좌표(AST): 없음

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `recordGateSettings` | 대상 호출 | 오류는 Fatalf | AST |

## State mutations and fallbacks

- 없음.

## a163 편집

- a163 3.6: `recordGateSettings` 호출에 끝 인자 `false` 하나. 앞부분 순서 단언 무편집 — 새 행은 끝에 붙어 양립.

## Safety conclusion

- Safe edit boundary: 호출 인자 하나.
- High-risk impact: no.
