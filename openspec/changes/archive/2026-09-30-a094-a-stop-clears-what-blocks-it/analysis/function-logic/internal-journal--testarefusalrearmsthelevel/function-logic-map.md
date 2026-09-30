# Function Logic Map: `TestARefusalReArmsTheLevel`

- Source: `internal/journal/exit_state_test.go` (`375`–`420`)
- Qualified: `TestARefusalReArmsTheLevel`
- AST evidence: `ast.json` (`source_sha256` 0bfab2fe1cdb4370…)
- Risk scan: `risk-pattern-report.md`

**역할.** 기존 시험 — ResolveExitProposal 이 기대 intent 를 요구하므로 발의에 intent 를 달고 그 intent 로 해제(단언 무변).

## Inputs and invariants

시험 픽스처(임시 원장). 생산 코드 아님 — 비례 원칙상 이 번들은 게이트의 요구 집합(수정된 기존 함수)을 채우는 기록이다.

## Branches and early returns

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:381` `if err := j.AttachExitIntent(ctx, p.ID, "i-refused"); err != nil {` | — |
| B2 | if | `:386` `if err := j.ResolveExitProposal(ctx, p.ID, "i-refused", ProposalRefused); err != nil {` | — |
| B3 | if | `:390` `if state.Pending() {` | — |
| B4 | if | `:393` `if state.TakenRatioTotal != "0" {` | — |
| B5 | if | `:397` `if err := j.RecordExitJudgement(ctx, ExitJudgement{` | — |
| B6 | if | `:408` `if err != nil {` | — |
| B7 | range | `:412` `for _, e := range events {` | — |
| B8 | if | `:413` `if e.Action == ExitEventProposalRefused {` | — |
| B9 | if | `:417` `if !refused {` | — |

## Calls and live bindings

시험 대상 API(원장 · 관측 루프)만 부른다. 브로커 호출 0.

## State mutations and fallbacks

임시 원장만.

## Safety conclusion

- High-risk impact: no(시험). 단언은 약화되지 않았다 — 바뀐 것은 새 호출 형태(기대 intent) 또는 이름 붙은 대가의 +1 관측뿐.
