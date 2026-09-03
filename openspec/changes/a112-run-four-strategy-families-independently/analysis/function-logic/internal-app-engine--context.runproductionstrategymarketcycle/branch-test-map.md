# Branch Test Map: `Context.runProductionStrategyMarketCycle`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (454-521); file SHA-256 `4eede127fbbec4233391d783660d1bca000e8d85ba61b02d394b5776840f4e50`. AST branch positions are authoritative.
- L5 5.3.3 이 이 본문에 두 갈래(B2·B3)를 더했다. **어떤 시험도 이 함수를 통째로 돌지 않는다** — 아래는 그 사실과 구조적 반증이다.
| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 456:2 — refresh 실패 | 없음 | 아니오 | 아니오 — **진입 0** |
| B2 | if at 480:2 — **durable latch 를 읽으며 레인을 세우지 못했다**(5.3.3 이 더함) | 이 함수로는 없음. 같은 판단을 `TestADurableLatchThatNamesNoLaneInThisBuildStopsTheCycleLoudlyAndCanBeClosed` 가 레인 런타임에서 값으로 잰다 | 아니오 | 아니오 — **이 함수 진입 0** |
| B3 | if at 483:2 — **레인 주기가 durable latch 를 남기지 못했다**(5.3.3 이 더함) | 이 함수로는 없음. 같은 판단을 `TestALedgerThatCannotTakeTheLatchStopsTheCycle` 이 값으로 잰다 | 아니오 | 아니오 — **이 함수 진입 0** |
| B4 | if at 489:2 — dispatch 부재(handoff 거절은 경계 안으로 갔다) | 없음 (단위 수준은 `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff` 가 같은 판단을 고정한다) | 아니오 | 아니오 — **진입 0** |
| B5 | if at 509:3 — 캠페인 CAS 실패 | 없음 | 아니오 | 아니오 — **진입 0** |
| B6 | if at 512:3 — 이미 점유된 캠페인 | 없음 | 아니오 | 아니오 — **진입 0** |
| B7 | if at 516:3 — lease 소모 | 없음 | 아니오 | 아니오 — **진입 0** |

## 반증 실측

| 뮤테이션 | 결과 |
|---|---|
| M3: `dispatch` 를 메서드 값으로 꺼내 둔다(`send := fresh.dispatch.dispatch`) | KILLED — `TestExactlyOneProductionCallSiteTurnsAHandoffIntoADispatch` 실패(철자 census). 5.5 판본은 호출식만 세었으므로 이 형태를 놓쳤다. |
| M4: 경계가 묶어 주지 않은 이름을 dispatch 에 넘긴다 | KILLED — 같은 시험 실패 |
| P1b-B1: `Deliver` 를 `Single()` 로 되돌리고 관문을 빈 `if !handedOff { }` 로 만든다 | KILLED — `TestNoProductionSiteDiscardsTheSeamsAdmissionAnswer` 실패(경계의 답을 묶는 자리가 3 이 아니라 4 가 된다) |
| P1b-B3: 몸통 안에서 값을 `entries` 의 마지막 것으로 덮어쓴다 | KILLED — 이제 **컴파일되지 않는다**. 몸통은 `strategyhandoff.Delivered` 를 받고 `dispatch` 도 그것만 받으므로, `strategyflow.Result` 를 넘기는 편집은 타입에서 멈춘다(뮤테이션 M1) |
| P1c: 엔진이 스스로 `Admit` 을 불러 봉투를 만들어 넘긴다 | KILLED — `TestExactlyOneProductionSiteAdmitsIntoTheSeam` 실패(뮤테이션 M2) |
| P1d: 같은 봉투로 `dispatch` 를 세 번 부른다 | SURVIVED — 소스 검사는 철자를 셀 뿐 실행을 세지 않는다. 막는 것은 원장의 position campaign CAS 이고, 그 사실은 `TestTheSameEnvelopeCannotPlaceASecondOrder` 가 값으로 잰다(뮤테이션 M3) |
| 5.3.3 N11: 복구 세대에 서명과 무관한 `0` 을 넘긴다(`456:3`) | KILLED — `TestTheRecoveryGenerationComesFromTheSignedActivationAndNothingElse` 실패. 행동 시험으로는 못 잡는다: 잘못된 수도 그냥 커지므로 잠금이 저절로 열려도 어떤 값 비교도 달라지지 않는다 |

이 함수의 동작 커버리지는 0 이다. 그 사실을 적지 않고 넘어가면 구조 가드가 동작 증거처럼 읽힌다.
