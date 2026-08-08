# Branch Test Map: `armExitProposalTx`

- Source: `internal/journal/apply_hook.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다. 따라서 「Test」 열은 **a094가 요구하는 시험**이며 현존 증명이 아니다.

| Branch | 조건 | 진입 실측 | Test (a094 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:660` `if errors.Is(err, sql.ErrNoRows) {` | 아니오 | 기존 | no | no |
| B2 | `:663` `if err != nil {` | 예 | 기존 | no | no |
| B3 | `:666` `if strings.TrimSpace(action.String) != "" {` | 예 | **a094 3.D1** — 미결 발의 위의 두 번째는 거부된다(R2 축소의 전제) | no | no |
| B4 | `:669` `if _, err := tx.ExecContext(ctx, `` | 예 | 기존 | no | no |

**미진입 분기 1개**: B1
**자체 블록 없는 분기 0개**: 없음 — 컴파일러가 별도 블록을 만들지 않는 형태이며 미커버와 다르다.
