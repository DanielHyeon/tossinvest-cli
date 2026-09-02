# Branch Test Map: `StrategyEntrySupervisor.waitMarketRestart`

- Source SHA-256: `66150078e25dfad6d1fec322b955e5f23e3aad77f0525321867a500e0960f58f`; AST branch locations are authoritative.
- Revision: base — 편집하지 않는다. 태스크 5.6 이 인용한다.

커버리지는 측정이다(측정 방법은 같은 change 의 다른 branch-test-map 과 동일하다).

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 853:2 — 기한이 0 | **없음** | no (base) | **no — block 837-839 count=0, 도달 불가(아래)** |
| B2 | if at 857:2 — 현재 시각이 0 | **없음** | no (base) | **no — block 841-843 count=0, 도달 불가(아래)** |
| B3 | if at 861:2 — 남은 시간이 30s 상한을 넘는다 | `TestTheFourEscalationsThatStopTheEngineAreExactlyTheSupervisorsOwnBrokenBookkeeping` 의 두 칸 | no (base) | yes (block 845-847 count=1, **5.6 이 처음 실행**) |
| B4 | if at 864:2 — 기한이 이미 지났다 | `TestPairedMarketRestartHonorsPublishedAbsoluteDeadlineAfterHandoffRace` | no (base) | yes (block 848-850) |
| 본문 | 867:2 — 남은 시간만큼 취소 가능한 대기 | `TestALatchedMarketSkipsTheTriggersAlreadySittingInItsQueue` 외 4 | no (base) | yes |

## 측정으로 확인한 빈칸

- `837-839` (B1) — 이 함수의 유일한 호출자 `runMarket` 은 `latchMarket` 이
  **성공했을 때만** 여기 온다. 성공 반환은 `fault.RestartNotBefore` 이고, 그것은
  `strategyRestartNotBefore(observedAt, …)` 이라 0 이 될 수 없다(observedAt 은
  이미 `944:2` 가 0 이 아님을 확인했다). 즉 방어이지 도달 가능한 갈래가 아니다.
- `841-843` (B2) — 같은 이유의 이중 방어다. 시계가 0 을 돌려주는 상황은
  `latchMarket` 의 B4 가 **먼저** 잡아 확대시키므로 여기까지 오지 않는다.
  5.6 의 첫 칸("관측 시각이 없으면…")이 그 순서를 값으로 확인한다.
