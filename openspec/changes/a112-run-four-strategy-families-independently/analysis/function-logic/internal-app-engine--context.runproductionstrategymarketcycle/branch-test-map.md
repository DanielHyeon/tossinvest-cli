# Branch Test Map: `Context.runProductionStrategyMarketCycle`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (422-445); file SHA-256 `17ad4c0c684b74686dd1e80b256a06971802afa26bcfd300dbeac9bd5f7e0496`. AST branch positions are authoritative.
- 이 태스크가 B2 를 편집했다. **어떤 시험도 이 함수에 닿지 않는다** — 아래는 그 사실과 구조적 반증이다.
| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 424:2 — refresh 실패 | 없음 | 아니오 | 아니오 — **진입 0** |
| B2 | if at 429:2 — **handoff 거절** 또는 dispatch 부재 | 없음 (단위 수준은 `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff` 가 같은 판단을 고정한다) | 아니오 | 아니오 — **진입 0** |
| B3 | if at 434:2 — 캠페인 CAS 실패 | 없음 | 아니오 | 아니오 — **진입 0** |
| B4 | if at 437:2 — 이미 점유된 캠페인 | 없음 | 아니오 | 아니오 — **진입 0** |
| B5 | if at 441:2 — lease 소모 | 없음 | 아니오 | 아니오 — **진입 0** |

## 반증 실측

| 뮤테이션 | 결과 |
|---|---|
| M6: dispatch 호출 자리를 하나 더 만든다 | KILLED — `TestExactlyOneProductionCallSiteTurnsAHandoffIntoADispatch` 실패 |

이 함수의 동작 커버리지는 0 이다. 그 사실을 적지 않고 넘어가면 구조 가드가 동작 증거처럼 읽힌다.
