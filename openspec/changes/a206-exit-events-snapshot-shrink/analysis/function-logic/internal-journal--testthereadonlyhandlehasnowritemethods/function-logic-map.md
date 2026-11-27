# Function Logic Map: `TestTheReadOnlyHandleHasNoWriteMethods`

- Source: `internal/journal/readonly_test.go` (116-167)
- Qualified function: `TestTheReadOnlyHandleHasNoWriteMethods` (시험 함수)
- Revision: `current` (편집 뒤, `source_sha256` 8cc018cdb848bee5…; base 파일 f4b38427349a3a51…)
- AST evidence: `ast.json` — AST branches 2, 반환 0
- Risk scan: `risk-pattern-report.md`
- 편집: 허용 집합 `allowed` 에 `"ExitEvidenceCompactionPlan": true` 한 줄과 근거 주석 2줄. 패키지 전체 시험이 이 census 의 FAIL 을 보여 준 뒤 더함(사후 작성).
  비례 원칙상 시험 전용 편집이지만 게이트 5단계가 수정된 기존 함수로 세므로 번들을 둠.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `allowed` | 메서드 이름 → true | 이 시험의 고정 집합(추가마다 근거 주석) | 집합 밖 메서드면 `t.Errorf` |
| `reflect.TypeOf(&ReadOnly{})` | `*ReadOnly` 의 메서드 전부 | `readonly.go` · `account_views.go` · `a206_exit_evidence_compact.go` 등 | — |

불변식: `*ReadOnly` 의 메서드는 전부 집합 안에 있음(쓰기 메서드가 콘솔 핸들에 생기지 않음).

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for at 160:2` 메서드 순회 | — | 없음 | 이 시험 자체 |
| B2 | `if at 162:3` 집합 밖 메서드 | `t.Errorf` | 없음 | 이 시험(줄 추가 전 「unlisted method "ExitEvidenceCompactionPlan"」 발화 관측) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `reflect.TypeOf` · `NumMethod` · `Method` | 메서드 열거 | 순수 | AST |
| `t.Errorf` | 실패 보고 | 없음 | AST |

## State mutations and fallbacks

- 없음(시험).

## Safety conclusion

- Safe edit boundary: 집합에 한 줄 추가뿐 — 다른 메서드의 판정 불변.
- High-risk impact: no(시험). 더한 메서드는 count · sum SELECT 하나이고 query-only 연결을 씀. 쓰기(`CompactClosedExitEvidence`)는 `*Journal` 에만 있음.
