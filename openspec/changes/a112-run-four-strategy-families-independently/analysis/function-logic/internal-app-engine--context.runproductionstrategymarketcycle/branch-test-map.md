# Branch Test Map: `Context.runProductionStrategyMarketCycle`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (506-565); file SHA-256 `9e24e93028b2728071d71d1d6ccea2c2a83fe768f6efe2dc09a57906c435a373`. AST branch positions are authoritative.
- **어떤 시험도 이 함수를 통째로 돌지 않는다**(진입 0). 이 함수가 하던 전달(옛 B5~B7)은 2026-09-30 리뷰 수리로
  `dispatchStrategyMarketHandoffs` 로 의미 무변경 이동했고, 거기서 행동 시험이 돈다.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 508:2 — refresh 실패 | 없음 | 아니오 | 아니오 — **진입 0** |
| B2 | if at 532:2 — durable latch 를 읽으며 레인을 세우지 못함(5.3.3) | 이 함수로는 없음. 같은 판단을 `TestADurableLatchThatNamesNoLaneInThisBuildStopsTheCycleLoudlyAndCanBeClosed` 가 레인 런타임에서 잰다 | 아니오 | 아니오 — **이 함수 진입 0** |
| B3 | if at 535:2 — 레인 주기가 durable latch 를 남기지 못함(5.3.3) | 이 함수로는 없음. 같은 판단을 `TestALedgerThatCannotTakeTheLatchStopsTheCycle` 이 잰다 | 아니오 | 아니오 — **이 함수 진입 0** |
| B4 | if at 541:2 — dispatch 부재 | 없음 (단위 수준은 `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`) | 아니오 | 아니오 — **진입 0** |

## 반증 실측

| 뮤테이션 | 결과 |
|---|---|
| 5.2.2.1-fix G06: 옮긴 몸통 안 카운터로 둘째 범위부터 버림(보이스 B X09) | `mutation-5.2.2.1-fix.tsv` 참조 — `TestTheMarketDeliveryDispatchesEveryAdmittedScopeInOrder`(생산 몸통 행동) · `TestTheDeliveryBodyHandsEveryHandoffToOneClosureThatWritesNothingOutside`(캡처 쓰기 금지) |
| 5.2.2.1-fix G07: 이름만 같은 다른 타입의 `dispatchHandoffs`(X08) | 같은 원장 — `TestTheProductionCycleEndsByDeliveringEveryOwnerScopeHandoff`(수신자 타입 해소) |
| 5.2.2.1-fix G08: 전달 함수의 지역 섀도(X10) | 같은 원장 — 같은 시험(패키지 수준 func 해소) |
| 5.2.2.1-fix G09: 시장 단위 handoff 를 넘김(옛 F11) | 같은 원장 — 같은 시험(handoff 원천 해소) |
| 5.2.2.1-fix N01: 동작이 같은 리팩터 — handoff 를 지역 변수로 받아 넘김(X11) | 같은 원장 — **초록이어야 함**(거짓 양성 대조) |
| 옛 M3 · M4 · P1b · P1c · P1d · 5.3.3 N11 | 편집 전 번들(`pre-edit-5.2.2.1-fix/`)의 표 그대로 — dispatch census 는 이제 `dispatchStrategyMarketHandoffs` 를 센다 |

이 함수의 동작 커버리지는 0 이다. 그 사실을 적지 않고 넘어가면 구조 가드가 동작 증거처럼 읽힌다.
