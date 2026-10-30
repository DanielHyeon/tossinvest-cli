# 분기·시험 대응

| 경로 | 선택 근거 |
|---|---|
| a130 required + 누락 manifest | TestA166OutcomeRequiredMissingManifest; baseline RED /tmp/a166-outcome-red.log |
| a130 lots 0 및 Close 오류 | TestA166OutcomeEmptyAndCloseFailure; 기존 noNetwork fixture |
| a130 OFF 실제 token + candle, 두 호출 auxiliary | TestA166OutcomeClosureLocalToken; auxiliary→normal 변이 RED |
| a130 required busy | TestA166OutcomeBusyAuxiliaryNoWire, wire 0 |
| d2 required + 누락 manifest | TestA166D2RequiredMissingManifest; 고정 TLS baseline RED·GREEN |
| d2 dry-run | TestA166D2DryRunNoBudgetOpen |
| d2 OFF token, auxiliary와 5분 deadline | TestA166D2OffLocalToken; auxiliary→normal 변이 RED |
| d2 required busy | TestA166D2BusyAuxiliaryNoWire, wire 0 |
| 기존 preview / 종목 선택 | TestThePreviewPrintsATokenAndAWrongTokenWritesNothing / TestListSymbols |

나머지 기존 분기는 변경 전 AST 및 손읽기로 보존을 확인했으며 이번 선택 시험에서 개별 실행했다고 주장하지 않는다. d2 Close 오류 주입과 a130 원장 lots가 있는 실제 preview 통합은 이번 선택 범위에서 직접 시험하지 않았다.
