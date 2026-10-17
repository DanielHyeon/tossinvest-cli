# Function Logic Map: `TestARefusedProposalIsReArmedAfterARestart`

- Source: `internal/journal/exit_state_test.go` (`790`–`821`)
- Qualified: `TestARefusedProposalIsReArmedAfterARestart`
- AST evidence: `ast.json` (`source_sha256` 0bfab2fe1cdb4370…)
- Risk scan: `risk-pattern-report.md`

**역할.** 기존 시험 — 발의에 intent 를 달고 그 intent 로 해제(단언 무변).

## Inputs and invariants

시험 픽스처(임시 원장). 생산 코드 아님 — 비례 원칙상 이 번들은 게이트의 요구 집합(수정된 기존 함수)을 채우는 기록이다.

## Branches and early returns

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:797` `if err := j.SetApplyHooks(ApplyHooks{Project: ProjectPosition, Exit: ApplyExitFill}); err != nil {` | — |
| B2 | if | `:803` `if err := j.AttachExitIntent(ctx, positionID, "i-restart"); err != nil {` | — |
| B3 | if | `:806` `if err := j.ResolveExitProposal(ctx, positionID, "i-restart", ProposalRefused); err != nil {` | — |
| B4 | if | `:812` `if err := restarted.RecordExitJudgement(ctx, ExitJudgement{` | — |

## Calls and live bindings

시험 대상 API(원장 · 관측 루프)만 부른다. 브로커 호출 0.

## State mutations and fallbacks

임시 원장만.

## Safety conclusion

- High-risk impact: no(시험). 단언은 약화되지 않았다 — 바뀐 것은 새 호출 형태(기대 intent) 또는 이름 붙은 대가의 +1 관측뿐.
