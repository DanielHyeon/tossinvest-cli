# Branch Test Map: Assess

- Source: `internal/protectionreadiness/readiness.go` (103-170); current — HEAD `648df8ef`
- 시험 칸은 측정값: 무태그 `go test -c -cover` 바이너리를 시험 함수마다 따로 돌린 커버 프로필에서 그 분기 본문 블록을 실행한 시험(`analysis/harness/51_matrix.sh`). `_test.go` 함수와 base 함수는 계측 밖이라 그 사실을 행에 적음.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 109:2 | `if stateValid {` | `FuzzArbitraryAttestationNeverWires` · `FuzzSerialMustStrictlyIncrease` (+21) | 5.1 에서 재실행 안 함 | 블록 109.16-111.3 을 시험 23개가 실행, 전부 PASS |
| B2 | if at 112:2 | `if stateValid && timeValid && !timeRollback {` | `FuzzArbitraryAttestationNeverWires` · `FuzzSerialMustStrictlyIncrease` (+21) | 5.1 에서 재실행 안 함 | 블록 112.46-114.108 을 시험 23개가 실행, 전부 PASS |
| B3 | if at 114:3 | `if result.NextState.TrustedTimeFloor.IsZero() \|\| input.Time.Now.After(result.NextState.TrustedTimeFloor) {` | `FuzzArbitraryAttestationNeverWires` · `TestBrokerIdentityCapabilityIsAllOrNothing` (+20) | 5.1 에서 재실행 안 함 | 블록 114.108-116.4 을 시험 22개가 실행, 전부 PASS |
| B4 | range at 118:2 | `for _, market := range []Market{MarketKR, MarketUS} {` | `FuzzArbitraryAttestationNeverWires` · `FuzzSerialMustStrictlyIncrease` (+22) | 5.1 에서 재실행 안 함 | 블록 118.54-120.15 을 시험 24개가 실행, 전부 PASS |
| B5 | if at 120:3 | `if !present {` | `FuzzArbitraryAttestationNeverWires` · `FuzzSerialMustStrictlyIncrease` (+16) | 5.1 에서 재실행 안 함 | 블록 120.15-121.12 을 시험 18개가 실행, 전부 PASS |
| B6 | switch at 124:3 | `switch {` | `FuzzArbitraryAttestationNeverWires` · `FuzzSerialMustStrictlyIncrease` (+22) | 5.1 에서 재실행 안 함 | 블록 123.3-124.10 을 시험 24개가 실행, 전부 PASS |
| B7 | case at 125:3 | `case !policyValid:` | 없음 | 5.1 에서 재실행 안 함 | 블록 125.21-126.33 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B8 | case at 127:3 | `case !stateValid:` | `TestCorruptDurableStateIsPreservedAndNeverAutoRepaired` | 5.1 에서 재실행 안 함 | 블록 127.20-128.38 을 시험 1개가 실행, 전부 PASS |
| B9 | case at 129:3 | `case !timeValid:` | `TestSerialAndTrustedTimeStateAreMonotonicAndPure` | 5.1 에서 재실행 안 함 | 블록 129.19-130.48 을 시험 1개가 실행, 전부 PASS |
| B10 | case at 131:3 | `case timeRollback:` | `TestSerialAndTrustedTimeStateAreMonotonicAndPure` · `TestTrustedTimeFloorAdvancesWithoutAcceptedEvidenceAndDetectsLaterRollback` | 5.1 에서 재실행 안 함 | 블록 131.21-132.45 을 시험 2개가 실행, 전부 PASS |
| B11 | case at 133:3 | `default:` | `FuzzArbitraryAttestationNeverWires` · `FuzzSerialMustStrictlyIncrease` (+21) | 5.1 에서 재실행 안 함 | 블록 133.11-136.27 을 시험 23개가 실행, 전부 PASS |
| B12 | if at 136:4 | `if code == RefusalNone {` | `FuzzSerialMustStrictlyIncrease` · `TestCachedProviderRejectsNewerDurableSerialFromPeer` (+14) | 5.1 에서 재실행 안 함 | 블록 136.27-148.5 을 시험 16개가 실행, 전부 PASS |
| B13 | if at 150:3 | `if market == MarketKR {` | `FuzzArbitraryAttestationNeverWires` · `FuzzSerialMustStrictlyIncrease` (+21) | 5.1 에서 재실행 안 함 | 블록 150.25-152.4 을 시험 23개가 실행, 전부 PASS |
| B14 | else at 152:10 | `} else {` | `TestCachedProviderRejectsNewerDurableSerialFromPeer` · `TestCorruptKRMarketSnapshotDoesNotInvalidateSealedUSMarket` (+8) | 5.1 에서 재실행 안 함 | 블록 152.9-154.4 을 시험 10개가 실행, 전부 PASS |
| B15 | if at 156:2 | `if result.StateCommitAllowed {` | `FuzzArbitraryAttestationNeverWires` · `FuzzSerialMustStrictlyIncrease` (+21) | 5.1 에서 재실행 안 함 | 블록 156.31-158.48 을 시험 23개가 실행, 전부 PASS |
| B16 | else at 161:9 | `} else {` | `TestCorruptDurableStateIsPreservedAndNeverAutoRepaired` · `TestSerialAndTrustedTimeStateAreMonotonicAndPure` (+1) | 5.1 에서 재실행 안 함 | 블록 161.8-165.3 을 시험 3개가 실행, 전부 PASS |
| B17 | if at 158:3 | `if result.NextState.seal != input.State.seal {` | `FuzzArbitraryAttestationNeverWires` · `FuzzSerialMustStrictlyIncrease` (+21) | 5.1 에서 재실행 안 함 | 블록 158.48-160.4 을 시험 23개가 실행, 전부 PASS |
