# Branch Test Map: `ExitObserver.ObserveOnce`

AST 분기 8 · return 5 · 무음 `continue` 2(`:457`·`:462`). 진입 실측은 `analysis/harness/observeonce.blocks`
(commit `eac13df1`, 깨끗한 detached worktree, `go test ./internal/app/engine/ -count=1 -covermode=set`).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:417` 체결 감지가 밀리면 주기 전체 양보 + outage 시계 유지 | `TestTheCycleYieldsToFillDetection` (`exitloop_test.go:673`) | no | yes |
| B2 | `:427` `workingSet` 오류가 주기 오류가 된다 | **없음 — 블록 count 0**(a092 번들 18라운드 B-P7 와 같은 사실: 시험 원장은 관측 중 항상 살아 있다). a090 은 이 분기를 바꾸지 않는다 | no | no |
| B3 | `:431` 무보유 계정은 outage 로 승격되지 않는다 | `TestAnAccountHoldingNothingIsNotInAnOutage` (`exitloop_test.go:659`) | no | yes |
| B4 | `:442` **전 종목** 미응답이 계정 사다리를 탄다 | `TestASustainedOutageBlocksEntriesAndAlertsOnce` (`:608`) · `TestAQuoteWithNoLastTradeIsNotAnObservation` (`:585`) · `TestA111MissingManagedSymbolIsInvalidEvidenceAndDoesNotResetTheOutage` (`a111_flat_exit_observation_test.go:735`) | no | yes |
| B5 | `:451` 보유 포지션마다 1회 방문 | `TestTheLoopOpensTheExitStateOfANewlyHeldPosition` (`exitloop_test.go:447`) · `TestASuccessfulObservationStampsThePriceFreshness` (`:518`) | no | yes |
| B6 | `:453` 일부 종목만 미응답 → 그 종목은 **무음으로** 빠진다 | `TestA111ValidSiblingIsJudgedWithoutLendingFreshnessToInvalidSymbol` (`a111_flat_exit_observation_test.go:759`) — 형제 하나가 NaN 이면 그 종목은 Seed 로 남고 **진입 게이트는 열린 채**임을 단언한다(무음을 고정하는 쪽) | no | yes |
| B7 | `:459` 앞 포지션 처리 동안 임대가 끝난 시세 → **무음으로** 빠진다 | `TestA111SlowFirstPositionExpiresLaterQuoteWithoutAbandoningStartedProtection` (`a111_flat_exit_observation_test.go:833`) — 첫 포지션 제출이 16초 걸리면 뒤 종목은 판정·원장 효과 0 | no | yes |
| B8 | `:465` 첫 `judge` 오류만 주기에 실린다 | **귀속 미측정** — 블록은 진입(패키지 커버리지)했으나 어느 시험인지 재지 않았다. 1·2판의 `TestAFailedObservationHoldsTheJudgement`(`exitloop_test.go:560`) 귀속은 **틀렸다** — 그 시험은 가격 읽기 실패를 주입해 B4 로 돌아간다(codex 1라운드 N9) | no | yes(블록 진입만) |

## a092 번들 이후 바뀐 것

a092 의 BTM(7 분기판)은 "B6 — **없음, 이 change 의 대상**" 이라고 적었다. 지금은 **진입한다** — a111(`882a0b49`)이 B6 에 닿는
시험(:759)과 새 B7 과 그 시험(:833)을 더했다. 다만 두 시험은 **무음이 옳다고** 단언한다(격리·원장 효과 0·게이트 열림).
**미관측이 셈해지고 보고되는지는 여전히 어느 시험도 묻지 않는다** — a090 의 RED 가 그 자리다.

## 필요한 RED

2판 목록은 `tasks.md` §2 가 정본이다(a092 R1~R6 승계 + a090 R7~R14). 이 표의 분기와의 대응: B6 → R1·R2 · B7 → R7 · 순회 뒤 → R3·R3a~e·R4·R8 ·
B3 → R12(보유 대상 전부 탈락) · B4 → R5 · B1 → R6 · B6·B7 기존 시험 → R9.
