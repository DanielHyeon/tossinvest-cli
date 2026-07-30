# Branch Test Map: `Console.positions`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (무분기) 브로커·원장 두 절반을 조인해 화면을 만들고, 어느 한쪽이 없어도 렌더한다 | `TestThePositionsScreenShowsTheExitLineOfAManagedPosition`, `TestThePositionsScreenRendersWithEitherSourceMissing`, `TestAnUnmanagedHoldingIsLabelledExactlyOnce`, `TestAVerificationInProgressSuspendsTheRefresh` | — | yes |
