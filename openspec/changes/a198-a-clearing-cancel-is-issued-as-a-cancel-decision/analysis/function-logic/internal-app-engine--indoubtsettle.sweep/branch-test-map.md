# Branch Test Map: `inDoubtSettle.sweep` (pre-edit 번호)

시험 이름은 grep 근거(internal/app/engine/a193_in_doubt_settle_test.go). RED/GREEN 은 4c.6·4c.8 원문에서 채움.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 원장 읽기 실패 | TestA193AnUnreadableLedgerAlertsAfterTenMinutes | n/a(기존) | 4c.8 |
| B2 | 읽기 실패 + ctx 취소 | TestA193AShutdownMidResolutionIsAGracefulAuxiliaryStop | n/a(기존) | 4c.8 |
| B3 | 순회 | TestA193OnlyInDoubtPlaceAttemptsAreTouched | n/a(기존) | 4c.8 |
| B4 | IN_DOUBT PLACE 만 대상 | TestA193OnlyInDoubtPlaceAttemptsAreTouched | n/a(기존) | 4c.8 |
| B5 | 1단 알림 | TestA193FirstSightIsAlertedOnce | n/a(기존) | 4c.8 |
| B6 | 2단 직렬 해소 | TestA193AttemptsAreResolvedOneAtATime | n/a(기존) | 4c.8 |
| B7 | 해소 중 ctx 취소 | TestA193AShutdownMidResolutionIsAGracefulAuxiliaryStop | n/a(기존) | 4c.8 |
