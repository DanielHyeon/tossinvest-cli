# Branch Test Map: `ExitObserver.alert`

- Source: `internal/app/engine/exitloop.go`

> Test 열은 그 함수를 지나는 현존 · 통과 시험이다. 「아니오」 분기의 인용은 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).
> a091 의 새 RED 는 구현 로트의 변이 원장이 잰다 — 이 표는 편집 **전** base 의 사실이다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:1815` `if o.opts.Alerts == nil {` | 예 | `TestNoFloorSourceCapsNothing` | n/a | yes |
| B2 | `:1818` `if err := o.opts.Alerts.Notify(ctx, e); err != nil {` | 예 | `TestA092ACappedLiquidationDoesNotWaitForTheTransport` | n/a | yes |
