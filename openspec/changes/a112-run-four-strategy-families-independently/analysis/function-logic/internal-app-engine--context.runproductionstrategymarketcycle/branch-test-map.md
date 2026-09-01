# Branch Test Map: `Context.runProductionStrategyMarketCycle`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (422-459); file SHA-256 `4e457c677157b2f8c73f813f8250575657b6beedddc1ad467db209a35579986d`. AST branch positions are authoritative.
- 이 태스크가 B2 를 편집했다. **어떤 시험도 이 함수에 닿지 않는다** — 아래는 그 사실과 구조적 반증이다.
| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 424:2 — refresh 실패 | 없음 | 아니오 | 아니오 — **진입 0** |
| B2 | if at 427:2 — dispatch 부재(handoff 거절은 경계 안으로 갔다) | 없음 (단위 수준은 `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff` 가 같은 판단을 고정한다) | 아니오 | 아니오 — **진입 0** |
| B3 | if at 447:3 — 캠페인 CAS 실패 | 없음 | 아니오 | 아니오 — **진입 0** |
| B4 | if at 450:3 — 이미 점유된 캠페인 | 없음 | 아니오 | 아니오 — **진입 0** |
| B5 | if at 454:3 — lease 소모 | 없음 | 아니오 | 아니오 — **진입 0** |

## 반증 실측

| 뮤테이션 | 결과 |
|---|---|
| M3: `dispatch` 를 메서드 값으로 꺼내 둔다(`send := fresh.dispatch.dispatch`) | KILLED — `TestExactlyOneProductionCallSiteTurnsAHandoffIntoADispatch` 실패(철자 census). 5.5 판본은 호출식만 세었으므로 이 형태를 놓쳤다. |
| M4: 경계가 묶어 주지 않은 이름을 dispatch 에 넘긴다 | KILLED — 같은 시험 실패 |
| P1b-B1: `Deliver` 를 `Single()` 로 되돌리고 관문을 빈 `if !handedOff { }` 로 만든다 | KILLED — `TestNoProductionSiteDiscardsTheSeamsAdmissionAnswer` 실패(경계의 답을 묶는 자리가 3 이 아니라 4 가 된다) |
| P1b-B3: 몸통 안에서 값을 `entries` 의 마지막 것으로 덮어쓴다 | KILLED — 이제 **컴파일되지 않는다**. 몸통은 `strategyhandoff.Delivered` 를 받고 `dispatch` 도 그것만 받으므로, `strategyflow.Result` 를 넘기는 편집은 타입에서 멈춘다(뮤테이션 M1) |
| P1c: 엔진이 스스로 `Admit` 을 불러 봉투를 만들어 넘긴다 | KILLED — `TestExactlyOneProductionSiteAdmitsIntoTheSeam` 실패(뮤테이션 M2) |
| P1d: 같은 봉투로 `dispatch` 를 세 번 부른다 | SURVIVED — 소스 검사는 철자를 셀 뿐 실행을 세지 않는다. 막는 것은 원장의 position campaign CAS 이고, 그 사실은 `TestTheSameEnvelopeCannotPlaceASecondOrder` 가 값으로 잰다(뮤테이션 M3) |

이 함수의 동작 커버리지는 0 이다. 그 사실을 적지 않고 넘어가면 구조 가드가 동작 증거처럼 읽힌다.
