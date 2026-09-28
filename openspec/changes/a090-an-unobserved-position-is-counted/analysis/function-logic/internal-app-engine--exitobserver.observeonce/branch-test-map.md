# Branch Test Map: `ExitObserver.ObserveOnce`

AST 분기 8 · return 5 · 무음 `continue` 2(`:457`·`:462`). 진입 실측은 `analysis/harness/observeonce.blocks`
(commit `b0a202b8`, 깨끗한 detached worktree, `go test ./internal/app/engine/ -count=1 -covermode=set`).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:417` 체결 감지가 밀리면 주기 전체 양보 + outage 시계 유지 | `TestTheCycleYieldsToFillDetection` (`exitloop_test.go:673`) | no | yes |
| B2 | `:427` `workingSet` 오류가 주기 오류가 된다 | **없음 — 블록 count 0**(a092 번들 18라운드 B-P7 와 같은 사실: 시험 원장은 관측 중 항상 살아 있다). a090 은 이 분기를 바꾸지 않는다 | no | no |
| B3 | `:431` 무보유 계정은 outage 로 승격되지 않는다 | `TestAnAccountHoldingNothingIsNotInAnOutage` (`exitloop_test.go:659`) | no | yes |
| B4 | `:442` **전 종목** 미응답이 계정 사다리를 탄다 | `TestASustainedOutageBlocksEntriesAndAlertsOnce` (`:608`) · `TestAQuoteWithNoLastTradeIsNotAnObservation` (`:585`) · `TestA111MissingManagedSymbolIsInvalidEvidenceAndDoesNotResetTheOutage` (`a111_flat_exit_observation_test.go:735`) | no | yes |
| B5 | `:451` 보유 포지션마다 1회 방문 | `TestTheLoopOpensTheExitStateOfANewlyHeldPosition` (`exitloop_test.go:447`) · `TestASuccessfulObservationStampsThePriceFreshness` (`:518`) | no | yes |
| B6 | `:453` 일부 종목만 미응답 → 그 종목은 **무음으로** 빠진다 | `TestA111ValidSiblingIsJudgedWithoutLendingFreshnessToInvalidSymbol` (`a111_flat_exit_observation_test.go:759`) — 형제 하나가 NaN 이면 그 종목은 Seed 로 남고 **진입 게이트는 열린 채**임을 단언한다(무음을 고정하는 쪽) | no | yes |
| B7 | `:459` 앞 포지션 처리 동안 임대가 끝난 시세 → **무음으로** 빠진다 | `TestA111SlowFirstPositionExpiresLaterQuoteWithoutAbandoningStartedProtection` (`a111_flat_exit_observation_test.go:833`) — 첫 포지션 제출이 16초 걸리면 뒤 종목은 판정·원장 효과 0 | no | yes |
| B8 | `:465` 첫 `judge` 오류만 주기에 실린다 | `TestAFailedObservationHoldsTheJudgement` (`exitloop_test.go:560`) | no | yes |

## a092 번들 이후 바뀐 것

a092 의 BTM(7 분기판)은 "B6 — **없음, 이 change 의 대상**" 이라고 적었다. 지금은 **진입한다** — a111(`882a0b49`)이 B6 에 닿는
시험(:759)과 새 B7 과 그 시험(:833)을 더했다. 다만 두 시험은 **무음이 옳다고** 단언한다(격리·원장 효과 0·게이트 열림).
**미관측이 셈해지고 보고되는지는 여전히 어느 시험도 묻지 않는다** — a090 의 RED 가 그 자리다.

## 필요한 RED (a092 FLM R1~R6 승계 + a090 추가)

| # | 출처 | Scenario | 기대 |
|---|---|---|---|
| R1 | a092 R1 | 보유 2종목, 1종목만 `Last = 0` | `cycle.Err == nil`, 다른 종목은 정상 판정, 빠진 종목이 **미관측으로 기록된다**(시작 시각) |
| R2 | a092 R2 | 보유 2종목, 1종목이 응답에 **부재** | R1 과 같다 — 원인이 0가격이든 부재든 결과가 같다 |
| R3 | a092 R3(개정) | 같은 포지션이 임계 이상 연속 미관측 | 임계(design D3) 초과 시 **포지션 key 의 critical 1회**, 다음 주기에 반복 없음 |
| R4 | a092 R4 | 미관측 포지션이 다음 주기에 판정에 닿는다 | 기록 해제 — 다시 빠지면 새로 센다 |
| R5 | a092 R5 | B4(전 종목 미응답) | **무변화** — 계정 사다리 그대로, 포지션 단위 경보는 B4 경로에서 나지 않는다 |
| R6 | a092 R6 | B1(양보) | **무변화** — `checkOutage` 가 계속 돈다 |
| R7 | a090 | B7(임대 만료)로 빠진 포지션 | R1 과 같이 기록된다(원인 필드만 다름) — 임계 전 한 번의 B7 은 경보가 **아니다** |
| R8 | a090 | 미관측 중 포지션이 보유에서 사라진다 | 기록이 정리되고 경보가 나지 않는다 |
| R9 | a090 | 기존 시험 :759 · :833 | **무변화로 통과**한다(한 주기 만에는 경보·게이트 변화 없음) |

R1·R2·R3 이 RED 로 실패하는 것이 결함의 존재 증명이다. R5·R6·R9 는 회귀 방지.
