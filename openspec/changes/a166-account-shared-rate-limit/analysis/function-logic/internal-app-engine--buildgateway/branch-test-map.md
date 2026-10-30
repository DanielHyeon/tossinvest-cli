# Branch Test Map: `buildGateway(params=2, results=2)`

OFF HTTPReplay 기존 형태, required는 동일 borrowed budget과 official에 결속한 독립 transport. 실패 반환, attestation nil 유지. 시험 TestA166GatewayReplayBinding.

| 분기 | AST 위치와 조건 원문 | 결과 범위 |
|---|---|---|
| B1 | if L255: `if err := checkProjectionWired(in.journal); err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B2 | if L277: `if err := tracker.Restore(ctx); err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B3 | if L285: `if err := restoreAlertEntryLatch(ctx, in.journal, entry); err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B4 | if L289: `if err := bindOperatingModeProjection(ctx, in.journal, entry, in.accountRef, in.logger); err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B5 | if L313: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B6 | if L317: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B7 | if L338: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
