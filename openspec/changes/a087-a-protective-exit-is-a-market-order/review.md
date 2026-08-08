# a087 proposal-freeze 리뷰

- **날짜**: 2026-08-06
- **대상**: `proposal.md` / `design.md` / `specs/order-execution/spec.md` / `tasks.md` (base `ec29dc72`)
- **위험 등급**: High-risk (자동 손절 주문의 가격 산출) → 적대적 Eng 관점 **필수**
- **판정**: **FREEZE 거부.** 문서 수정 후 재리뷰.

## 보이스 구성

| 보이스 | 상태 |
| --- | --- |
| Claude CEO (독립, 사전 리뷰 미열람) | 실행 — 9건 (critical 2 / high 4 / medium 3) |
| Claude Eng (독립, 적대적, 사전 리뷰 미열람) | 실행 — 14건 (critical 3 / high 4 / medium 7) |
| Codex CEO·Eng | **`[codex-unavailable]`** — 사용량 한도 소진(2026-08-08 회복). 축소 태그 `[subagent-only]` |

WORKFLOW의 "리뷰어 주장은 Manager가 코드로 재검증"에 따라 아래 표의 주장을 전부 직접 확인했다.
**재검증에서 살아남지 못한 주장은 없다. 두 건은 리뷰어 추정보다 더 나빴다.**

## 재검증 결과

| # | 주장 | 확인 방법 | 결과 |
| --- | --- | --- | --- |
| C2 | `big.Rat`을 `SetFloat64`로 만들면 온그리드 US 가격이 내려간다 | Go 재현, 센트가 99,901개 | **확인 — 47,952건(48%) 이동.** 리뷰어 추정 "1/20"보다 10배 나쁘다 |
| C3 | flatten과 verifylive의 US 동작이 다르다 | 같은 재현 | **확인 — 4,543건 불일치.** design의 "유일한 실질 차이는 sub-$1"은 거짓 |
| C1 | 합성 손절은 §0.9 근거로 **올림**이 정본 | `internal/exitpolicy/adoption.go:75-82` | **확인.** tasks 6.2(`TickFloor`)는 문서화된 안전 결정의 역전 |
| F1 | drift 가드는 수학적으로 발화 불가 | `floor(p/t)·t` ⇒ `|Δ| < t` 항상 | **확인.** D5와 spec Requirement 3 둘째 절은 만족 불가능 |
| F2 | 245,750은 245,500과 246,000의 **정확한 중간값** | 산술 + 원장 전수 | **확인.** KR intents 14건 중 off-grid 5건 전부 `rem == tick/2` |
| F3 | flatten의 1순위는 테이블이 아니라 **거래소 하한가** | `internal/flatten/liquidate.go:372-378` | **확인.** proposal의 "비상 청산은 그리드를 지킨다(테이블로)"는 오독 |
| F8/C1 | `sellIntent`는 익절도 처리한다 | `exitloop.go:1208` `isFullExit`에 `ActionLadderTakeProfit` 포함 | **확인.** 방향 논거가 익절에 미적용 |
| F4b | `RoundUpToTick`·`OneTickFurther`가 `TickSize`를 호출 | `pricing.go:289,319` | **확인.** tasks 2.4의 "건드리지 않는다"는 거짓 |
| H4 | 시장 문자열이 대소문자로 갈린다 | `verifylive.MarketKR="KR"` vs `clock.MarketKR="kr"` | **확인** |
| M7 | `docs/trading/measurements.md`가 없다 | `find` | **확인.** 실재는 `openspec/changes/verify-execution-capability/measurements.md` |
| F5 | 3번째~4번째 시도 사이 7분 43초 공백 | engine.log | **확인.** 관측 주기 5s·지연 한계 30s인데 **지연 이벤트 0건** |

## 양 보이스 합의 (독립적으로 같은 결론)

| # | 항목 | CEO | Eng |
| --- | --- | --- | --- |
| 1 | `big.Rat`은 순손실 — 모든 호출자가 float64 | F7 | C2 |
| 2 | "동작 무변화" 주장이 거짓 | F4a | C3 |
| 3 | AST 리터럴 가드는 양방향으로 불건전 | F9 | H3 |
| 4 | 브로커가 이미 `tickSize`·`nearestPrices`를 준다 — 안 쓴다 | F3 | M1/M2 |
| 5 | `internal/costs`는 잘못된 집 | F9 | M4 |
| 6 | spec의 SHALL이 구현 범위를 초과 (2c 위반 예약) | F6 | M6 |
| 7 | 방향은 값 종류마다 달라야 하는데 미적용 | F8 | C1 |
| 8 | 전제가 n=1 — 측정이 먼저 | F2 | M7 |

**합의 8/8.** 두 보이스가 서로를 못 본 상태에서 같은 8개 축에 도달했다. DISAGREE 없음.

## 차단 사유 (freeze 거부의 근거)

1. **C1 — 문서 내부 모순.** `proposal.md`의 Non-goal("baseline 정렬 안 함")과 `design.md` D2 + `tasks.md` §6(신규 스톱 정렬)이 서로를 부정한다. 게다가 그 정렬 방향이 `adoption.go`가 §0.9 근거로 고정한 방향의 **역전**이다. High-risk change가 "어떤 값을 어느 방향으로 정렬하는가"에 대해 자기모순인 채로 얼면 안 된다.
2. **C2 — 대표 설계 선택이 틀렸다.** `big.Rat`+`SetFloat64`는 이 change의 spec Scenario("이미 그리드 위면 변경되지 않는다")를 US 가격의 48%에서 위반한다.
3. **C3 — 대표 안전 주장이 거짓.** 두 사본 위임이 동작을 보존하지 않는다.
4. **H1 — 이 change가 없애려는 상태를 새로 만든다.** `sellIntent`에 거부 경로를 추가하면 미지 시장·0 이하에서 포지션이 손절 없이 남는다. D5가 200줄에 걸쳐 금지한 바로 그 결과다.

## 수용한 수정 (구현 전 문서 반영)

| ID | 수정 | 출처 | 원칙 |
| --- | --- | --- | --- |
| A1 | `big.Rat`+`SetFloat64` 폐기. 정본 API는 **십진 문자열**(또는 `SetString`)을 받는다. 엔진은 이미 `decimalOf`로 정확 십진을 들고 있다 | C2/F7 | P5 explicit |
| A2 | **tasks §6 삭제.** 신규 스톱 정렬은 이 change에서 하지 않는다. `proposal.md`의 Non-goal이 정본 | C1 | §0.9 |
| A3 | **tasks §3(AST 가드) 삭제.** 대신 `execgw.checkOrderShape`에 그리드 불변식을 **단언**한다 — 모든 자동 주문이 지나는 유일한 길목이고 이미 "LIMIT 전용"·"양수 가격"을 강제한다 | F9/H3/H2 | P5 |
| A4 | `verifylive` 위임 **철회**. `flatten`만 위임하고, US sub-$1 행은 명시적 **동작 변경**으로 별도 §0.3 논거와 테스트를 갖는다 | C3/F4 | P1 completeness |
| A5 | `sellIntent`는 그리드 때문에 **거부하지 않는다.** 미지 시장·0 이하 → 정렬 없이 원값 제출 + critical 관측 | H1 | §0.3 |
| A6 | D5 재작성 — "한 틱 초과 drift"는 발화 불가. 실제 탐지기는 **브로커 400의 `tickSize`·`nearestPrices`를 기록하는 것** | F1/M1 | 측정 우선 |
| A7 | spec Requirement 1의 SHALL을 **축소 주문의 지정가**로 한정. 트리거 가격은 범위 밖이며 방향이 반대(ceil)임을 명시 | F6/M6 | — |
| A8 | 정본 위치를 `internal/costs`에서 재검토. `costs`는 override 가능한 비용 모델이고 fingerprint 감사 대상이라 고정 거래소 사실이 섞이면 안 된다 | F9/M4 | P5 |
| A9 | tasks 8.1 대상 파일 교정 — `docs/trading/measurements.md`는 없다 | M7 | — |
| A10 | 익절 방향을 **명시적으로 결정**하고 bps 비용을 기록. 보호와 같은 논거를 상속하지 않는다 | F8/C1 | §0.9 |
| A11 | `costs.ParseMarket` 도입, 세 호출부 모두 사용. 변환(`costs.Market(s)`)으로 만들지 않는다 | H4 | — |
| A12 | KOSPI 보통주 표임을 코드·spec에 명시. KOSDAQ을 ETF·우선주와 함께 갭으로 기록 | M1 | 측정 우선 |
| A13 | 정렬된 제출가를 원장에 남긴다(additive-nullable). 이 결함의 진단 자체가 `intents`↔`exit_states` 대조였다 | M3 | §0.6 |
| A14 | 테스트 추가: 밴드 교차 연속 판정, `applyFloor`×가격, 부분체결 잔량 재가격, `record` 크래시 순서에 새 실패 지점, 재판정 시 관측 중복 | M5 | P1 |

## 이연 (issues.md)

- **I1 — 7분 43초 재제안 공백.** 관측 주기 5초·지연 한계 30초인데 거부된 보호 제안이 7분 43초 동안 재시도되지 않았고 지연 이벤트가 로그 전체에 0건이다. 호가 정렬과 **독립적인 가용성 결함**이며, 그쪽이 §0.3 노출에 더 크게 기여할 수 있다. 별도 change.
- **I2 — 400 분류.** 호가 이탈은 보호 주문이 400을 받는 여러 이유 중 하나다(가격 밴드, 호가 단위, 수량 단위, 장 시간). 나머지 분류는 미처리.

## 사용자 판단 대기

아래 두 건은 원칙으로 자동 결정하지 않는다 — 안전 수정의 **순서**를 바꾸고, 재시도 금지 규칙의 해석에 걸린다. 본문 대화에서 제시했고, 사용자 결정으로 change가 교체됐다(아래 2차).

---

# 2차 리뷰 — 교체본 `a087-a-protective-exit-is-a-market-order`

- **날짜**: 2026-08-06
- **대상**: 교체된 proposal/design/spec/tasks (보호 청산 = 시장가)
- **보이스**: Claude CEO (독립) 실행 · Codex `[codex-unavailable]` 한도 소진 · **Eng 보이스 미실행**
  (CEO가 범위 산정 자체의 blocking 오류를 찾아 거기서 종료. 교정본이 서면 Eng를 돌린다)
- **판정**: **FREEZE 거부 (2차)**

## 재검증 — 내 proposal의 사실 주장 3건이 거짓이었다

| 주장 (proposal) | 확인 | 결과 |
| --- | --- | --- |
| "**차단은 두 줄이다**" (`failclosed.go:84`, `exitloop.go:1486`) | `internal/trading/service.go:281,188` | **거짓 — 세 번째 관문이 있다** |
| "**StockOS가 이미 검증했다**" (KRX 청산 = MARKET) | `auto_exit_execution.py:2897-2924` | **거짓 — 생략에 의한 오인용** |
| "엔진 청산 루프는 **정규장 기준**이라 차단 요소가 아니다" | `InRegularSession` 호출자 전수 | **거짓 — 엔진에 세션 게이트가 없다** |

### C1 (critical) — 세 번째 관문

```go
// internal/trading/service.go:281 — non-fractional
if intent.OrderType != "limit" { return false }
// service.go:188  → return MutationResult{}, ErrPlaceUnsupported
// execgw/gateway.go:120 → trading *trading.Service   (인터페이스 아님, 구체 타입)
```

`checkOrderShape`와 `sellIntent`만 고치면 **보호 청산이 100% 거부된다. 영구히.** 오늘은
지정가가 가끔 통과한다(실측 6번째가 통과했다). 이 change 이후에는 한 건도 통과하지 않는다.
`alertProposalRefused`의 dedup 키가 `position|action|level`이라 운영자는 **알림 1건 뒤 침묵**이다.

그리고 그 사실이 내가 인용한 줄 **여섯 줄 위 주석에 적혀 있었다**:

> This mirrors internal/trading's capability check rather than calling it (that
> predicate is unexported, and **internal/trading is not ours to change — design D1**).
> … anything it accepts **still faces the real check at the service**.

High-risk change의 범위 산정이 주석 여섯 줄 거리에서 틀렸다.

### C2 (critical) — 인용한 선례가 반대를 말한다

```python
# auto_exit_execution.py:2900 — _apply_exit_order_style
if _is_emergency_breach(plan) or EOD_FLATTEN or early_surge_fast_partial_profit:
    order_type = _aggressive_exit_order_type(...)      # KRX → MARKET
return replace(plan, order_type=BrokerOrderType.LIMIT, ...)   # ← 그 외 전부

# _is_emergency_breach: quote ≤ trigger × (1 − 0.5%),  DEFAULT_..._PCT = 0.5
```

StockOS의 평범한 `STOP_LOSS`·`STOP_LOSS_LADDER` **1차 제출은 LIMIT**이다. MARKET은
**0.5% 초과 이탈에 걸린 에스컬레이션**이다. a087은 임계 없이 1차부터 MARKET이므로
**인용한 실운영 시스템보다 엄격히 더 공격적**이고, 그것을 정당화하는 사다리(a089)는
뒤로 미뤘다. 종착점을 가져와 출발점으로 삼았다.

### C3 (critical) — "체결이 보장된다"는 거짓이고, 최강 대안을 비교하지 않았다

시장가는 **접수**를 보장하지 갭다운·하한가·거래정지·얇은 호가에서 **체결**을 보장하지
않는다. proposal·design 어디에도 슬리피지 bps가 없다. 초안 리뷰 A10은 **한 틱 방향**에도
bps 기록을 요구했는데, 교체본은 가격 통제를 통째로 없애면서 bps를 하나도 적지 않았다.

그리고 대안이 **이미 이 저장소에 구현·출시되어 있다**:

```go
// internal/flatten/liquidate.go:374-378
// The exchange's own floor: maximally aggressive, always a valid tick,
// and never rejected for being outside the band.
return limits.LowerLimit, "exchange lower limit", nil
```

거래소 하한가 지정가는 — 시장가만큼 공격적이고 · 구조적으로 온그리드이며 · 밴드 이탈
거부가 불가능하고 · **원장에 가격이 남고** · `internal/trading`을 건드리지 않는다(여전히 LIMIT).
proposal은 이것을 "별도 논거를 갖는다" 한 줄로 치우고 그 논거를 끝내 말하지 않았다.

### H1 (high) — 세션 무마 근거가 거짓

`InRegularSession`의 비테스트 호출자는 저장소 전체에서 `internal/verifylive/hours.go:130`
**하나뿐**이고 `internal/app/engine`에는 없다. 청산 루프는 정규장 밖에서도 판정한다.
KRX 시간외단일가에는 시장가가 없으므로, 이 change는 없애려는 거부 클래스를 다른 시계
구간으로 옮겨 새로 만든다.

### H3 (high) — 가장 무거운 이연이 더 멀어졌다

실측 9분 중 **7분 43초는 재제안 공백**이지 호가 그리드가 아니다. a087이 완벽히 동작해도
거부 *원인* 하나만 없애고 가용성 결함은 남는다. 초안 리뷰가 I1을 "§0.3 노출에 더 크게
기여할 수 있다"고 했는데, 교체본은 그것을 a089로 **더 밀었다**.

## 초안 리뷰 발견별 해소/회피

| 초안 발견 | 2차 판정 |
| --- | --- |
| C2 `big.Rat` US 48% 이동 · C3 flatten/verifylive 불일치 · C1 문서 모순 · F1 drift 가드 | **진짜 해소** (a088 이연) |
| **H1/A5 `sellIntent`에 거부 경로 만들지 말 것** | **해소를 넘어 개선** — D2가 기존 "가격 없음" 거부를 없앤다. **이 문서에서 유일하게 확실히 옳은 부분이고, 주문 유형과 분리 가능하다** |
| F8 익절 분리 | 절반 — `isProtective` 사용은 옳으나 A10의 bps 기록이 삭제됐다 |
| **F3/A8 flatten 1순위 = 거래소 하한가** | **회피** — 사실은 교정했으나 그것이 직접 대안이라는 함의를 버렸다 |
| **F2 "측정이 먼저"(합의 8)** | **회피** — 기전만 바꾸고 측정은 다시 구현 뒤로 |
| I1 7분 43초 · I2 400 분류 | **더 멀어짐** |

**산술 4건 해소 · 1건 개선 · 2건 회피 · 1건 절반 · 최중요 이연 후퇴 · 그리고 초안에 없던
critical 1건 신설.** 초안은 LIMIT 경로를 벗어나지 않아 C1을 만들 수 없었다.

## 옳았던 것 (기록)

LIMIT 전용 오적용 주장은 **독립 확인에서 옳았다** — `CheckAutomatedEntry(intent EntryIntent)`,
`ErrMarketEntry` 주석의 "an automated **entry** … exposure valuation is undefined",
그리고 **`ErrMarketEntry`의 청산측 호출자는 저장소에 0건**이다. 원장·Guardian이 가격을
요구하지 않는다는 주장도 코드로 확인됐다.

**그러나 진짜 이유는 못 찾았다** — 청산이 지정가인 세 번째 이유는 `placeIntentSupported`,
즉 이 fork의 전송 계층이 온주 주문에 대해 실제로 지원하는 유일한 형태가 LIMIT이라는
사실이다. 잘못된 근거를 반박하는 데는 성공했고 옳은 근거는 놓쳤다.

## 3차 교정 방향 (수용)

1. **a089를 먼저** — 재가격 + 긴급 게이트 우회 + 400 분류. 주문 유형과 무관하게 실측
   노출의 다수를 없애고, 위험도가 낮으며, MARKET을 정당화하는 전제다.
2. **D2만 떼어 지금** — "보호 청산은 가격을 못 읽었다고 거부하지 않는다".
   가격은 **거래소 하한가**를 쓴다(`flatten`이 이미 하는 것). MARKET 없이 성립하고,
   `internal/trading`을 안 건드리고, 원래 사고(245,750 호가 거부)를 **완전히** 없애며,
   원장에 가격이 남는다. §0.3 순개선이고 초안 리뷰 A5와 정확히 일치한다.
3. **실측** — KR/US MARKET 접수 가능 여부, 세션 경계 응답, **청산측 슬리피지 계기**
   (현재 `slippagePct`는 진입 전용이고 청산 슬리피지를 재는 코드가 없다).
4. **그 다음에** MARKET vs 하한가 지정가를 bps 데이터로 결정. a088은 익절 지정가에만 필요.
