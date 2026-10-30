# Branch Test Map: `ExitObserver.submit(params=7, results=1)`

보호 Place는 critical, 익절 Place는 normal. floor·Guardian·attach 실패와 결과 분류 유지. 시험 TestA166ExitProposalPriority, TestA166FullTakeProfitCancelIsNormal.

| 분기 | AST 위치와 조건 원문 | 결과 범위 |
|---|---|---|
| B1 | if L1496: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B2 | if L1503: `if isZeroQuantity(submitQuantity) {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B3 | if L1525: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B4 | if L1536: `if err := o.opts.Journal.AttachExitIntent(ctx, m.position.ID, intentID); err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B5 | if L1542: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B6 | if L1549: `if isProtective(proposal) {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B7 | if L1558: `if isProtective(proposal) {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B8 | switch L1563: `switch {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B9 | case L1564: `case out.State == journal.StateConfirmed:` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B10 | case L1573: `case out.State == journal.StateInDoubt \|\| out.State == journal.StateUnresolvedInDoubt:` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B11 | if L1578: `if out.State == journal.StateUnresolvedInDoubt {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B12 | case L1583: `case out.Reason == execgw.ReasonSymbolInFlight:` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B13 | case L1588: `case out.AttemptID != "" && out.State != journal.StateNotDispatched && out.State != journal.StateFailedConfirmed:` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B14 | if L1592: `if err == nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B15 | case L1598: `default:` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B16 | if L1600: `if detail == "" && err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B17 | if L1605: `if out.State == journal.StateFailedConfirmed {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
