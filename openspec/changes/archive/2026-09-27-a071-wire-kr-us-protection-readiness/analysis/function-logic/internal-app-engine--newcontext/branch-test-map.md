# Branch Test Map: NewContext

- Source: `internal/app/engine/engine.go` (432-602); current — HEAD `648df8ef`
- 시험 칸은 측정값: 무태그 `go test -c -cover` 바이너리를 시험 함수마다 따로 돌린 커버 프로필에서 그 분기 본문 블록을 실행한 시험(`analysis/harness/51_matrix.sh`). `_test.go` 함수와 base 함수는 계측 밖이라 그 사실을 행에 적음.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 434:2 | `if err != nil {` | `TestNewRefusesIncompleteCredentials` · `TestNewRefusesWithoutCredentials` (+1) | 5.1 에서 재실행 안 함 | 블록 434.16-436.3 을 시험 3개가 실행, 전부 PASS |
| B2 | if at 441:2 | `if err != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 441.16-443.3 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B3 | if at 445:2 | `if clk == nil {` | `TestProductionEngineAssemblesPairedUnwiredReadinessProvider` · `TestProtectionStorageFailureCannotPreventSafetyRuntimeAssembly` (+24) | 5.1 에서 재실행 안 함 | 블록 445.16-447.3 을 시험 26개가 실행, 전부 PASS |
| B4 | if at 459:2 | `if opts.Publisher != nil {` | `TestAWiredPublisherDeliversTheCriticalAlert` · `TestAnInjectedPublisherWins` | 5.1 에서 재실행 안 함 | 블록 459.27-461.3 을 시험 2개가 실행, 전부 PASS |
| B5 | if at 467:2 | `if err := recordGateSettings(auditLog, gate, cfg.Engine.Adoption, notifications,` | 없음 | 5.1 에서 재실행 안 함 | 블록 468.39-473.3 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B6 | if at 482:2 | `if err != nil {` | `TestStartupRefusesWhenTheAccountCannotBeResolved` | 5.1 에서 재실행 안 함 | 블록 482.16-484.3 을 시험 1개가 실행, 전부 PASS |
| B7 | if at 489:2 | `if err != nil {` | `TestStartupRefusesOnADisallowedFilesystem` | 5.1 에서 재실행 안 함 | 블록 489.16-491.3 을 시험 1개가 실행, 전부 PASS |
| B8 | if at 498:2 | `if err := bindApplyHooks(jrn); err != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 498.44-501.3 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B9 | if at 517:2 | `if err != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 517.16-520.3 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B10 | if at 532:2 | `if gate.Enabled && guardian == nil && !opts.disableProductionGuardian {` | `TestProductionGuardianConstructionFailureClosesTheEngineJournal` · `TestProductionGuardianUsesConfiguredUSDLimitsAndReachesExitObserver` | 5.1 에서 재실행 안 함 | 블록 532.72-534.21 을 시험 2개가 실행, 전부 PASS |
| B11 | if at 534:3 | `if factory == nil {` | `TestProductionGuardianUsesConfiguredUSDLimitsAndReachesExitObserver` | 5.1 에서 재실행 안 함 | 블록 534.21-536.4 을 시험 1개가 실행, 전부 PASS |
| B12 | if at 540:3 | `if err != nil {` | `TestProductionGuardianConstructionFailureClosesTheEngineJournal` | 5.1 에서 재실행 안 함 | 블록 540.17-544.4 을 시험 1개가 실행, 전부 PASS |
| B13 | if at 560:2 | `if err != nil {` | `TestAGuardianSizedAgainstOtherNumbersIsRefused` · `TestEachInterlockClauseHasALine` (+4) | 5.1 에서 재실행 안 함 | 블록 560.16-563.3 을 시험 6개가 실행, 전부 PASS |
| B14 | if at 564:2 | `if !automation.Verified {` | `TestProductionEngineAssemblesPairedUnwiredReadinessProvider` · `TestProtectionStorageFailureCannotPreventSafetyRuntimeAssembly` (+46) | 5.1 에서 재실행 안 함 | 블록 564.26-566.3 을 시험 48개가 실행, 전부 PASS |
| B15 | if at 569:2 | `if err != nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 569.16-572.3 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
