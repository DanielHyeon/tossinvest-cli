# Branch Test Map: `validateProductionFamilyActivation`

- Source: `internal/strategyrouter/production_family_activation.go` (522-584); file SHA-256 `c8e2efb9b1a8243bcec000f0c2fa9e96bd8576c97aa35a694a1e6c053407288f`. AST branch positions are authoritative.

- Measurement regime: Go coverage profiles, count mode. arm = 분기 좌표 **뒤에서 처음 시작하는** 커버리지 블록(`if`/`range` 몸통)이며, "arm entered Nx" 는 그 몸통이 N 번 실행됐다는 뜻이다.
- engine tagged suite (and the strategyrouter tagged suite for the two router bundles, same flags): `go test -c -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine,./internal/strategyrouter ./internal/app/engine/` 바이너리를 `systemd-run --user --scope -p MemoryMax=16G -p MemorySwapMax=0` 안에서 실행(-trimpath 없이 — 소스를 읽는 시험 둘이 깨진다). 스위트 전체 PASS.
- Per-test attribution set: 같은 바이너리의 **전체 시험 목록**을 `-test.run '^<Test>$'` 로 하나씩 돈 프로파일(하네스 `analysis/harness/a872_pertest_cover.sh`, 표 생성 `analysis/harness/a872_attribute.py`).
- **귀속 완전성은 등식이다**: 모든 행에서 시험별 진입 수의 합 == 스위트 진입 수. 깨진 행은 `ATTRIBUTION MISMATCH` 로 찍히며 아래에는 하나도 없다.
- 이 regime 은 **몸통 진입**을 센다. 같은 change 의 옛 번들 일부(예: dispatch 의 5.x 표)는 조건 평가를 센 값이라 수가 다르다 — 섞어 읽지 말 것.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 525:2 | arm entered 12x (strategyrouter tagged suite, post-edit); `TestAnActivationWhoseBindingsDoNotMatchPromotesNothing`, `TestTheOtherMarketsWholeManifestInThisMarketsFilePromotesNothing` |
| B2 | if | 539:2 | arm entered 4x (strategyrouter tagged suite, post-edit); `TestAnActivationOutsideItsApprovedLifetimePromotesNothing` |
| B3 | if | 544:2 | arm entered 4x (strategyrouter tagged suite, post-edit); `TestAnActivationOutsideItsApprovedLifetimePromotesNothing`, `TestLoadingAndLeasingJudgeExpiryAtTheSameInstant`, `TestOnlyAnEmptyPinMeansTheActivationWasNeverDeclared` |
| B4 | range | 562:2 | arm entered 53x (strategyrouter tagged suite, post-edit); `TestAVerifiedActivationCarriesTheProtectionFloorTheOrderPathMustBind`, `TestAVerifiedFourFamilyActivationPromotesExactlyTheLanesItNames`, `TestAnActivationPromotesOnlyTheFamiliesItTurnsOn`, `TestAnActivationWhoseDescriptorSetIsNotExactlyTheFourPromotesNothing`, `TestLoadingAndLeasingJudgeExpiryAtTheSameInstant`, `TestTheCommittedGoldenManifestMatchesItsPinAndPromotesTheFourLanes`, `TestTheLeaseCeilingOnlyEverShrinks` |
| B5 | if | 564:3 | arm entered 5x (strategyrouter tagged suite, post-edit); `TestAnActivationWhoseDescriptorSetIsNotExactlyTheFourPromotesNothing` |
| B6 | if | 571:3 | arm entered 1x (strategyrouter tagged suite, post-edit); `TestAnActivationWhoseDescriptorSetIsNotExactlyTheFourPromotesNothing` |
| B7 | if | 575:3 | arm entered 2x (strategyrouter tagged suite, post-edit); `TestAnActivationWhoseDescriptorSetIsNotExactlyTheFourPromotesNothing` |
| B8 | if | 580:2 | arm entered 1x (strategyrouter tagged suite, post-edit); `TestAnActivationWhoseDescriptorSetIsNotExactlyTheFourPromotesNothing` |


A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
