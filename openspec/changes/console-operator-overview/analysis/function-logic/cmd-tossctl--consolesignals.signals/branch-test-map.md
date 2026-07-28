# Branch Test Map: `consoleSignals.Signals`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 저장소 열기 실패 → 화면이 사유 문장을 받는다 | `TestTheSignalsSeamOpensTheStoreUnderTheCallersContext` (취소된 context) | yes (`context.Background()` 사용 시 요청 취소가 전파되지 않음) | yes |
| B2 | KR·US 두 시장이 같은 instant로 평가되어 둘 다 목록에 존재 | `TestTheSignalsSeamReadsTheStoreAndCallsNoSource` | yes | yes |
