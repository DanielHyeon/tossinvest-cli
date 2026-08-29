# Branch Test Map: `Console.overview`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (무분기) 일곱 패널이 각자 조립되고 어느 하나의 실패가 나머지를 비우지 않으며 브로커 호출은 0이다 | `TestTheOverviewMakesNoBrokerCall`, `TestAnUnreadableJournalEmptiesOnlyItsOwnPanels`, `TestTheLedgerIsOpenedOncePerRenderAndTheNoticePrintsOnce`, `TestTheOverviewHasNoFormAndNoConfirmationInput` | — | yes |
