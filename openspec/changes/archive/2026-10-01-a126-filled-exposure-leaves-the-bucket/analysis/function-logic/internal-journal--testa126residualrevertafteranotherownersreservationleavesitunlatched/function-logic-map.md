# Function Logic Map: `TestA126ResidualRevertAfterAnotherOwnersReservationLeavesItUnlatched`

- Source: `internal/journal/a126_filled_exposure_leaves_the_bucket_test.go` (`551`–`578`) — base 판본
- Qualified: `TestA126ResidualRevertAfterAnotherOwnersReservationLeavesItUnlatched`
- AST evidence: `ast.json` (`source_sha256` 5dbfbfb20f105166…, revision base) — **편집 전(base — 이 함수는 1.5.2 에서 개명으로 지워짐)**(시험 전용)
- Risk scan: `risk-pattern-report.md`
- AST branches 6 · return 1 · 호출 19

**역할.** codex #2 잔여 핀의 옛 판본(admission 대리). 비례 원칙(WORKFLOW 「비례 원칙」): 시험 전용 — 무거운 규율 대상 아님. base 재고정(review 2.1) 뒤 게이트가 수정 함수로 세므로 경량 번들을 둔다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 시험 임시 원장 · fixture | 시험 | 생산 작성자 경로 + 손상만 SQL | `t.Fatal`/`t.Errorf` |

## Branches and early returns

> AST 분기 열거(시험 코드 — 커버리지 블록 없음).

| Branch | 종류 | 조건 (원문) |
|---|---|---|
| B1 | if | `:556` `if err := a126AdmitSymbol(t, j, "c2-b", "MSFT", "100", "0", 12); err != nil {` |
| B2 | if | `:559` `if _, err := j.db.Exec(`UPDATE mutation_attempts SET state='CONFIRMED' WHERE id='a126-late-attempt-c2'`); e…` |
| B3 | if | `:562` `if err := j.SetApplyHooks(ApplyHooks{Campaign: func(context.Context, *ApplyTx, AppliedFill) error { return …` |
| B4 | if | `:565` `if res, err := j.RecordFill(context.Background(), a126LateFill(o, order)); err != nil \|\| !res.Changed {` |
| B5 | if | `:569` `if usage.FilledMinor != "50" \|\| usage.HeldMinor != "60" {` |
| B6 | if | `:573` `if err := j.db.QueryRow(`SELECT (SELECT count(*) FROM risk_bucket_scope_latches WHERE prospective_generatio…` |

## Calls and live bindings

시험 fixture · 생산 journal/riskbucket API 호출. 브로커 · 실계좌 호출 없음.

## State mutations and fallbacks

시험 임시 원장만.

## Safety conclusion

- **Safe edit boundary**: 지워지고 TestA126ResidualRevertAfterAnotherOwnersIssuancePassesSubmitRevalidation(실제 발급 + 제출 재검증)으로 대체(R2, Manager 판정 (가)). latch 0 단언은 새 시험에 그대로 옮김.
- **High-risk impact**: no — 시험. 생산 동작 변화 0(`64ad3058` 비시험 `.go` 변경 0 — codex 재리뷰 확인).
