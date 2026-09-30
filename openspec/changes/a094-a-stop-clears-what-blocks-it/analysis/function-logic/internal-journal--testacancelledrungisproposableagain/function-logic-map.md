# Function Logic Map: `TestACancelledRungIsProposableAgain`

- Source: `internal/journal/exit_state_test.go` (`435`–`472`)
- Qualified: `TestACancelledRungIsProposableAgain`
- AST evidence: `ast.json` (`source_sha256` 0bfab2fe1cdb4370…)
- Risk scan: `risk-pattern-report.md`

**역할.** 기존 시험 — 발의에 intent 를 달고 그 intent 로 해제(단언 무변: rung 되돌림 · 기준선 유지).

## Inputs and invariants

시험 픽스처(임시 원장). 생산 코드 아님 — 비례 원칙상 이 번들은 게이트의 요구 집합(수정된 기존 함수)을 채우는 기록이다.

## Branches and early returns

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:439` `if _, err := j.RecordFill(ctx, terminalFill(o, "10", "70000")); err != nil {` | — |
| B2 | if | `:443` `if _, err := j.OpenExitState(ctx, ExitStateSeed{` | — |
| B3 | if | `:450` `if err := j.RecordExitJudgement(ctx, ExitJudgement{` | — |
| B4 | if | `:457` `if got := exitStateOf(t, j, p.ID); got.ActiveRung != 1 {` | — |
| B5 | if | `:461` `if err := j.ResolveExitProposal(ctx, p.ID, "i-rung", ProposalCancelled); err != nil {` | — |
| B6 | if | `:465` `if state.ActiveRung != 0 {` | — |
| B7 | if | `:469` `if state.Baseline != "70700" {` | — |

## Calls and live bindings

시험 대상 API(원장 · 관측 루프)만 부른다. 브로커 호출 0.

## State mutations and fallbacks

임시 원장만.

## Safety conclusion

- High-risk impact: no(시험). 단언은 약화되지 않았다 — 바뀐 것은 새 호출 형태(기대 intent) 또는 이름 붙은 대가의 +1 관측뿐.
