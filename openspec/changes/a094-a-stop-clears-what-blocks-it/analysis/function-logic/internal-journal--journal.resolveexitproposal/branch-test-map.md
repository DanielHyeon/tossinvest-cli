# Branch Test Map: `Journal.ResolveExitProposal`

- Source: `internal/journal/apply_hook.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다. 따라서 「Test」 열은 **a094가 요구하는 시험**이며 현존 증명이 아니다.

| Branch | 조건 | 진입 실측 | Test (a094 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:814` `switch resolution {` | — | 기존 | no | no |
| B2 | `:815` `case ProposalRefused:` | 예 | 기존 | no | no |
| B3 | `:816` `case ProposalCancelled:` | 예 | 기존 | no | no |
| B4 | `:818` `default:` | 아니오 | 기존 | no | no |
| B5 | `:825` `if err != nil {` | 아니오 | 기존 | no | no |
| B6 | `:836` `if errors.Is(err, sql.ErrNoRows) {` | 아니오 | 기존 | no | no |
| B7 | `:839` `if err != nil {` | 예 | 기존 | no | no |
| B8 | `:842` `if strings.TrimSpace(pendingAction.String) == "" {` | 예 | **a094 4.4** — 이미 해제된 발의에 대한 재호출은 무동작 | no | no |
| B9 | `:846` `if _, err := tx.ExecContext(ctx, `` | 예 | **a094 4.2** — NULL 쓰기가 해동의 실제 지점 | no | no |
| B10 | `:852` `if kind == ExitPolicyLadder {` | 예 | **a094 4.5** — LADDER의 rung 되돌림이 손절 가격을 바꾸지 않는다 | no | no |
| B11 | `:853` `if rung, err := exitpolicy.RungIndex(strings.TrimSpace(pendingLevel.String)); err == nil {` | 예 | **a094 4.5** — 음수 label은 되돌림 없이 통과 | no | no |
| B12 | `:854` `if err := rollBackRungTx(ctx, tx, id, rung-1, now); err != nil {` | 아니오 | 기존 | no | no |
| B13 | `:859` `if err := appendExitEventTx(ctx, tx, exitEventRow{` | 예 | 기존 | no | no |
| B14 | `:865` `if err := tx.Commit(); err != nil {` | 예 | 기존 | no | no |

**미진입 분기 4개**: B4, B5, B6, B12
**자체 블록 없는 분기 1개**: B1 — 컴파일러가 별도 블록을 만들지 않는 형태이며 미커버와 다르다.
