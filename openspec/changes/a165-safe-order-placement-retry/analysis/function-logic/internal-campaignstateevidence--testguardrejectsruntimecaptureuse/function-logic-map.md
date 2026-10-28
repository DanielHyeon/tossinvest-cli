# Function Logic Map: `TestGuardRejectsRuntimeCaptureUse`

- Source: `internal/campaignstateevidence/authority_guard_test.go`
- AST evidence: `ast.json`(**revision base** — base `2bbee345f` 의 소스, 분기 3. 게이트는 `--unified=0` 의 순수 삽입 훅(@@ -105,0 …)이 이 함수의 닫는 줄과 맞닿아 base 쪽 함수로만 잡는다)
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a170-inspection-guard-reconciles-offline-assessment/base-commit.txt`
- a170 은 이 함수의 본문을 바꾸지 않았다(diff 의 `-` 줄 0). 게이트의 수정 집합이 이 함수를 잡는 것은 a170 이 바로 뒤에 새 시험 함수를 붙인 순수 삽입 훅 때문이고
  (`check_analysis.py` 의 base 쪽 `intersects`), 그래서 번들은 base 소스를 `revision: base` 로 고정한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| lane import 리터럴 2 표기 | 평문 · `a` 이스케이프 | 시험 리터럴 | 하나라도 통과하면 `t.Fatalf` |
| 점 import 원문 | `import .` continuationlane | 시험 리터럴 | 통과하면 `t.Fatalf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range` 두 import 표기 | 없음 | — | 이 시험 자신 |
| B2 | `allowedImport` 가 오류 또는 허용 | 없음 | `t.Fatalf`(시험 실패) | 이 시험 자신 |
| B3 | 점 import 원문이 오류 또는 허용 | 없음 | `t.Fatalf` | 이 시험 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `allowedImport` | 엔진 경로에 둔 위조 원문을 가드가 거절하는지 | 오류도 실패로 침 | AST |

## State mutations and fallbacks

- 없음(시험 함수).

## Safety conclusion

- Safe edit boundary: 본문 불변 — a170 의 B8 · B9 추가가 이 시험의 엔진 경로 위조 원문 판정을 바꾸지 않는다(GREEN, green.txt).
- High-risk impact: no.
