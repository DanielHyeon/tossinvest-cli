# Branch Test Map: `Context.runProductionStrategyMarketCycle`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (422-457); file SHA-256 `1c2432d0f49db59209fc147f57a0c68d30d15596e68642aff8356ea29b0d69d5`. AST branch positions are authoritative.
- 이 태스크가 B2 를 편집했다. **어떤 시험도 이 함수에 닿지 않는다** — 아래는 그 사실과 구조적 반증이다.
| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 424:2 — refresh 실패 | 없음 | 아니오 | 아니오 — **진입 0** |
| B2 | if at 427:2 — dispatch 부재(handoff 거절은 경계 안으로 갔다) | 없음 (단위 수준은 `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff` 가 같은 판단을 고정한다) | 아니오 | 아니오 — **진입 0** |
| B3 | if at 445:3 — 캠페인 CAS 실패 | 없음 | 아니오 | 아니오 — **진입 0** |
| B4 | if at 448:3 — 이미 점유된 캠페인 | 없음 | 아니오 | 아니오 — **진입 0** |
| B5 | if at 452:3 — lease 소모 | 없음 | 아니오 | 아니오 — **진입 0** |

## 반증 실측

| 뮤테이션 | 결과 |
|---|---|
| M3: `dispatch` 를 메서드 값으로 꺼내 둔다(`send := fresh.dispatch.dispatch`) | KILLED — `TestExactlyOneProductionCallSiteTurnsAHandoffIntoADispatch` 실패(철자 census). 5.5 판본은 호출식만 세었으므로 이 형태를 놓쳤다. |
| M4: 경계가 묶어 주지 않은 이름을 dispatch 에 넘긴다 | KILLED — 같은 시험 실패 |
| P1b-B1: `Deliver` 를 `Single()` 로 되돌리고 관문을 빈 `if !handedOff { }` 로 만든다 | KILLED — `TestNoProductionSiteDiscardsTheSeamsAdmissionAnswer` 실패(경계의 답을 묶는 자리가 3 이 아니라 4 가 된다) |
| P1b-B3: 몸통 안에서 `result` 를 `entries` 의 마지막 값으로 덮어쓴다 | KILLED — `TestSeamConsumersCannotReadTheRawEntryListAgain` 실패 |

이 함수의 동작 커버리지는 0 이다. 그 사실을 적지 않고 넘어가면 구조 가드가 동작 증거처럼 읽힌다.
