# Branch Test Map: `Context.runProductionStrategyMarketCycle`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (421-444); file SHA-256 `12586e3cf90b708e66988931ad424d7312593bf518f0987a0893bf4f6f4b6fb9`. AST branch positions are authoritative.
- 이 태스크가 B2 를 편집했다. **어떤 시험도 이 함수에 닿지 않는다** — 아래는 그 사실과 구조적 반증이다.
| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 423:2 — refresh 실패 | 없음 | 아니오 | 아니오 — **진입 0** |
| B2 | if at 428:2 — **handoff 거절** 또는 dispatch 부재 | 없음 (단위 수준은 `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff` 가 같은 판단을 고정한다) | 아니오 | 아니오 — **진입 0** |
| B3 | if at 433:2 — 캠페인 CAS 실패 | 없음 | 아니오 | 아니오 — **진입 0** |
| B4 | if at 436:2 — 이미 점유된 캠페인 | 없음 | 아니오 | 아니오 — **진입 0** |
| B5 | if at 440:2 — lease 소모 | 없음 | 아니오 | 아니오 — **진입 0** |

## 반증 실측

| 뮤테이션 | 결과 |
|---|---|
| M3: `dispatch` 를 메서드 값으로 꺼내 둔다(`send := fresh.dispatch.dispatch`) | KILLED — `TestExactlyOneProductionCallSiteTurnsAHandoffIntoADispatch` 실패(철자 census). 5.5 판본은 호출식만 세었으므로 이 형태를 놓쳤다. |
| M4: 경계가 묶어 주지 않은 이름을 dispatch 에 넘긴다 | KILLED — 같은 시험 실패 |
| M6: B2 의 관문을 지우고 `_ = handedOff` 로 컴파일러를 달랜다 | KILLED — `TestNoProductionSiteDiscardsTheSeamsAdmissionAnswer` 실패 |

이 함수의 동작 커버리지는 0 이다. 그 사실을 적지 않고 넘어가면 구조 가드가 동작 증거처럼 읽힌다.
