# Function Logic Map: `validateProductionFamilyActivation`

- Source: `internal/strategyrouter/production_family_activation.go` (522-584)
- Function: `validateProductionFamilyActivation` in package `strategyrouter`
- File SHA-256: `c8e2efb9b1a8243bcec000f0c2fa9e96bd8576c97aa35a694a1e6c053407288f`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file (post-edit).
- AST evidence: `ast.json` — AST branches 8.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

**편집 뒤 (태스크 8.7.2, 2026-09-27).** 편집 전 번들은 커밋 `c1d1e295`(validate 는 `cb378a63`) 에 있다 — 이 파일은 GREEN 뒤 현재 소스의 AST 와 측정이다.

편집: 만료 판정(B3)을 `familyActivationRemaining(expires, now)` 호출로 바꿨다 — 같은 함수를 `FamilyActivation.LeaseCeiling`
이 쓰므로 패키지의 만료 규칙이 한 곳이다. 경계 의미(만료 시각에 닿은 순간부터 만료)는 그대로.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. arm = 분기 좌표 **뒤에서 처음 시작하는** 커버리지 블록(`if`/`range` 몸통)이며, "arm entered Nx" 는 그 몸통이 N 번 실행됐다는 뜻이다.
- engine tagged suite (and the strategyrouter tagged suite for the two router bundles, same flags): `go test -c -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine,./internal/strategyrouter ./internal/app/engine/` 바이너리를 `systemd-run --user --scope -p MemoryMax=16G -p MemorySwapMax=0` 안에서 실행(-trimpath 없이 — 소스를 읽는 시험 둘이 깨진다). 스위트 전체 PASS.
- Per-test attribution set: 같은 바이너리의 **전체 시험 목록**을 `-test.run '^<Test>$'` 로 하나씩 돈 프로파일(하네스 `analysis/harness/a872_pertest_cover.sh`, 표 생성 `analysis/harness/a872_attribute.py`).
- **귀속 완전성은 등식이다**: 모든 행에서 시험별 진입 수의 합 == 스위트 진입 수. 깨진 행은 `ATTRIBUTION MISMATCH` 로 찍히며 아래에는 하나도 없다.
- 이 regime 은 **몸통 진입**을 센다. 같은 change 의 옛 번들 일부(예: dispatch 의 5.x 표)는 조건 평가를 센 값이라 수가 다르다 — 섞어 읽지 말 것.

Exact AST return positions: 533:3, 541:3, 545:3, 567:4, 572:4, 576:4, 581:3, 583:2.

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

엔진 스위트에서의 진입(전체 509 개 per-test):

- 엔진 B1: arm not entered (engine tagged suite, post-edit); no per-test profile entered it
- 엔진 B2: arm not entered (engine tagged suite, post-edit); no per-test profile entered it
- 엔진 B3: arm entered 2x (engine tagged suite, post-edit); `TestADeclaredActivationThatLapsesRollsItsMarketBackInsteadOfWidening`
- 엔진 B4: arm entered 8x (engine tagged suite, post-edit); `TestADeclaredActivationThatLapsesRollsItsMarketBackInsteadOfWidening`
- 엔진 B5: arm not entered (engine tagged suite, post-edit); no per-test profile entered it
- 엔진 B6: arm not entered (engine tagged suite, post-edit); no per-test profile entered it
- 엔진 B7: arm not entered (engine tagged suite, post-edit); no per-test profile entered it
- 엔진 B8: arm not entered (engine tagged suite, post-edit); no per-test profile entered it

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `validMarket` | 526:60 |
| `productionRouteIdentity` | 532:4 |
| `productionRouteTime` | 535:26 |
| `productionRouteTime` | 536:22 |
| `productionRouteTime` | 537:24 |
| `config.ObservedAt.UTC` | 538:9 |
| `issued.Before` | 539:47 |
| `issued.After` | 539:74 |
| `issued.Before` | 540:4 |
| `expires.Sub` | 540:30 |
| `familyActivationRemaining` | 544:15 |
| `productionRouteDescriptors` | 553:10 |
| `make` | 561:11 |
| `len` | 561:72 |
| `validDesiredState` | 566:5 |
| `validDesiredState` | 566:47 |
| `len` | 580:5 |
| `len` | 580:19 |

## State mutations and fallbacks

없음. 서술자 map 을 만들어 돌려준다.

## Safety conclusion

- `TestLoadingAndLeasingJudgeExpiryAtTheSameInstant` 가 경계 세 점에서 적재와 lease 가 같은 답을 내는지 잰다.
- `TestExpiryIsJudgedInExactlyOnePlace` 가 만료 오류를 **내는** 자리를 비시험 파일 전체에서 세어 하나(`familyActivationRemaining`)로 못 박는다 — 동치 사본 변이 M7 CAUGHT.
- High-risk impact: yes.
