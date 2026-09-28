# Branch Test Map: `ExitObserver.ObserveOnce`

AST 기준 분기 7 / 이탈 6(`return` 5 + `continue` 1). 기존 테스트는 `internal/app/engine/exitloop_test.go`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:418` 체결 감지가 밀리면 주기 전체 양보 + outage 시계 유지 | 기존 harness 경로 | no | yes |
| B2 | `:428` `workingSet` 오류가 사이클 오류가 된다 | 기존 harness 경로 | no | yes |
| B3 | `:432` 무보유 계정은 outage로 승격되지 않는다 | `TestAnAccountHoldingNothingIsNotInAnOutage` `:659` | no | yes |
| B4 | `:444` **전 종목** 미응답이 사다리를 탄다 | `TestASustainedOutageBlocksEntriesAndAlertsOnce` `:608` · `TestAQuoteWithNoLastTradeIsNotAnObservation` `:585` | no | yes |
| B5 | `:453` 보유 포지션마다 정확히 1회 방문 | 모든 테스트가 간접 경유 | no | yes |
| B6 | `:455` **일부 종목만 미응답 → 그 종목이 조용히 빠진다** | **없음 — 이 change의 대상** | no | no |
| B7 | `:462` 첫 `judge` 오류만 사이클에 실린다 | 기존 harness 경로 | no | yes |

## B6가 테스트되지 않았는데 테스트된 것처럼 보이는 이유

`TestAQuoteWithNoLastTradeIsNotAnObservation` (`exitloop_test.go:585`)은 이름이 B6를
덮는 것처럼 읽히지만 **보유 포지션을 하나만 세운다**.

```go
p := h.entry("005930", ...)   // 보유 1종목
h.quote("005930", 0)
cycle := h.observe()
if cycle.Err == nil { t.Fatal(...) }   // ← B4를 단언한다
```

`observe`(`exitloop.go:748-762`)는 `Last <= 0`인 시세를 버린 뒤 `len(out) == 0`일 때만
오류를 낸다. 보유가 1종목이면 버린 결과가 빈 map이라 **B4**로 간다. 그래서 이 테스트가
확인하는 것은 "전 종목이 답하지 않았다"이고, **B6는 한 번도 실행되지 않는다.**

## 필요한 RED (a090 후보)

| # | Scenario | 기대 |
|---|---|---|
| R1 | 보유 2종목, 그중 1종목만 `Last = 0` | `cycle.Err == nil`(다른 종목은 정상 판정)이면서 빠진 종목이 **미관측으로 계수된다** |
| R2 | 보유 2종목, 그중 1종목이 price read 응답에 **부재** | R1과 동일 — 원인이 0가격이든 부재든 결과가 같다 |
| R3 | 같은 종목이 연속 N주기 미관측 | 한계 초과 시 critical 1회, 반복 없음 |
| R4 | 미관측 종목이 다음 주기에 관측된다 | 계수 초기화 |
| R5 | B4(전 종목 미응답)의 기존 동작 | **무변화** — 계정 사다리를 그대로 탄다 |
| R6 | B1(양보)의 기존 동작 | **무변화** — `checkOutage`가 계속 돈다 |

R1·R2가 RED로 실패하는 것이 이 결함의 존재 증명이다. R5·R6은 회귀 방지.

## 산출물 근거

- 분기·이탈 열거: `ast.json` (branches 7, returns 5) + `continue` `:459`는 소스 스캔
- 헤더 불변식: `exitloop.go:40-41` — "we chose not to look" and "we could not look"
  leave a position equally unprotected
- 사다리가 계정 단위임: `exitloop.go:760-762` (`len(out) == 0`에서만 오류)
- 침묵의 증거: `reportCycle` `:381-384` (`cycle.Err != nil`일 때만 기록)
