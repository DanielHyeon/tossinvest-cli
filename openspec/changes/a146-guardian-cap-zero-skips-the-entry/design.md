# a146 design — Guardian 상한 0 = 이번 진입만 건너뜀

근거 표기: 「AST:」는 `analysis/ast-evidence/<번들>.ast.json`(`tools/logic-map`, HEAD `befea6ee`, 편집 전) — 분기 번호는 그 열거의 `B<n>`(줄).
「특성:」은 `internal/app/engine/a146_cap_zero_characterization_test.go`(편집 전 HEAD 에서 통과 — `analysis/red/characterization-head.txt`).
「raw:」는 `analysis/code-context/raw/`(codegraph 1.6.0 · grep 원문).

## 사용자 결정 (2026-10-15 — Manager 전달 원문)

> USER DECISION (2026-10-15, quoted): Guardian 상한 0 → 레인 잠금 대신 「이번 진입만 건너뜀」. Budget stays 1/3 (a139). Explicitly REJECTED by the
> user-facing analysis: moving latch state out of the ledger / in-memory control / async ledger writes (breaks a112 5.3.3 durable latch and the a136/a138
> reservation-tx serialization). Out of scope: scaling quantity down to the remaining daily loss.
>
> GOAL: cap-0 (even 1 share exceeds the per-trade budget) becomes a typed, non-abnormal per-candidate SKIP outcome that never reaches latchMarket and
> is visible (audit/outcome reason), while genuine abnormal failures latch exactly as today. Decide the scope-refusal path with evidence (it may be a
> separate finding — report, don't silently widen).

## F. 편집 전 사실 (HEAD `befea6ee`)

| # | 사실 | 근거 |
|---|---|---|
| F1 | Guardian 상한은 `AccountBaseStrategyEntryQuantity`(생산 짝 경로) 또는 `StrategyEntryQuantity`(legacy 시험 identity)의 답이다. 그 함수가 오류면 precheck 는 `EXISTING_GUARDIAN_CAP`(field `q_existing_guardian`)로 거절하고 **원인으로 그 오류를 싣는다**(`qFinalRefusal(…, err)` — `Unwrap` 이 돌려줌). 답이 정수가 아니거나 0 이면 같은 코드에 해석 오류(또는 nil)를 싣는다 | AST `execgw--riskguardian.precheckqfinalentry` B11(142) · B13(147) · B14(151); riskguardian_qfinal.go:33 `Unwrap` |
| F2 | `risk.ErrStrategyQuantityZero` 는 두 자리에서만 나온다: (a) 진입가 × 승수 ≤ 0 또는 주문 notional ≤ 0, (b) `min(위험 상한, 주문 수량 상한, notional 상한) ≤ 0`. 첫 다리 경로에서 (a)는 닿지 않는다 — precheck 전에 `validateStrategyFirstLegAuthority` 가 진입 · 손절가를 봉인된 조건과 같게 묶고(B3 201), 조건은 `0 < 손절 < 진입 < 목표` 의 양의 정수 minor 이고(`validExecutionTermsFields` — 분기 없는 단일 식), FX 승수는 rate > 0 · haircut 검사 뒤의 곱이며(AST `risk--accountbasefx.multiplierat`), notional · 수량 상한 · 예산은 `Policy.Validate` 가 양수로 강제한다(AST `risk--policy.validate`). 수량 상한 ≥ 1 이므로 (b)는 **위험 상한 0(예산이 한 단위 위험을 못 덮음)** 또는 **notional 상한 0(한 단위 가격이 주문 notional 상한을 넘음)** 이다 | AST `risk--accountbasestrategyentryquantity` B9(280) · B12(292), `risk--strategyentryquantity` B8(154) · B11(166), `risk--accountbaseriskquantity` B5(232) |
| F3 | `EXISTING_GUARDIAN_CAP` 의 다른 생산자 `riskbucket.CalculateAdmission` B2(71)은 원인 nil 이고, precheck 가 B14 에서 0 을 먼저 거절하므로 이 경로에서 닿지 않는다 | AST `riskbucket--calculateadmission` B2; raw grep `RefusalExistingGuardianCap` 3 자리 |
| F4 | 첫 다리 승인은 precheck 거절을 a137 표지(측정 불가 B6 · 한도 B8) → **버킷 고갈만 범위 거절**(B9 128) 순으로 가르고, 나머지는 원인을 실은 `AUTHORITY_MISMATCH` 다(121-123). dispatch 는 원인을 `%w` 로 감싸 돌려준다 | AST `engine--strategyfirstlegadmissionbridge.admit` B5~B10, `engine--strategydispatchcycle.dispatch` B12 · B13(152 · 155) |
| F5 | 전달 루프는 범위 거절만 건너뛰고(B3 74 — 모아서 끝에 오류로 돌려줌), 일일 손실 표지는 주기를 오류 없이 끝내며(B4 82), **그 밖의 오류에서 멈춘다**(B5 87 · return). 그래서 상한 0 범위 뒤의 범위는 그 주기에 평가되지 않는다 — 특성: precheck 1 회(두 범위) | AST `engine--delivereachstrategyhandoff`; 특성 `…StopsTheCycleAtItsScope` |
| F6 | 감독자 `runMarket`: 오류 없는 주기는 B11(953) `continue`. `refreshOnly`(= `!effective && RefreshesAuthority`, 923)면 B12(956) 삼킴 계수 + `continue`(중앙 무결성은 B13 진입 차단). 아니면 중앙 무결성 B14(980), 그 밖은 `latchMarket`(984) — `latched = true`(1139) · fault 하나(1164). `evaluationState` B1(1059)이 잠긴 시장을 다시 평가하지 않는다. **잠금은 메모리 칸이다**: `Faults()` 를 읽는 생산 코드 0(raw `grep-Faults-readers.txt`), 원장의 durable 잠금은 전략군 레인 `persistMarketLatches` 만 쓴다(raw `grep-durable-lane-latch-writers.txt`) — a139 의 「원장 잠금 · 재시작으로 안 풀림」은 이 경로에서 사실이 아니고, 잠긴 시장은 **프로세스 재시작까지** 닫힌다. 특성: effective → 첫 주기에 잠김 · fault 1 · 재평가 0, refreshOnly → 매 주기 삼킴 | AST `engine--strategyentrysupervisor.runmarket` · `.latchmarket` · `.evaluationstate`; 특성 `…LatchesAnEffectiveMarket` · `…IsASwallowedFaultOnARefreshOnlyMarket` |
| F7 | **오늘 생산은 F6 의 refreshOnly 갈래다.** 실행되는 감독자는 `NewRefreshingPairedStrategyEntrySupervisor` 하나(cmd/tossctl/engine.go:690 · :742)이고 그 worker 는 둘 다 `RefreshesAuthority: true`, `Effective` 없음(AST B3 408 루프). `Effective: true` 서술자를 만드는 `buildProductionStrategyMarketWorker`(:526)의 감독자는 투영(`publishStrategyRuntime`)에만 쓰이고 `Run` 되지 않는다(raw `grep-production-supervisor.txt`). 그래서 생산에서 상한 0 은 잠금이 아니라 (i) 삼킴 계수 증가, (ii) 그 주기의 뒤 범위 굶음(결정적 순서라 매 주기 반복), (iii) 그 시장 shadow 관측 무효화(`strategyCycleWithShadow` B3 79)다 | AST `engine--context.newrefreshingpairedstrategyentrysupervisor`, `engine--strategycyclewithshadow` |
| F8 | 범위 거절(버킷 고갈 · 범위 국소 위험 권한 부재)은 같은 주기의 다른 범위 발급을 막지 않지만, 건너뛴 거절이 주기 오류로 돌아와(B3 → 92 `errors.Join(skipped...)`) effective 시장을 잠근다(refreshOnly 는 삼킴). 이것은 a112 J4(「감독자가 사이클 오류로 센다」)와 a137 의 명시 시나리오(risk-management 「앞서 건너뛴 범위 거절과 섞인 한도 도달」)가 정한 동작이다. 특성: `…ABucketExhaustedScopeStillLatchesAnEffectiveMarket` | AST `engine--delivereachstrategyhandoff` B3; openspec/specs/risk-management/spec.md:618-621 |
| F9 | 편집할 함수는 전부 첫 다리(매수) 진입 경로에만 있다: `admit` 의 생산 호출자는 `dispatch` 하나(raw `grep-bridge-admit-nontest.txt` — CodeGraph 는 이름이 같은 다른 `admit` 들과 섞음, reconciliation), `deliverEachStrategyHandoff` 는 `dispatchStrategyMarketHandoffs` 하나, `runMarket` 은 `Run` 하나(raw callers). 매도 · 손절 · 비상 청산 · 대사 · 체결 감지 루프는 이 셋을 부르지 않는다 — 스펙 「entry worker 장애는 safety lifecycle을 지연하지 않는다」가 그 분리를 정한다 | raw `callers-*.txt` · `impact-*.txt` |

## D1. 분류 — 타입 · 코드 · sentinel 셋의 곱

`admit` 의 precheck 거절 갈래에서, a137 두 갈래(측정 불가 · 한도) **뒤**, `AUTHORITY_MISMATCH` 를 만들기 **앞**에:

```go
if qFinal := (*execgw.QFinalRefusal)(nil); errors.As(err, &qFinal) && qFinal.Code == riskbucket.RefusalExistingGuardianCap &&
	errors.Is(err, risk.ErrStrategyQuantityZero) {
	// → 코드 GUARDIAN_CAP_ZERO, cause = &strategyEntrySkip{market, symbol, cause: err}, 기록자 호출
}
```

- 세 조건 모두 필요하다. 코드만 보면 F1 의 해석 실패 · F3 의 nil 원인 · FX/정책 오류(같은 코드, 다른 원인)가 조용해진다. sentinel 만 보면 다른 경로가
  같은 sentinel 을 실은 경우를 받는다. 문구는 보지 않는다(a137 P2-1 과 같은 규칙).
- 다루는 것(F2): **위험 상한 0** 과 **notional 상한 0** 둘 다 — 사용자 결정의 「Guardian 상한」은 세 상한의 min 이고, 둘을 가르는 타입이 `risk` 에 없다.
  가르려면 `risk` 사이징 함수의 반환을 바꿔야 하므로(High-risk 사이징 편집) 하지 않는다 — **사람 결정 잔여로 보고**(notional 쪽도 같은 「건너뜀」인지).
- `strategyEntrySkip` 은 엔진 패키지 비공개 타입(`Error` · `Unwrap` 만, `As` 없음 — `TestNoEngineErrorTypeImplementsAs`). 만드는 자리는 `admit` 하나이고
  언급 census 시험이 그것을 고정한다(범위 거절 census 와 같은 규칙 — 만드는 자리가 늘면 결함이 조용해진다).
- `isCentralStrategyIntegrity` · `*strategyScopeRefusal` 판정은 이 원인에 대해 거짓이다(표 시험).

## D2. 전달 — 건너뛰되 모으지 않음

`deliverEachStrategyHandoff` 에서 범위 거절 판정 다음에 `errors.As(err, &skip)` 이면 `continue`. `skipped` 에 넣지 **않는다** — 넣으면 주기 오류가 되어 F6 의
잠금 · 삼킴으로 돌아간다(결정 위반). 흔적은 D3 이 남긴다(조용한 건너뛰기 금지 J4 의 취지는 기록이지 오류가 아님 — a137 이 같은 이유로 일일 손실 표지를 오류 밖에 둠).

결과 표(편집 뒤):

| 주기 안의 결과들 | 반환 | effective | refreshOnly |
|---|---|---|---|
| 스킵만 / 스킵 + 발급 | nil | 잠기지 않음 | 삼킴 0 |
| 스킵 + 범위 거절 | `Join(범위 거절들)` (오늘과 같음 — F8) | 잠김(F8, 범위 밖) | 삼킴 |
| 스킵 뒤 멈추는 결함 | 그 결함(앞 범위 거절이 있으면 함께) — 스킵은 합치지 않음 | 잠김(오늘과 같음) | 삼킴 |
| 스킵 뒤 일일 손실 표지 | `Join(범위 거절들)` | 오늘과 같음 | 오늘과 같음 |

## D3. 흔적 — 계수 · 첫 원인 · 하루 한 줄(종목별)

a137 측정 불가의 모양을 그대로 쓴다: `runMarket` 이 주기 ctx 에 기록자를 싣고(`withStrategyEntrySkipRecorder` — a137 기록자 바로 옆 한 줄), `admit` 이 스킵을
만들 때 `recordStrategyEntrySkip(ctx, skip)` 으로 넘긴다(기록자가 없으면 — 시험 · 감독자 밖 — 아무것도 안 함).

- 스냅숏: `GuardianCapZeroSkips`(포화 계수) · `FirstGuardianCapZeroSkip`(첫 원인 — 고정 문구 + 시장 · 종목 · 코드, 계좌 값 · 금액 없음).
- 알림: 옵션 `EntrySkipNotice(market, day, symbol, cause)` — 생산 조립이 경고 줄 `engine.strategy_entry_skipped` 하나로 연결. **(계좌 일 키, 종목)마다
  시장별 한 번**(같은 후보가 5 초마다 다시 와도 하루 한 줄). 일 키가 바뀌면 기억을 비운다(하루 종목 수만큼으로 유계).
- 원장 · 운영 모드 · 진입 게이트는 건드리지 않는다(정책 결과이지 차단이 아님 — 매 주기 다시 판정되므로 예산 · 가격이 바뀌면 다음 주기에 진입할 수 있다).

## D4. 바뀌지 않는 것 (시험으로 고정)

- 그 밖의 모든 precheck 거절 분류(a137 표 시험 + a146 표: 해석 실패 · nil 원인 · FX 오류 · 다른 q_final 코드 + 같은 sentinel 의 다른 코드).
- 범위 거절 보고와 effective 잠금(F8 특성 시험 — 편집 뒤에도 통과).
- 표지 없는 결함은 effective 잠금 · refreshOnly 삼킴(a137 `TestA137AnUnmarkedFaultStillLatchesOrIsSwallowedAsToday` 그대로).
- 상한 0 진입은 브로커에 닿지 않는다(발급 · lease · Gateway 호출 0 — 스파이).

## D5. 거절되던 것 중 이제 통과하는 것 (fail-closed 를 푸는 쪽의 열거)

- 상한 0 범위 **뒤의 범위**가 같은 주기에 평가된다 — 그 범위는 여전히 전체 Guardian precheck · 원자 발급 · Gateway 재확인을 지나야 주문이 된다(넓어지는 것은
  「평가될 기회」뿐이고 판정은 그대로).
- 스킵만 있던 주기의 shadow 관측이 무효화되지 않는다(측정 전용 — 주문 경로 무관).
- effective 시장이 상한 0 하나로 재시작까지 닫히지 않는다.
- 상한 0 이 진짜 결함을 가리는 경우: precheck 는 상한 계산 **앞**의 결함(정책 버전 · 소유자 · 통화 · FX — B2~B10)을 먼저 거절하므로 그것들은 가려지지 않는다.
  상한 계산 **뒤**의 판정(버킷 · 체인 · 노출 — B15~B17)은 상한 0 일 때 오늘도 돌지 않는다(같은 순서 — 새로 가려지는 것 없음).

## D6. 기각한 안

- 잠금 상태를 원장 밖으로 / 메모리 제어 / 비동기 원장 쓰기 — 사용자 기각(위 원문). a146 은 잠금 저장소를 건드리지 않는다.
- 범위 거절로 재분류(`strategyScopeRefusal`) — 범위 거절은 주기 오류로 모여 effective 를 잠근다(F8). 결정의 목표에 못 닿는다.
- 감독자에서 오류 문구로 걸러 잠금만 피하기 — 문구 분류 금지(a137 P2-1), 그리고 뒤 범위 굶음(F7 ii)이 남는다.
- 수량을 남은 일일 손실에 맞춰 줄이기 — 범위 밖(사용자).

## D7. 범위 거절 경로 — 별도 발견 (보고만)

F8 은 명시된 동작이므로 a146 은 바꾸지 않는다. 다만 결과는 「한 family 버킷 고갈 → 같은 주기의 다른 family 는 발급되지만 그 시장 전체가 재시작까지 잠김」(effective)
이고, 스펙 「한 family risk bucket 고갈」의 「continuation 은 계속 평가」와 주기 단위로만 맞는다. 생산(refreshOnly)에서는 잠기지 않고 삼킴 계수로 남는다. 이것을
스킵과 같은 정책 결과로 볼지는 **사람 결정** 후보다(이 change 밖).

## D8. 변이 계획 (원장 `analysis/mutation/`)

분류: 코드 조건 제거 · sentinel 조건 제거 · `errors.Is` 를 다른 sentinel 로 · 갈래를 a137 갈래 앞으로(순서) · cause 를 원래 오류로(타입 잃음) · 코드 상수를
`AUTHORITY_MISMATCH` 로. 전달: 스킵 `continue` 를 `skipped` 에 넣기 · 스킵 판정 제거 · 스킵을 `return nil`(뒤 범위 굶음). 흔적: 기록자 설치 제거 · 기록 호출 제거 ·
계수 증가 제거 · 첫 원인 덮어쓰기 · 하루 dedup 제거 · 일 키 비우기 제거 · 알림 배선 제거. 무변이 대조군은 GREEN 이어야 한다.
