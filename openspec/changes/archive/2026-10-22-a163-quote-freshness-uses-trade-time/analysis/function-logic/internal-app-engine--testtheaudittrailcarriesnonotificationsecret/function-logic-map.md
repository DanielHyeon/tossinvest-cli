# Function Logic Map: `TestTheAuditTrailCarriesNoNotificationSecret`

- Source: `internal/app/engine/a074_notification_audit_test.go` (78-99)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- 기준: 편집 뒤(브랜치 `a163-impl` GREEN 로트, 2026-10-22). 편집 전 번들 문서는 커밋 `78efe1951` · 생성 입력 `../../flm-gen/spec-base.json`
- AST 분기 4 · 반환 0 · defer 0 · go 0
- 위험 등급: 시험 전용

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| audit 고정물 | 임시 경로 | `auditFixture` | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 84 `if err := recordGateSettings(log, config.AutomationGate{}, config.Adoption{},` | — | `t.Fatalf` | 자기 자신 |
| B2 | if 90 `if err != nil {` | — | `t.Fatalf` | 자기 자신 |
| B3 | range 93 `for _, e := range entries {` | 행 순회 | — | 자기 자신 |
| B4 | if 95 `if strings.Contains(line, "top-secret") {` | — | `t.Fatalf` | 자기 자신 |

반환 좌표(AST): 없음

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `recordGateSettings` | 대상 호출 | 오류는 Fatalf | AST |

## State mutations and fallbacks

- 없음.

## a163 편집

- a163 3.6: `recordGateSettings` 호출에 끝 인자 `false` 하나. 단언 무편집. — **적용됨**, 편집 전·뒤 AST 분기·반환·defer 수 대조는 `../../flm-gen/README.md`

## Safety conclusion

- Safe edit boundary: 호출 인자 하나.
- High-risk impact: no.
