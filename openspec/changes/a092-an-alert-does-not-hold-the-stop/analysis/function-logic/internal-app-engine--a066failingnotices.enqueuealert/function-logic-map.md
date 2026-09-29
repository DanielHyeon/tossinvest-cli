# Function Logic Map: `a066FailingNotices.EnqueueAlert`

- Source: `internal/app/engine/a066_risk_relaxation_test.go`
- AST evidence: `ast.json` — **base**(삭제 전, `revision: base`), :214–216, 분기 0, source_sha256 `78c53266581f…`, 추출 base `721d0338`.
- Risk scan: `risk-pattern-report.md`(base 파일)
- 편집: **삭제된 메서드**. a092 25.6 이 해제 원장 면(`riskRelaxationRepository`)에서 `EnqueueAlert` 를 없애 덮어쓸 대상이 사라짐.
  같은 이름의 타입(`a066FailingNotices`)은 이제 `RecordCritical` 을 실패시키는 기록자다(새 메서드).
- 비례 원칙: 시험 코드 → 변이 원장 · 다중 리뷰 `not-applicable`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| (없음) | — | — | 상수 오류 `outbox disk full` 반환 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| — | 분기 없음 | — | `0, errors.New("outbox disk full")` | (삭제됨 — 대체: `a066FailingNotices.RecordCritical`, `TestA066ReleaseStandsWhenTheNoticeFails`) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `errors.New` | 실패 주입 | 상수 | AST(base) |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Safe edit boundary: 시험 가짜의 삭제 — 생산 코드 무관.
- High-risk impact: no — 시험 코드.
