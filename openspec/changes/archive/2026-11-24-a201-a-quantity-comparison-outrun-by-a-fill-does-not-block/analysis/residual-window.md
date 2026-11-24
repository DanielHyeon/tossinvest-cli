# a201 잔여 창 — 재판정과 차단 기록 사이(tasks 1.8)

AST: `cited-ast/RunOnce.ast.json` · `Tracker.Observe.ast.json` · `Tracker.persist.ast.json` · `Tracker.{beginOrderTerminalObservation, updateOrderTerminalBindings, orderTerminalCandidate, advanceStreaks, clearStreaks, refuteUsableAdjustmentCredits}.ast.json`. 재판정 자리 = design D1(RunOnce call@458 Refresh 성공 뒤 · call@464 앞).

## 1. 재판정의 마지막 읽기 → QM 행 삽입까지 끼는 호출(차단을 세우는 갈래 = Observe B3 else@703)

| 순서 | 호출(AST 좌표) | 무엇 | 장부 I/O | 브로커 · 네트워크 | 잠금 |
|---|---|---|---|---|---|
| 1 | RunOnce → `d.opts.Tracker.Observe` call@464 | — | — | — | — |
| 2 | Observe `Now`@630 · `t.interval`@631 · `t.maxFailures`@632 | 시계 · 설정 | 없음 | 없음 | — |
| 3 | `t.mu.Lock`@634 | 추적기 뮤텍스 | — | — | **대기 가능**(`EntryAllowed` · `Blocks` 등과 경합) |
| 4 | `symbolsInDispute`@644 · `promotionKeys`@645 · `beginOrderTerminalObservation`@647(→ `updateOrderTerminalBindings`) · `diff.BlocksEntry`@650 | 메모리 | 없음 | 없음 | 쥔 채 |
| 5 | `blocksFor`@704 · `advanceStreaks`@710/@735 · `refuteUsableAdjustmentCredits`@716(조건부) · `syncGate`@729(조건부) | 메모리 | 없음 | 없음 | 쥔 채 |
| 6 | `resolvePendingPermanentAuthority`@730(조건 `pendingPermanent` — B11 if@708) | 계좌 영구 승격 재시도의 권위 확인 | **있음**: `ActiveReconcileStates` SELECT(mismatch.go:1189 [손]) — 조건부 | 없음 | 쥔 채 |
| 7 | `thresholdKey`@746 · `syncGate`@778 · `pendingBlocksForPersistence`@785 | 메모리 | 없음 | 없음 | 쥔 채 |
| 8 | `persist`@786 → `enterReconcileWithBinding`(persist call@902) | 행 삽입 | **있음**: (missing-order 블록이면) `MaxFillEventID`(order_terminal_release.go:351), 그리고 `EnterReconcile` tx(`BEGIN IMMEDIATE`, 단일 연결) | 없음 | 쥔 채 |

- **브로커 호출 · 네트워크 0.** 장부 I/O 는 8(항상) · 6(조건부). 단일 연결(`SetMaxOpenConns(1)`, journal.go:174 [손])이라 그 사이 체결 tx 가 있으면 `EnterReconcile` 은 그 tx 뒤에 직렬화됨.
- design §4 「브로커 호출 · I/O 없는 메모리 구간」 — **부분 거짓**: 브로커 · 네트워크는 0 이 맞지만 장부 I/O 가 1~3 회(EnterReconcile tx, 조건부 SELECT · MaxFillEventID) 있음. 폭은 ms 단위 추정(미측정).

## 2. 이 구간이 실제로 고착을 만드는 입력 — W2 · W3 에는 없음

- W2 · W3 의 체결은 정의상 **그 종목의 조정 tx 전**에 커밋됨(그래서 조정이 못 쓰임). 조정 tx 는 수렴(call@451) 안이고 재판정은 그 뒤(call@458 이후) → **W2 · W3 의 체결은 항상 재판정 전**. 위 §1 구간에 떨어지는 W2 · W3 체결은 없음.
- 위 구간(그리고 그 뒤)에 떨어져 고착을 만드는 것은 **그 주기에 조정이 시도되지 않은 종목**의 체결뿐 — 수렴이 그 종목에 닿기 전에 멈췄거나(판 중단 G1, converge.go:228-229), 다른 사유로 거절됐거나(복수 시장 등 `soleLiveMarket`), (b) 단독 stale(수량 불변 fill_events 행)로 버려진 종목. 이 종목들은 계좌가 감지기보다 앞선 상태(계좌 매도 반영 · 체결 미커밋)에서 QM 이 서고, 체결이 **재판정 뒤 · 다음 주기 로컬 읽기 전 어디든** 오면 조정 없이 일치 → 고착.
- 그 창의 폭은 §1 의 ms 가 아니라 **다음 주기 로컬 읽기까지(운영 60초 + 안정화 2초 + 수집 2회)**. 재현: `TestA201CurrentBehaviourAStaleStopsTheTrailingSymbolWhoseFillLandsAfterObserve`(B 의 체결을 주기 뒤에 커밋 → B 고착), 탐침에서도 D5 켜고 끄고 모두 B 고착(evidence-phase0.md §4).
- 그래서 design §4 의 잔여 서술은 **창을 잘못 짚음**: 남는 것은 「재판정~Observe 의 메모리 구간」(W2 · W3 에는 0)이 아니라 「조정이 시도되지 않은 종목의 체결이 재판정 뒤에 오는 경우」(≈ 한 주기). 여러 종목의 손절이 한꺼번에 체결되는 날 — 이 change 가 겨냥한 그날 — 앞 종목의 W3 이 판을 멈추면 뒤 종목들이 정확히 이 모양이 됨(claim-check C-R1, Manager 질문).
