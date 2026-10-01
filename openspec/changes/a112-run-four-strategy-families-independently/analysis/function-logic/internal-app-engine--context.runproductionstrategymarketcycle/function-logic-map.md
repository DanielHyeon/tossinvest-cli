# Function Logic Map: `Context.runProductionStrategyMarketCycle`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (520-579)
- Function: `Context.runProductionStrategyMarketCycle` in package `engine`
- Signature: `Context.runProductionStrategyMarketCycle(params=3, results=1)`
- File SHA-256: `64f1cc0b85ecf5693dc5df0622b0f19f665697ea1b6f7616eb6755598f546e97`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file (2026-09-30 리뷰 수리 로트).
  앞 판본 머리의 `22855de0…` 는 어느 커밋의 파일과도 맞지 않는 오염 값이었다(보이스 B #7 — 4f49a8eb/44d0fb58 에서 유입) — 정정.
- AST evidence: `ast.json` — AST branches 4.
- Risk scan: `risk-pattern-report.md`.
- 편집 전 번들: `analysis/measurements/lot-5.6.2-5.2.2/pre-edit-5.2.2.1-fix/`(분기 7 — 몸통 closure 의 셋 포함).

## Inputs and invariants

한 시장의 권한을 새로 고치고(`521:16`), 그 시장의 네 전략군 레인을 한 번씩 돌린 뒤
(`545:16` 레인 런타임 · `549:12` `lanes.evaluate`), 주문 경로에 이 시장의 handoff 들을 넘긴다(`578:9`).

**2026-09-30 리뷰 수리 — 전달 몸통의 의미 무변경 이동.** 5.2.2.1 판본까지 이 함수의 마지막 문장은
`deliverEachStrategyHandoff(…dispatchHandoffs(), func(delivered) error { CAS 읽기 · dispatch })` 였고 몸통 closure 에
분기 셋(캠페인 CAS 실패 · 이미 점유 · lease 소모)이 있었다. 이 함수는 권한 새로 고침 전체가 필요해 **어떤 시험도 돌지 못했고**
(진입 0 — 아래), 그 자리를 지키던 구조 못은 이름 모양만 봐서 셋에 뚫렸다(codex P1-1 · 보이스 B #2: 몸통 안 카운터로 둘째 범위부터
버림 · 이름만 같은 다른 타입 메서드 · 지역 섀도). 그래서 몸통을 `dispatchStrategyMarketHandoffs`
(`strategy_market_handoff_delivery.go`, 새 함수 — 편집 전 번들 없음, FLM not-applicable)로 옮기고 원장 · dispatch 를 좁은
인터페이스로 받게 했다. 옮기기 전후 몸통이 같은 코드라는 영수증은
`analysis/measurements/lot-5.6.2-5.2.2/extract-receipt-5.2.2.1-fix.txt`(gofmt 정규형 동일 · 비교기 양성 대조 포함). 이제 이 함수의
마지막 문장은 `return dispatchStrategyMarketHandoffs(ctx, c.Journal, fresh.dispatch, fresh.proposals.forMarket(market).dispatchHandoffs())`
하나이고, 그 모양(부르는 함수 · 넘기는 원장 · dispatch 주기 · handoff 원천)은 `TestTheProductionCycleEndsByDeliveringEveryOwnerScopeHandoff`
가 **식별자 해소(go/types)** 로 못 박는다. 옮겨 간 몸통은 스파이로 직접 돈다(`a112_owner_scope_delivery_test.go`).

**활성화 세대의 출처(`550:3`).** `fresh.schedule.forMarket(market).restore.Activation.Generation()` — durable latch 의 복구 조건이
서명과 묶인다(`TestTheRecoveryGenerationComesFromTheVerifiedActivationAndNothingElse`).

**레인 사이클이 새로 고침 잠금 밖인 이유**와 **관측은 돌려주지 않고 오류만 돌려주는 이유(5.3.3)** 는 이전 판본 서술 그대로 —
`evaluate` 오류(B3 `549:2`)는 durable latch 를 남기지 못했다는 뜻이다.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- **측정: 어느 스위트도 이 함수에 들어오지 않는다(진입 0).** 이 수리 로트가 그 사실을 바꾸지 않았다 — 바꾼 것은 이 함수가 하던
  전달 일을 시험이 돌 수 있는 함수로 옮긴 것이다. 보이스 B 가 00e1b9bd 에서 태그 스위트 커버리지를 재측정해 이 함수 본문 블록
  전부 count 0 을 확인했다.

Exact AST return positions: 523:3, 547:3, 553:3, 556:3, 578:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 522:2 | 진입 0 — refresh 실패 |
| B2 | if | 546:2 | 진입 0 — durable latch 를 읽으며 레인을 세우지 못함(5.3.3) |
| B3 | if | 549:2 | 진입 0 — 레인 주기가 durable latch 를 남기지 못함(5.3.3) |
| B4 | if | 555:2 | 진입 0 — dispatch 부재 |

옛 B5~B7(캠페인 CAS 실패 · 이미 점유 · lease 소모)은 몸통과 함께 `dispatchStrategyMarketHandoffs` 로 옮겨 갔고, 거기서
`TestAFaultInOneScopeStopsTheCycleBeforeTheNext`(CAS 실패) · `TestASkippedScopeDoesNotStopTheNextOne`(점유 · 비-FLAT · lease 소모)이
**행동으로** 돈다(편집 전에는 셋 다 진입 0).

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `c.refreshPairedStrategyEntryProductionAssembly` | 521:16 |
| `c.productionStrategyLanes` | 545:16 |
| `lanes.evaluate` | 549:12 |
| `restore.Activation.Generation` | 550:3 |
| `fresh.schedule.forMarket` | 550:3 |
| `familyActivation` | 551:3 |
| `fresh.proposals.forMarket` | 551:3 |
| `strategyLaneInputs` | 552:3 |
| `fresh.proposals.forMarket` | 552:36 |
| `dispatchStrategyMarketHandoffs` | 578:9 |
| `dispatchHandoffs` | 578:72 |
| `fresh.proposals.forMarket` | 578:72 |

## State mutations and fallbacks

- AST assignments: 3. Defers: 0. Goroutine statements: 0.
- 이 함수 자신은 상태를 쓰지 않는다. 쓰기는 `dispatchStrategyMarketHandoffs` 가 부르는 공유 dispatch 안, 그리고 레인 런타임의
  durable latch 기록(`evaluate` 안)에서 일어난다.

## Safety conclusion

- 공유 dispatch 를 부르는 생산 자리는 여전히 하나다 — 이제 `dispatchStrategyMarketHandoffs`(census 상수 `dispatchCallSiteFunc`,
  `dispatchMentionCensus` 의 `dispatcher.dispatch` 1).
- 활성화 없는 시장(오늘 생산 전부): `dispatchHandoffs` 가 `dispatchHandoff()` 하나를 돌려주고 옮긴 몸통이 편집 전과 같은 코드이므로
  토글 OFF = upstream 동작 불변(영수증 · 행동 시험).
- High-risk impact: yes(주문 경로) — 편집은 의미 무변경 이동과 인자 둘의 인터페이스화뿐.

a112 5.2.2.2: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)

a112 5.2.2.2 리뷰 수리: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)
