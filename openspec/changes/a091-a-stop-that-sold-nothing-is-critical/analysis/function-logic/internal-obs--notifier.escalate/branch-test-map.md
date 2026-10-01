# Branch Test Map: `Notifier.escalate`

- Source: `internal/obs/notifier.go`

> Test 열은 그 함수를 지나는 현존 · 통과 시험이다. 「아니오」 분기의 인용은 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).
> a091 의 새 RED 는 구현 로트의 변이 원장(`analysis/implementation/mutation-ledger.md`)이 잰다 — 이 표는 편집 뒤 `540aebe6` 의 사실이다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:426` `if n.Journal == nil \|\| strings.TrimSpace(n.AccountRef) == "" {` | 예 | `TestA092RecordOnlyFailureLatchesAndEscalates` | n/a | yes |
| B2 | `:431` `switch {` | — | `TestA091TheEscalationLinesCarryNoAccount` | n/a | yes |
| B3 | `:432` `case err != nil && n.Log != nil:` | 예 | `TestA091TheEscalationLinesCarryNoAccount` · `TestA092TheEscalationFailureErrorMasksTheAccount` | n/a | yes |
| B4 | `:438` `case changed && n.Log != nil:` | 예 | `TestA091TheEscalationLinesCarryNoAccount` | n/a | yes |
