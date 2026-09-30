# Branch Test Map: `RecordOnly.Notify`

- Source: `internal/obs/record_only.go`

> Test 열은 그 함수를 지나는 현존 · 통과 시험이다. 「아니오」 분기의 인용은 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).
> a091 의 새 RED 는 구현 로트의 변이 원장(`analysis/implementation/mutation-ledger.md`)이 잰다 — 이 표는 편집 뒤 `3ec1efd2` 의 사실이다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:47` `if n == nil {` | 예 | `TestA092RecordOnlyCriticalNeverPublishes` | n/a | yes |
| B2 | `:52` `if severity != SeverityCritical {` | 예 | `TestA092ANormalAlertIsHandedOffNotSent` | n/a | yes |
| B3 | `:53` `if r.Relay == nil {` | 예 | `TestA092WithoutARelayANormalAlertIsDroppedNotSent` | n/a | yes |
