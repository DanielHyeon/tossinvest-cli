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

---

# Phase 1 착수 기록 (2026-09-30, Teammate)

## base 재고정 (tasks 0.1) — 승인 기록

- 커밋 `5491451b`: `base-commit.txt` ec29dc72 → `102d4e99` 단독 커밋.
- 승인: 사용자 2026-09-28 결정 ①(D2 분리 선행) + 2026-09-30 재개 지시, Manager 위임(WORKFLOW 「사람 승인 base 재고정」 2).
- 귀속 실측(조건 1): 옛 base 이후 이 change 디렉터리를 만진 비병합 `.go` 커밋 = `a30eb35a` 1건. 그 커밋의 a087 몫은 문서
  6개뿐이고 `.go`(journal/outbox.go · obs/notifier.go + 시험 4)는 a096 몫 — a087 자기 Go 작업 0.
- 옛 base 에서 `check_analysis` required = 295 함수(형제 착지 603 커밋 몫, a087 production 편집 0).

## P1.1 FLM 과 모집단 측정 — Phase 1 편집 중단

- FLM/BTM: `analysis/function-logic/internal-app-engine--exitobserver.sellintent/` (ast.json · FLM · BTM · risk report).
- **발견(blocking)**: D2a 가 앞에 서려는 거부(`sellIntent` B2)는 생산 경로에서 도달 불가다. 호출 사슬 record→submit→sellIntent,
  관측가는 평가기의 `positive` 검사를 지난 값뿐(AST), 엔진 스위트 커버리지 B1·B2 = 0. 상세·선택지는 `issues.md` I-P1.
- Pre-Edit 선언·RED·GREEN·변이·gstack 코드 리뷰: **미착수** — production 편집 0, Manager 결정 대기.

## Phase 1 종결 (2026-09-30, Manager 판정)

- 판정: `issues.md` I-P1 선택지 1 — D2a 불구현, 전제 반증. Manager 독립 대조로 모집단 0 확인.
- 대체 산출물: `7bd8f197` 평가기 관측가 가드 핀 시험(8 부분시험). 변이 `go test -overlay`: 무변이 대조군 GREEN →
  ratchet `positive→nonNegative` CAUGHT · ratchet 가드 제거 CAUGHT · ladder `positive→nonNegative` CAUGHT · ladder 가드 제거 CAUGHT.
- 문서: design D2a 「반증」 절 · D2b 제거 예약 취소 · proposal 3판 처분 · tasks P1.x 종결 + P1.7.
- 남은 범위: Phase 2(§1~§3) — §5 실측(§0.7 사람 승인) 대기.

---

## 2차 proposal-freeze 재리뷰 (2026-10-10, 적대적 Eng)

- **대상**: proposal(3판)·design·tasks·`specs/order-execution/spec.md`·issues — HEAD `b3dadd23`(base `102d4e99`, 그 뒤 237 커밋)
- **보이스**: Claude Eng(독립, 적대적) 1. 교체본에 Eng 보이스가 돈 것은 이번이 처음(2026-08-06 「2차 리뷰」는 CEO 단독, Eng 미실행). Codex·CEO 미실행 — `[subagent-only]`
- **판정**: **수정 필요 — freeze 불가.** P0 1 · P1 4 · P2 5 · P3 3 (계 13)
- **본 범위**: 위 다섯 문서 전문; 현재 HEAD 의 `sellIntent`·`isProtective`·`submit`·`record`(exitloop.go), `checkOrderShape`(failclosed.go), `placeIntentSupported`·`Place`(trading/service.go), `Gateway.place`·저널 intent 조립(execgw/gateway.go), `buildOrderCreate`(official/orders_write.go), `PlaceWireBody`(execgw/wirebody.go), `NormalizePlace`(orderintent/intent.go), refusal_code.go·classify.go·indoubt.go 의 가격 소비부, openapi `POST /api/v1/orders` 계약(`docs/migration/openapi.latest.json`), a089·a100 문서의 a087 인용.
- **방법과 한계**: `rg`/`sed` 로 HEAD 읽기, `git diff` 로 base↔HEAD 함수 본문 대조, CodeGraph **1.6.0** `node`(placeIntentSupported·checkOrderShape 호출자), openapi JSON 추출. **AST 산출물은 만들지 않았다**(task 2.1 소관) — 아래에서 함수 내부 분기·순서를 근거로 쓴 줄은 전부 **손으로 읽은 것이고 「검증 필요(AST 미작성)」**로 표기한다. 시험·실주문은 돌리지 않았다.

### P0-1 — 셋째 관문이 그대로다. tasks 대로 착지하면 보호 청산이 **100% 로컬 거부**된다

2026-08-06 「2차 리뷰」 C1 이 critical 로 적은 결함이 **문서에서 한 줄도 해소되지 않았다.**

- `internal/trading/service.go:281-283` — 비분수 주문은 `OrderType != "limit"` 이면 `false`; `:188-189` `Place` 가 `ErrPlaceUnsupported` 반환. 엔진의 모든 주문은 `execgw/gateway.go:396` `g.trading.Place(...)` 를 지난다(구체 타입 `*trading.Service`, `gateway.go:120`).
- CodeGraph 1.6.0: `placeIntentSupported` 의 비시험 호출자 = `PreviewPlace`(`service.go:109`)·`Place`(`:186`) 둘. `checkOrderShape` 의 호출자 = `CheckPlace`(`failclosed.go:40`) 하나 — **둘은 독립된 관문이다.**
- 거부 경로(검증 필요 — AST 미작성): `ErrPlaceUnsupported` → `classify.go:153` `ReasonUnsupportedOrderType`(주석 「all local, all provably unsent」) → `submit` 의 `default` 갈래(`exitloop.go:1480-1486`) → `alertProposalRefused`(dedup 키 `event|position|action|level`, `:1784-1785`) + `release`. 오늘 LIMIT 은 가끔 통과한다(6번째 시도 체결); 이 change 이후는 **0건**이다. §0.3 의 엄격한 약화.
- 문서 쪽: proposal `:114` 「**차단은 두 줄이다**」가 3판에도 남아 있고, tasks §1·§2 에 `internal/trading` 작업이 없으며, spec 델타 `spec.md:39` Scenario 「통과하고 브로커로 전달된다」는 HEAD 에서 **충족 불가**다.
- 고치는 길이 공짜가 아니다: `failclosed.go:49-52` 주석이 「internal/trading is not ours to change — design D1」이라 적고, 같은 술어가 사람 CLI(`cmd/tossctl/order.go:229` → `tradingService.Place`)와 `PreviewPlace` 경고 문구(`service.go:112`)를 함께 지배한다. 바꾸면 upstream 상속 동작(제품 fork 규칙·불변식 3)이 엔진 토글과 무관하게 바뀐다. **design 이 (가) internal/trading 수정과 그 upstream 영향 처분, (나) 우회 경로, (다) 아래 P1-3 의 LIMIT 대안 중 무엇인지 정해야 freeze 가 가능하다.**

### P1-1 — §5.1 실측은 출하된 도구로 **실행할 수 없다** (착수 게이트가 교착)

- 사람이 KR MARKET 매도를 낼 수 있는 경로는 `tossctl order place`(`cmd/tossctl/order.go:229`)와 ops 쓰기(`internal/ops/write_operations.go:191`)뿐이고, 둘 다 `trading.Service.Place` → `service.go:188` 에서 **브로커에 닿기 전에** 막힌다. 조건주문(`order.go:538-578`, MARKET 허용)은 다른 endpoint 라 `/api/v1/orders` MARKET 접수를 재지 못한다(M12·M38 은 조건주문 실측).
- a100 `ma-runbook.md` R2(`:71-78`)는 5.1·5.2 를 사람 세션 큐에 올렸으나 **실행 명령을 적지 않았다.** 그대로 돌리면 「성공·실패 모두 기록」(tasks `:88-89`)이 우리 자신의 로컬 관문을 브로커 응답으로 기록하는 **거짓 음성**이 된다.
- 즉 착수 조건(tasks `:35-36` 「§1~§3 의 착수 조건은 §5 실측 기록」)이 P0-1 의 변경을 **선행 조건으로 요구**한다. 측정 전용 경로(별도 probe 도구 또는 실측 한정 빌드)와 그 승인 절차를 tasks §5 에 명시해야 한다.

### P1-2 — 이 문서가 서 있는 근거 셋이 이미 반증됐는데 본문에 그대로 있다

3판 머리말(`proposal.md:20-21`)은 Phase 1 처분만 다뤘고, 2026-08-06 「2차 리뷰」가 거짓으로 판정한 주장이 본문에 원문 그대로다.

| 위치 | 주장 | 반증 (2차 리뷰, HEAD 재확인) |
| --- | --- | --- |
| `proposal.md:114` | 차단은 두 줄 | P0-1 |
| `proposal.md:74-90`, design D5 `:187` | StockOS 가 이미 검증 | 2차 C2 — StockOS 의 평범한 손절 1차 제출은 LIMIT, MARKET 은 0.5% 초과 이탈 에스컬레이션. a087 은 임계 없이 1차부터 MARKET 이라 **인용 선례보다 엄격히 공격적** |
| `proposal.md:192` | 거부가 사라지고 **체결이 보장된다** | 2차 C3 — 시장가는 접수를 보장할 뿐 갭다운·하한가·거래정지·얇은 호가에서 체결을 보장하지 않음 |
| `proposal.md:209-210` | 엔진 청산 루프는 정규장 기준이라 차단 요소 아님 | 2차 H1 — HEAD 재확인: `InRegularSession`(`clock/market.go:146`) 비시험 호출자 = `verifylive/hours.go:130` 하나, `internal/app/engine` 0 |

High-risk change 가 자기 리뷰가 기각한 정당화를 단 채 얼 수 없다. 정정하거나 철회하라.

### P1-3 — 최강 대안이 한 번도 비교되지 않았고, 비교할 계기도 tasks 에 없다

- 2차 리뷰 C3·「3차 교정 방향」 2 는 **거래소 하한가 지정가를 보호 청산의 가격으로** 쓰라고 했다(`flatten` 이 이미 그렇게 함 — `internal/flatten/liquidate.go:374-377`). D2a 는 이것을 관측가·기준선 **뒤의 셋째 폴백 단**으로 좁혔고(design `:74`), 그 폴백의 도달 불가로 반증됐다. **원안(보호 제안에서 관측가 대신 하한가)은 평가된 적이 없다** — proposal `:143-144` 가 스스로 「9분 사건은 Phase 1 이 고치지 않는다」고 적은 것이 그 증거다.
- 원안은 (KR 한정) 세 관문을 다 통과하고(여전히 LIMIT — P0-1 소멸), 온그리드라 245,750 사건을 없애며, 원장 가격을 남긴다(D4·§3 화면 작업 불요). 비용도 적어야 한다: KR 전용(`client/marketdata.go:97`), 보호 청산마다 `PriceLimits` GET 1회(§0.4), 시간외단일가 가격 범위 밖일 가능성 `[미측정]`.
- 「3차 교정」 3(청산측 슬리피지 계기 — 현재 `slippagePct` 는 진입 전용)과 4(MARKET vs 하한가 지정가를 bps 데이터로 결정)가 tasks 에 **없다.** §5.1 은 접수만 재고 체결 품질을 재지 않으므로, 실측이 통과해도 MARKET 을 고를 근거는 생기지 않는다.

### P1-4 — 실측이 반증하면 무엇이 무효가 되는지 문서가 말하지 않는다; 측정 범위 ≠ spec 범위

- **세션 경계의 예상 응답이 틀렸을 가능성.** openapi `POST /api/v1/orders` 422 예시에 `order-type-not-allowed`(「현재 사용할 수 없는 호가 유형」)가 따로 있다. proposal `:209`·tasks `:90-91`·runbook `:76` 은 `order-hours-closed` 만 상정한다. 저장소 Go 코드의 `order-type-not-allowed` 참조 0(`rg`), `refusal_code.go:46-48` 목록에도 없다.
- **SHALL NOT 지정가의 반대편.** spec `spec.md:11` 이 보호 청산의 지정가를 금지한다. 브로커가 MARKET 은 거부하고 LIMIT 은 받는 구간(KRX 시간외단일가는 시장가 없음 — proposal 스스로 인정; SOR/확장 세션·US 프리/애프터 `[미측정]`)에서 보호 청산은 **제출 가능성이 0** 이 된다. 엔진에는 세션 게이트가 없다(P1-2 표 넷째 줄). 5.2 가 거부를 돌려주면 「세션별 유형(정규장 MARKET, 그 밖 LIMIT)」으로 spec 을 좁혀야 하는데, 그 분기와 무효화 대상(spec Requirement 1 의 SHALL NOT, tasks 2.4 의 「가격을 읽지 않는다」)이 문서에 없다. **실측 결과 → 처분 표를 design 에 넣어라.**
- **US 는 측정 없이 착지한다.** spec Requirement 1 은 시장을 가리지 않는데 실측은 KR 뿐이고 runbook `:101` 은 「US 는 범위 밖」. `[미측정 — US MARKET 매도 실주문 없음]` 태그(proposal `:160`, design `:198`)는 게이트가 아니다. US 를 Phase 2 에서 빼거나 US 실측을 게이트에 넣어라.

### P2-1 — Why 의 「9분 무보호」는 a089 감사가 반증했다

a089(아카이브 `2026-09-28-a089-an-unserved-stop-is-counted/proposal.md`)가 `exit_events` 15행·`engine.log` 전수로 「말할 수 없는 것: 그 9분 28초가 연속 무보호였다는 것」이라 적었고, 7분 42초 공백은 **주문 가능한 제안이 없었던 구간**이다(`snapshot.go:245` `|| s.Orderable`). a087 은 여전히 「포지션은 9분간 손절 없이 있었고」(`proposal.md:38`), spec 근거 「포지션이 그동안 무보호로 남았다」(`spec.md:13`)라고 쓴다. 척도를 a089 처럼 **횟수(6회 결정 중 5회 미제출)**로 바꿔라. 결론(5회 400 거부)은 그대로 선다.

### P2-2 — D3 게이트 분기의 위치·수량 검사·인용 오독

검증 필요 — AST 미작성(`checkOrderShape` 손 읽기).

- design `:148-153` 의 「sell + market → 통과. 단 Price != 0 이면 거부」를 `failclosed.go:84` 자리에 조기 반환으로 넣으면 그 뒤의 KR 통화 검사(`:88`)와 수량 양수 검사(`:91`)를 건너뛴다. 새 분기는 그 검사들 **뒤**에 서거나 같은 검사를 반복해야 한다고 design 이 명시해야 한다.
- 비분수 MARKET 의 **정수 수량**을 아무 관문도 검사하지 않는다(openapi: 소수 수량은 US 시장가 매도 전용, 그 외 400). 넣을지, 브로커 400 에 맡길지 적어라.
- design `:155-158` 은 `checkOrderShape` 주석의 「strict subset filter」(`failclosed.go:51-52`)를 이중 확인의 근거로 인용하는데, 그 주석이 실제로 말하는 것은 「진짜 검사는 서비스에 있다」 — P0-1 이다. 그리고 `orderintent` 정규화(`intent.go:140-142`)는 `NormalizePlace` 안에 있어 `sellIntent` 경로에 **없다**(`sellIntent` 는 구조체 리터럴로 조립 — `exitloop.go:1714-1722`). 그러므로 `Price != 0` 거부는 「두 번째 확인」이 아니라 **이 경로의 유일한 확인**이다. 문구를 고쳐라(결론은 옳다).

### P2-3 — 시장가가 더 나빠지는 경계 조건이 설계에 없다

design 「검증」(`:208-218`)과 본문 어디에도 다음이 없다: 하한가 도달·매수 호가 공백(MARKET 매도도 체결 안 됨 — 하한가 지정가 대비 이득 0), VI 단일가 구간, 부분체결 잔량의 처리(MARKET 잔량이 남는지·`filldetect` 가 어떻게 닫는지), 1억 이상 주문(`orders_write.go:123` `ConfirmHighValueOrder: false` → `400 confirm-high-value-required` — 유형 무관이지만 MARKET 의 금액 평가 기준 `[미측정]`). 각 행을 「다룬다 / 범위 밖 + 사유」로 열거하라. 1차 리뷰 A14(부분체결 잔량)의 행방도 여기서 정해진다.

### P2-4 — a100 D8 이중 매도 계약과의 상호작용이 a087 쪽에 없다

a100 design `:410-430` 은 계약 1 을 「a087 Phase 2 가 착지하면 창은 줄어든다」로 개정해 a087 을 **전제로** 삼았다. 계약 2 는 「상주 조건주문 취소 확인 실패는 인프로세스 매도를 막지 않는다」이다. 인프로세스 매도가 MARKET 이 되면 그 매도는 즉시 체결되고, 취소가 확인되지 않은 상주 SINGLE+MARKET 이 같이 발동하는 창(M13)의 의미가 바뀐다(브로커 매도가능수량 예약 M29 가 둘째 주문을 막는지 `[미측정]`). a100 은 HOLD(`f3a061e5`)라 착지 순서가 열려 있다 — a087 design 이 이 상호작용을 다루거나 a100 으로 명시 이연해야 한다.

### P2-5 — 인용 좌표가 전부 낡았다; FLM 은 재생성 대상

- 측정: `sellIntent`·`isProtective` 본문은 base `102d4e99` 와 HEAD 에서 **바이트 동일**(함수 추출 diff). 내용 주장은 유지된다. `failclosed.go:84` 도 정확(base 이후 `AllReasonCodes` +6 줄뿐).
- 좌표: proposal `exitloop.go:1452/1251/1486/1217/1571-1578`·design `:765/:1190/:1301/:1382` → HEAD 는 주석 `1683-1696`, `risk.Intent` `1418-1424`, `OrderType: "limit"` `1718`, `isProtective` `1382-1384`, 가격 사다리 `1698-1705`, `observe` `Last <= 0` `808`, `ObservedPrice` `1245`, `submit` 호출 `1360`, `sellIntent` 호출 `1442`. 좌표는 심볼 상대로 바꿔라.
- `analysis/function-logic/internal-app-engine--exitobserver.sellintent/ast.json` 은 `start.line 1571`·base 의 `source_sha256` 이다. task 2.1 은 이것을 재사용하지 말고 HEAD 에서 재생성해야 한다(본문 동일이라 분기 집합은 같을 것 — 검증 필요).

### P3-1 — 후속 change 장부가 존재하지 않는 대상을 가리킨다

tasks `:104-109`·proposal Non-goals 는 a088(호가 그리드)·a089(재가격)를 후속으로 둔다. a088 은 `openspec/changes`·`archive` 어디에도 없고, a089 는 다른 내용(「나가지 못한 손절은 세어진다」)으로 2026-09-28 불구현 아카이브됐다. 1차 리뷰 A1~A14 와 I2(400 분류)는 소유자가 없다 — a094 R1 이 가져간 것은 `opposite-pending-order-exists` 하나(`refusal_code.go:46-48`). design D5 의 「a089로」 행도 같다. 「무소유」로 고치거나 새 소유자를 적어라.

### P3-2 — `isProtective` 의 소비자가 셋이 됐다

base 이후 `isProtective` 는 재판정 보류 예외(`exitloop.go:1279`)·a091 floor 의미(`:1405`)에 쓰이고, a087 이 유형 결정을 더하면 셋이다. D1 의 「술어 하나」 논거는 더 강해졌다. risk-pattern-report 에 세 소비자를 적고, a087 시험이 술어 자체를 바꾸지 않음을 고정하라.

### P3-3 — 확인된 것(재검증 통과)과 작은 공백

- 전송: `buildOrderCreate`(`orders_write.go:140-150`)와 사본 `PlaceWireBody`(`wirebody.go:91-99`) 모두 MARKET 에 가격·TIF 를 싣지 않는다. 표류 가드 골든은 **US** market sell 만 있다(`decision_test.go:759-766`) — KR market sell 행 추가 권고. proposal 은 사본(`wirebody.go`)과 재생 경로를 표면으로 적지 않았다.
- 원장: `intents.price` 는 시장가 NULL 허용(`journal/schema.go:212`), `priceString` 빈 값 → NULL(`gateway.go:1041`). NULL 소비부 표본 셋 통과 — `fills.go:1858` `coalesce`, `exit_held_proposal.go:111-116` `workingOrderPrice`(a094 가 「a087 대비」로 명시), `indoubt.go:697` 가격 와일드카드·`:545-555` notional 0(`:491` 가드). **전수 아님** — tasks 3.x 에서 열거할 것.
- `risk.Intent` 에 가격 없음(`exitloop.go:1418-1424`) — 유지. 자동 매도 조립은 `sellIntent`·`flatten` 둘뿐이고 a112 전략 레인은 매수 전용(`strategydispatch/adapters.go:93-107`) — 범위 누락 없음.

### freeze 재개 조건

P0-1·P1-1~P1-4 를 문서로 닫는다: 셋째 관문 처분 결정(사람 결정 포함 가능), 실측 실행 경로와 결과→처분 표, 반증된 근거 정정, 하한가 지정가 원안과의 비교(또는 사용자 기각 기록), US 처분. P2 는 같은 개정에서, P3 는 구현 착수 전까지. 개정 뒤 Eng 재리뷰 1회.

### 사용자 결정 (2026-10-10)

재리뷰 판정 보고 후 Manager 권장안에 사용자 「동의」:

1. **P0-1 처분 = a087 범위 확장.** `internal/trading` `placeIntentSupported` 의 비지정가
   거부를 failclosed 와 동일한 모양 — **sell+market·가격 없음** — 에만 연다. 사람 CLI
   (`tossctl order place`)에서도 시장가 **매도**가 가능해지는 upstream 동작 변경을 수반함을
   알고 결정(청산 즉시성 강화 = 보수 방향, 불변식 6). `buy+market` 거부는 시험으로 고정.
2. **5.1 실측 수단 = 별도 실측 도구 선행.** `internal/official` 직접 호출(토큰 캐시 공유)
   하는 독립 도구(새 파일만)로 브로커의 KR MARKET 매도 수용을 **production 변경 전에**
   측정한다. 주문 생성은 mutating — **실행은 사람 터미널 + 주문별 승인**, 에이전트는 준비까지.
   측정이 반증하면(수용 거부) a087 Phase 2 전제가 죽고 production 은 만지지 않는다.
   대상 수량은 010170 매수 산식에 이미 포함(5.1 매도 1주·5.2 경계 1주).

후속: 문서 수리(P0-1·P1 전부, P2 동일 개정, P3 구현 전) → Eng 재리뷰 1회 → §5 실측(장중,
사람) → 결과→처분 표에 따라 §1~§3.

### 처분 대조표 (2026-10-10 문서 수리 — tasks 0.6, 3차 Eng 재리뷰 대조용)

production 편집 0. 「문서」= 본문 정정, 「task」= 구현 task 추가, 「issues」= 미해결 기록.

| # | 처분 | 위치 |
| --- | --- | --- |
| P0-1 | 문서+task — 「차단은 세 곳」 정정, ③ `placeIntentSupported` sell+market·가격 없음 개방(사용자 결정 1), Pre-Edit·FLM 선행 task, 종단 회귀 시험, spec Requirement 2 를 두 관문으로; 불변식 3 대조는 issues | proposal 「차단은 세 곳이다」·§2·Impact · design D3 ③ · tasks 1.5.0~1.5.5·1.6 · spec Req 2 · issues I-R1 |
| P1-1 | 문서 — 5.1 실행 수단 = `tools/a087-market-sell-probe/`(사용자 결정 2, 내용 미선취), 보류 블록 갱신, 결과 → 처분 표 | tasks §5 · proposal 「실측 필요」 1 · issues I-R5(runbook) |
| P1-2 | 문서 — 「StockOS 가 이미 검증」·「체결 보장」·「정규장 기준」을 반증 사실과 함께 정정(D5 첫 행 승계 → 이탈 포함) | proposal 머리말·「StockOS 대조」·Impact §0.3·「실측 필요」 2 · design 「문제의 형태」·D5 · spec Req 1 근거 |
| P1-3 | 문서+task — 하한가 지정가 원안 미평가 명시, 비교표를 기각 기록으로(사용자 결정이 시장가 택함), 슬리피지 계측 task; bps 재비교는 issues | design D6 · tasks 3.8 · proposal Non-goals flatten · issues I-R3 |
| P1-4 | 문서+task+issues — `order-type-not-allowed` 를 5.2 예상 집합에, 세션 교집합 표와 무효화 대상 명시, 엔진 세션 게이트 = **미해결, §1~§3 착수 전 확인**(tasks 0.5), US 미측정 착지 한계를 proposal 에 | design D7 · tasks 0.5·5.2·§5 표 US 행 · proposal 「US 한계」 · issues I-R2 |
| P2-1 | 문서 — 「9분 무보호」→ 「6회 결정 중 5회 미제출」(a089 감사 인용) | proposal Why · spec Req 1 근거 |
| P2-2 | 문서+task — 새 분기는 조기 통과 금지·통화·수량 검사 생존(검증 필요 표기), 「두 번째 확인」→ 「유일한 로컬 확인」, 정수 수량 처분 task | design D3 ② · tasks 1.0·1.0.1·1.1·1.1a |
| P2-3 | 문서+task+issues — 경계 조건 5행 「다룬다/범위 밖+사유」, 부분체결 잔량 task(A14 행방), VI·1억 금액 기준 미측정 | design D8 · tasks 3.6 · issues I-R4 |
| P2-4 | 문서+issues — a100 으로 명시 이연, a100 착수 시 재대조; a100 문서 반영은 범위 밖 | design D9 · issues I-R5 |
| P2-5 | 문서+task — 현재 인용 좌표를 심볼 상대로, base 좌표 절은 「base 102d4e99 기준」 표기, FLM 은 HEAD 재생성(재사용 금지) | proposal·design 본문 · tasks 2.1 |
| P3-1 | 문서 — a088·a089·A1~A14·I2 를 **무소유**로 | proposal 머리말·Non-goals · design D5 · tasks 「후속 change」 |
| P3-2 | 문서+task — 세 소비자 명시, risk-pattern-report 기재, 술어 불변 시험 | design D1 · tasks 2.1·2.2 |
| P3-3 | 문서+task — 사본 `PlaceWireBody`·재생 경로를 Impact 표면에, KR market sell 골든 행, 재생 시험, NULL 소비부 전수 열거 | proposal Impact · tasks 4.5·4.6·3.7 |

---

## 3차 proposal-freeze 재리뷰 (2026-10-10, 적대적 Eng)

- **대상**: 문서 수리 `b6ee1e26`(change 디렉터리 6파일) + 실측 도구 `dfebb590`(`tools/a087-market-sell-probe/` 10파일). HEAD `dfebb590`
- **보이스**: Claude Eng(독립·적대적, 이전 리뷰·수리 미참여) 1. Codex·CEO 미실행 — `[subagent-only]`
- **판정**
  - **A. 문서 수리 → freeze 재개 불가.** P1 2 · P2 5 · P3 4. 13건 중 완전 이행 9, 부분 이행 4(P0-1·P2-1·P2-2·P3-1).
    남은 수리는 좁다 — P1 두 건(I-R1 처분 + 사용자 재확인, §5 「수용」 기준)을 닫으면 전면 재리뷰 없이 Manager 대조로 freeze 가능하다고 본다.
  - **B. 실측 도구 → 조건부 적합(수리 1건).** P1 1 · P2 1 · P3 6. 이중 주문 경로(재시도·리다이렉트·연결 재사용·토큰 재사용·동시 실행)와
    영수증 fail-closed 는 **발견 0**. P1 은 도구 본체가 아니라 같은 디렉터리의 변이 하네스가 실돈 도구 소스를 제자리에서 바꾸는 것이다.
- **I-R1 판정 (한 줄)**: 현 설계(③ 무조건 개방)는 불변식 3 과 **양립 불가**. 완화 = ③ 개방을 엔진이 만드는 `trading.Service` 인스턴스에만
  두는 생성자 옵션(토글·config 무변경) — 그러면 CLI·ops·MCP 는 upstream 과 바이트 동일하고 a087 의 목표는 그대로다.

### 본 범위·방법·한계

- 읽음: proposal·design·tasks·spec·issues 전문, review.md 2차 절·사용자 결정·처분 대조표, `analysis/gate-0.4-codegraph.md`, 도구 소스·시험 10파일 전문,
  `internal/trading/service.go`(`placeIntentSupported`·`PreviewPlace`·`Place`·`guard`), `internal/execgw/failclosed.go`(`CheckPlace`·`checkOrderShape`),
  `internal/official/{client.go,auth_headers.go,orders_write.go}`, `internal/ops/write_operations.go`, `cmd/tossctl/{mcp.go,order.go}`, `internal/orderintent/intent.go`
  `NormalizePlace`, `docs/migration/openapi.latest.json` `POST /api/v1/orders`(응답 코드·예시·주문 상태 어휘), a100 `ma-runbook.md` 머리·R2,
  verify-execution-capability `measurements.md` M14·M35.
- upstream 대조: `git show upstream/main:internal/trading/service.go`(ref `0d49e218`, 2026-08-01) — `placeIntentSupported` 가 HEAD 와 본문 동일(비분수 = limit 전용).
  함수 이력의 저자는 전부 upstream 저자다(`git log -L`) — ③ 은 upstream 상속 코드다.
- 실행: `rtk proxy go test -count=1 -race -v ./tools/a087-market-sell-probe/...` → rc=0, 최상위 28 · 부분시험 포함 RUN 51 · PASS 51 · SKIP 0.
  `openspec validate a087-… --strict --no-interactive` → rc=0.
- **하지 않은 것**: 실 API 호출 0, 도구 `--execute` 0, `mutants.sh` 미실행(도구 소스를 제자리 편집하므로 — 아래 B-P1; 「13/13 CAUGHT」 주장은 미검증).
  **AST 산출물 미작성** — `checkOrderShape`·`placeIntentSupported` 의 분기 순서를 근거로 쓴 줄은 손 읽기이며 「검증 필요(AST 미작성)」로 표기한다.

### A. 문서 수리

#### 처분 대조표 실물 대조 (13건)

| # | 대조표 주장 | 실물 | 판정 |
| --- | --- | --- | --- |
| P0-1 | 세 곳 정정·③ task·종단 시험·spec Req 2·I-R1 | 전부 있음(proposal 「차단은 세 곳이다」, design D3 ③, tasks 1.5.0~1.5.5·1.6, spec `:33-45`, issues I-R1). 단 ③ 의 상류를 「사람 경로(CLI·운영 쓰기)」로 적었고 **MCP 에이전트 표면이 빠졌다**(A-P1-1) | 부분 |
| P1-1 | 5.1 수단 = 도구·결과→처분 표 | tasks §5 머리·표 `:170-178`, proposal 「실측 필요」 1 | 이행 (표 내용 결함은 A-P1-2) |
| P1-2 | 반증 근거 셋 정정 | proposal 「StockOS 대조」·Impact §0.3 취소선·「실측 필요」 2, design 「문제의 형태」·D5, spec Req 1 근거 | 이행 |
| P1-3 | D6 기각 기록·3.8·I-R3 | 있음. 다만 「사용자 기각」은 추론이다(A-P2-3) | 이행(형식) |
| P1-4 | `order-type-not-allowed`·D7·0.5·US 한계 | 있음 | 이행 |
| P2-1 | 「9분 무보호」→ 횟수 | Why·spec 근거는 고침. **`proposal.md:78` 「그 대가가 위의 9분이다」 잔존** | 부분 |
| P2-2 | 조기 통과 금지·유일한 로컬 확인·1.1a | ② 에만 적용. **③ 의 RED/GREEN(tasks 1.5.2·1.5.3)에는 같은 규칙·통화 행이 없다**(A-P2-2) | 부분 |
| P2-3 | D8 5행·3.6·I-R4 | 있음 | 이행 |
| P2-4 | D9·I-R5 | 있음 | 이행 |
| P2-5 | 좌표 심볼 상대·base 표기·2.1 재생성 | 있음(base 좌표 절마다 표기 확인) | 이행 |
| P3-1 | a088·a089 → 무소유 | proposal·tasks·D5 는 고침. **`design.md:339` 「호가 그리드 정본화. a088」 잔존** | 부분 |
| P3-2 | 세 소비자·술어 불변 시험 | design D1·tasks 2.1·2.2 | 이행 |
| P3-3 | Impact 표면·4.5·4.6·3.7 | 있음(3.7 방법 결함은 A-P3-3) | 이행 |

#### A-P1-1 — I-R1: ③ 무조건 개방은 불변식 3 과 양립하지 않는다. 사용자 결정의 범위도 실제보다 좁게 기록됐다

**사실** (손 읽기 — 술어 내부 순서는 검증 필요, AST 미작성):

- `placeIntentSupported`(`internal/trading/service.go:269-288`)는 upstream 상속 코드이고 upstream 에서도 비분수는 limit 전용이다(upstream/main 동일 본문).
  이 술어를 거치는 `trading.Service` 인스턴스는 셋이다 — 엔진(`internal/app/engine/engine.go:401`), CLI 앱(`internal/app/app.go:158` → `cmd/tossctl/order.go:229`),
  **MCP 서버**(`cmd/tossctl/mcp.go:69`). ops `place_order`(`internal/ops/write_operations.go:92-107`, handler `:133-191`)는 MCP 카탈로그로 노출된다
  (`internal/mcp/catalog.go:26-36`) — 즉 **에이전트가 부르는 표면**이다. `NormalizePlace` 는 `market` 을 이미 받는다(`intent.go:110`), CLI `--type` 도
  「limit or market」(`order.go:418`). ③ 이 열리면 CLI·`tossctl ops`·MCP 에서 KR/US 비분수 시장가 매도가 브로커에 닿는다.
- 엔진·CLI·MCP 가 같은 `cfg.Trading` 을 공유한다(`config.Trading`, `internal/config/service.go:23-32`) — config 토글을 새로 만들어도 사람 경로와 엔진을
  가르지 못한다.

**판정**: 불변식 3 의 취지는 「TossOS 가 더한 동작은 토글 뒤에 있고, 토글이 꺼지면 흔적이 없다」이다. 이 변경은 어떤 토글 뒤에도 없고, 모든 토글이
OFF 여도 upstream 의 `ErrPlaceUnsupported` 가 브로커 전송으로 바뀐다 — 문언과 취지 모두 위반이다. 사용자 결정(본 문서 「사용자 결정 (2026-10-10)」 1)이
이것을 알고 수용했으므로 **승인된 예외일 수는 있지만**, 그대로 freeze 할 수 없는 이유가 셋이다.

1. 결정 기록이 범위를 「사람 CLI(`tossctl order place`)」로 적었다. 실제 범위는 **에이전트 MCP `place_order`** 를 포함한다(spec `:37` 도 「사람 경로(CLI·운영 쓰기)」).
   사람이 동의한 대상과 바뀌는 대상이 다르다.
2. 결정의 근거 「청산 즉시성 강화 = 보수 방향, 불변식 6」은 **자동 보호 청산**의 논거다. 사람·에이전트가 임의로 내는 시장가 매도는 손절·익절·사이징
   변경이 아니라 주문 능력 확장이라 불변식 6 이 덮지 않는다.
3. 같은 목적을 불변식 3 위반 없이 이루는 길이 비용 거의 0 으로 있다.

**완화 조건 (택1, 사람 결정)**:

- **(권장) 인스턴스 한정 개방** — `trading.Service` 에 생성자 옵션(예: `WithProtectiveMarketSell()`)을 두고 엔진(`engine.go:401`)만 켠다. `placeIntentSupported`
  의 기본 동작·`PreviewPlace` 문구·CLI·ops·MCP 는 upstream 과 동일. 엔진이 꺼져 있으면(기존 엔진 토글) 흔적 0 → 불변식 3 충족. 엔진 경로의 ② 는 그대로
  `buy+market` 을 막는다. 시험: CLI·MCP 인스턴스에서 `sell+market` 이 여전히 `ErrPlaceUnsupported`(upstream 동등 고정), 엔진 인스턴스에서만 지원.
  spec Req 2 `:37` 의 「사람 경로에도 적용된다」 삭제, tasks 1.5.5 를 「사람·에이전트 경로 무변화」 시험으로 뒤집는다.
- **(대안) 명시 예외** — 사용자가 「불변식 3 예외: CLI·`tossctl ops`·**MCP 에이전트** 경로에서 비분수 시장가 매도 허용」을 문언으로 재확인하고, 그 문장을
  `.claude/CLAUDE.md` 가 아니라 이 change 의 Pre-Edit(1.5.0)·spec Req 2 에 범위째 기록. 이때 1.5.4 문구 갱신 대상에 `ops/write_operations.go:94`
  `place_order` Summary(「limit or US fractional market」)를 더한다(A-P3-2).
- **(대안) design D6 원안 재개** — KR 은 ③ 무변경으로 끝난다(A-P2-3).

#### A-P1-2 — §5 표의 「수용 = 2xx 접수」는 거짓 양성을 낸다

openapi 는 주문 상태 어휘에 `REJECTED`(「브로커가 주문을 거부한 상태」)를 두고, **200 으로 생성된 MARKET 주문이 REJECTED 로 끝나는 예시**까지 싣는다
(`docs/migration/openapi.latest.json` 주문 조회 예시 — `"orderType": "MARKET", … "status": "REJECTED"`). `POST /api/v1/orders` 의 성공 응답은 `200`
하나뿐이다(응답 키 `200·400·401·409·422·429·500` — `201` 없음). 그런데 tasks §5 표 `:174` 는 「**수용 — 201/2xx 접수** → Phase 2 전제 성립 … §1~§3 착수 가능」이고,
도구는 설계상 주문 상태를 읽지 않는다(`receipt.go:180-181` 안내문 「fill status is NOT measured here」). 2xx 뒤 비동기 거부가 나면 표는 High-risk production
변경을 연다. **수리**: 「수용」 = 2xx **그리고** 사람이 `tossctl orders list/completed` 로 읽은 상태가 `PENDING`·`PARTIAL_FILLED`·`FILLED` 중 하나(영수증
orderId 로 대조). `REJECTED` 는 사유에 따라 「거부」 또는 「판정 불가」 행. `201` 표기는 `200` 으로.

#### A-P2-1 — spec Req 2 의 전칭 「매수 시장가는 모든 관문이 거부」는 현행과 모순

spec `:33` 「매수 시장가 주문은 모든 관문이 계속 거부해야 한다(SHALL)」와 Scenario 「서비스 지원 검사 단독」 `:45` 「매수 시장가는 지원되지 않는 형태」는
**US 분수 시장가 매수**(금액 기반)를 두 관문이 지금 통과시키는 사실(`service.go:273-276`, `failclosed.go:71-81`, upstream 시험
`TestPlaceIntentSupportedAcceptsFractionalMarketUS`)과 충돌한다. proposal §2 표는 「US fractional 예외 존치」라 적었으나 spec 에는 없다. 아카이브되면 정본에
현행이 위반하는 SHALL 이 들어간다. 「비분수」 한정을 넣어라. (이 문장은 수리 전에도 있었으나 수리가 「모든 관문」으로 넓혔다.)

#### A-P2-2 — P2-2 의 「조기 통과 금지」가 ③ 에는 옮겨지지 않았다

검증 필요(AST 미작성). `placeIntentSupported` 는 `OrderType != "limit"` 거부 **뒤에** KR→`KRW`, US→`KRW|USD` 통화 검사를 둔다(`service.go:281-287`).
tasks 1.5.3 「비분수 비지정가 거부에 `sell+market`·가격 없음 예외 추가」를 조기 `return true` 로 구현하면 그 통화 검사를 건너뛴다 — ② 에서 2차 리뷰가
지적한 것과 같은 결함이고, CLI·ops·MCP 경로에는 ② 가 없으므로 ③ 이 그 경로의 **유일한** 관문이다. 1.5.2 RED 표에 「`sell+market` + KR 비 KRW → 비지원」·
「US `sell+market` + 통화 KRW/USD 외 → 비지원」 행을, design D3 ③ 에 「조기 통과 금지」를 더하라. 같은 맥락에서 ② 의 마지막 `intent.Price <= 0` 거부
(`failclosed.go:96-98` 「a limit order needs a positive price」)는 `sell+market` 에 적용되면 안 되는데 design D3 ② 는 「뒤의 통화·수량 검사는 그대로」만 적었다 —
그 가격 검사가 limit 전용으로 남아야 함을 명시하라(RED 1.1 의 `sell+market` 통과 행이 잡겠지만 설계가 침묵하면 안 된다).

#### A-P2-3 — D6 「사용자 기각」은 기록이 아니라 추론이다

사용자 결정(`review.md` 「사용자 결정 (2026-10-10)」 `:335-343`)에는 하한가 지정가 원안이 한 번도 나오지 않는다. D6 비교표는 그 결정 **뒤**에 같은 수리 커밋으로
작성됐고, 「처분: 기각 (사용자 결정 2026-10-10)」은 「(가)를 골랐으니 (다)는 기각」이라는 추론이다. 2차 재리뷰의 재개 조건은 「비교(또는 사용자 기각 기록)」였다.
A-P1-1 로 ③ 개방의 비용이 커졌으므로(불변식 3), D6 표를 사용자에게 보이고 기각을 문언으로 받아라 — 특히 「세 관문」 행(원안은 ③ 무변경)과 I-R1 의 관계를 함께.

#### A-P2-4 — tasks 3.8 「원장·관측에 남긴다」와 Impact 「Schema: 없음」이 충돌

tasks `:128-129` 는 청산 슬리피지 bps 를 「원장·관측에」 남긴다. proposal Impact `:236` 은 「Schema: 없음」. 원장 기록이면 schema 변경(High-risk change 안의 새
열/표)이고, 관측 전용이면 「원장」을 지워야 한다. 저장 위치·schema 영향·체결가의 출처(`filldetect` 의 어느 값)를 정하라. 한 change 에 계측 기능을 더하는 것이
범위인지도 적어라(비례 원칙).

#### A-P2-5 — §5 표가 덮지 않는 결과 셋

1. **5.2 가 2xx 로 접수**되면(장전 동시호가 대기·익일 이월 등) 실주문이 다음 세션에 체결된다 — 표 `:177` 은 「응답 코드 기록만」이고 그 주문의 처리(취소 여부·
   010170 3주 예산 소진)를 말하지 않는다.
2. **종목 상태 사유의 거부**(VI 발동 중 단일가·투자경고·단일가 매매 종목 등)는 MARKET 유형 일반의 반증이 아닌데 `:175` 「MARKET 매도 자체를 거부하는 4xx」와
   구분 규칙이 없다. 「판정 불가」로 보내는 기준을 적어라.
3. **같은 종목의 상주 SELL 조건주문**(a100 M-A 가 같은 세션·같은 010170 에 SINGLE+MARKET 손절을 건다 — `ma-runbook.md` R0 머리·§5)이 매도가능수량을 묶으면
   5.1 이 수량 사유로 거부된다. 측정 전 매도가능수량 read-only 스냅숏(👁)을 5.1 선행 조건으로 두라(도구는 이 표면을 일부러 안 읽는다 — `static_test.go:155-162`).

#### A-P3

- **A-P3-1 낡은 문장 둘**: `proposal.md:78` 「그 대가가 위의 9분이다」(P2-1 잔존), `design.md:339` 「호가 그리드 정본화. a088」(P3-1 잔존).
- **A-P3-2 1.5.4 문구 목록 누락**: `internal/ops/write_operations.go:94` `place_order` Summary 「limit or US fractional market」, `:105`. (A-P1-1 대안 경로일 때만.)
- **A-P3-3 3.7 방법**: CodeGraph 는 심볼 단위라 SQL 열 `intents.price` 의 읽기 자리를 열거하지 못한다. SQL 문자열 grep + 저널 접근자 호출자(CodeGraph)의 합집합으로.
- **A-P3-4 D7 의 열린 질문**: 0.5 는 지금 할 수 있는 read-only 작업이고 그 답이 spec Req 1 의 SHALL NOT 을 무효화할 수 있다. freeze 를 0.5 뒤로 두는 편이
  재개정 비용이 작다(차단 아님 — 착수 조건으로 이미 묶여 있음).

### B. 실측 도구 (`tools/a087-market-sell-probe/`)

#### 발견 0 영역 (본 범위와 근거)

- **재시도**: 전송 호출 자리는 `sendOnce` 의 `.Do` 하나(`send.go:85`, 구조 시험 `static_test.go:61-101`), `execute` 안 1회·반복/고루틴 없음(`:104-133`).
  `official` 의 401 재전송 `send`(`internal/official/client.go:320-366`)를 쓰지 않음을 확인 — `AuthHeaders`(`auth_headers.go:48-61`)는 토큰·계좌 seq 만 준다.
- **리다이렉트**: `CheckRedirect → ErrUseLastResponse`(`send.go:63`), 3xx 는 `unknown`(`:159-161`), mock 307 에서 추종 0(`execute_test.go:179-181`).
- **연결 재사용 재전송**: `DisableKeepAlives`(`send.go:61`) + POST 는 `Idempotency-Key` 헤더가 없어 net/http 의 replayable 조건에도 안 걸림 — 이중 방어 확인(읽기).
- **토큰 재사용·동시 실행(같은 `--out`)**: `Lstat` 사전 검사(`main.go:202-207`) + `O_EXCL`(`receipt.go:210`) — 둘이 동시에 사전 검사를 지나도 하나만 생성,
  나머지는 `errAlreadySent`로 전송 0. 층별 시험 `execute_test.go:255-285`.
- **`--out` 우회(다른 디렉터리·다른 cwd·다른 worktree)**: 같은 토큰 = 같은 키 = 같은 바이트(`order.go:152-155` 토큰이 본문 전체를 묶음)이고, 토큰 창 5분 <
  브로커 멱등 창 10분이라 둘째 전송은 언제나 창 안이다. KR 멱등 실측 M35(같은 키+같은 본문 → 같은 orderId, 둘째 주문 없음)가 이 방어의 근거 — 단 M35 는
  LIMIT 표본이고 TTL 끝(`idempotency-ttl-edge`)은 미측정이다(5분 여유로 충분하다고 판단).
- **영수증 fail-closed**: `sending` 영수증 생성 실패 → 전송 0(`main.go:225-228`), 파일 막힘·읽기 전용 두 경우 시험(`execute_test.go:317-351`), 요청 도착 순간
  `sending` 이 디스크에 있음(`:289-313`). 최종 영수증 실패 → 화면에 전문 출력 + 종료 3(`main.go:233-237`, `:264-268`).
- **결과 분류**: 409(`request-in-progress` — openapi 409 예시와 일치)·타임아웃·5xx·3xx·2xx 무 orderId·echo 불일치 → 전부 `unknown`(`send.go:144-163`),
  본문 읽기 오류는 2xx 여도 `unknown`(`:146` 의 `TransportError` 우선).
- **시크릿**: 헤더는 이름만(`send.go:166-173`), sentinel 5종이 영수증·출력에 없음(`execute_test.go:355-393`).
- **계약 대조 강도**: 생산 직렬화기와 바이트 동일(실제 `official.Client.PlaceOrder` → mock, qty 1·2 — `contract_test.go:23-47`) + openapi 원문에서 필드 집합·
  required·enum·pattern·maxLength·정수 수량·price/TIF 부재·문구 셋을 읽어 대조(`:84-161`). 키 31자 ≤ 36, pattern 일치. 엔진이 a087 뒤 보낼 본문과 같은 모양을
  재므로 측정 대상이 옳다.

#### B-P1-1 — `mutants.sh` 가 실돈 도구의 소스를 공유 작업 트리에서 제자리로 바꾼다

`mutants.sh:58` `perl -0pi` 로 `tools/a087-market-sell-probe/*.go` 를 직접 고치고 `:69` 에서 되돌린다. `trap cleanup EXIT`(`:27`)는 SIGKILL·전원 단절·OOM 에서
돌지 않고, 변이 하나가 살아 있는 동안 같은 작업 트리에서 사람이 `go run ./tools/a087-market-sell-probe --execute …`(도구가 안내하는 바로 그 명령, `main.go:175`)를
치면 **변이체가 컴파일돼 실주문을 낸다.** 변이 목록에 M3(5xx 에 재전송 — `:37`), M7(리다이렉트 추종 = 307 본문 재전송 — `:42`), M4/M4b/M9(확인·만료 우회)가
있다 — 이중 주문과 확인 없는 전송이 정확히 그 변이들이다. 병행 세션이 같은 체크아웃을 쓰는 것은 이 저장소의 실측 관행이다(기억 「미커밋 로트는 살아 있는 peer
세션 것일 수 있다」·「병행 세션 게이트 간섭」). **수리(택1 이상)**: 변이를 `go test -overlay` 또는 임시 사본에서 돌린다(a087 P1.5 의 `7bd8f197` 이 이미 overlay 방식);
그리고 §5 실행 절차에 「`git status --porcelain -- tools/a087-market-sell-probe` 가 비고 HEAD 가 착지 커밋일 때만」(또는 착지 커밋에서 미리 빌드한 바이너리 sha
대조)을 넣는다. 도구 본체는 이 결함이 없다.

#### B-P2-1 — 영수증이 실계좌 주문 기록(orderId·원 응답 본문)을 git 추적 대상 경로에 쓴다

기본 `--out` 은 `openspec/changes/a087-…/analysis/market-sell-receipts`(`main.go:42`)이고 `.gitignore` 에 없다(`git check-ignore` rc=1). 영수증에는 orderId,
원 응답 본문 전문(`receipt.go:155-167` — 오류 `data` 칸에 매도가능수량 같은 계좌 값이 실릴 수 있다), 종목·수량·시각이 들어간다. 불변식 8 의 「계좌 개인정보」에
걸리는지는 사람 판단이지만 **판단 없이 커밋되는 경로**인 것은 결함이다. 처분을 정하라: gitignore + measurements.md 에 필요한 칸만 옮겨 적기, 또는 본문 redaction,
또는 명시 수용. (measurements.md 기존 M 계열은 orderId 를 적지 않는다 — `M14`·`M35` 확인.)

#### B-P3

- **B-P3-1 「사람만 실행」 주장이 강도보다 크다**: `stdin` TTY 검사(`main.go:184-186`)는 `script -qc` 같은 pty 래퍼로 통과되고, confirm token 은 비밀 없는 해시
  (`order.go:152-155`)라 미리보기 없이도 계산 가능하다. 실제 성질은 「본문 일치 확인 + 우발 실행 방지」다 — 문구(`main.go:185`, 커밋 메시지 「에이전트 실행 금지」)를
  그 성질로 고치거나, 에이전트 차단은 권한 체계·불변식 2 가 맡는다고 적어라. (추가 마찰을 넣으라는 권고가 아니다.)
- **B-P3-2 `refused` 가 §5 표의 「거부」와 같은 말이 아니다**: 401·403·429·매도가능수량 부족도 `refused`·종료 2·「the broker refused」(`send.go:157-158`,
  `receipt.go:182-184`)다. 표에서는 「판정 불가」다. 안내문에 「refused ≠ MARKET 반증 — error code 로 tasks §5 표를 볼 것」 한 줄을 두라.
- **B-P3-3 만료를 전송 직전에 다시 보지 않는다**: `verifyConfirm`(`main.go:195`) 뒤 `AuthHeaders`(토큰 교환 + 계좌 조회, `official` 기본 시한 15초씩)가 끼어 실제 전송은
  창 끝을 수십 초 넘을 수 있다. 브로커 10분 창 안이라 이중 주문 위험은 없음 — 「5분 유효」 문구 정확성 문제.
- **B-P3-4 키가 다르면 막는 것이 없다**: 결과 `unknown` 뒤 사람이 미리보기를 다시 돌리면 새 키 = 새 실주문이다(안내문 `receipt.go:186-187` 뿐). 5.2 가 의도적 둘째
  주문이라 전면 잠금은 맞지 않지만, 같은 디렉터리에 `sending`/`no-answer`/`unknown` 영수증이 있으면 실행을 거절하는 정도는 값싸다.
- **B-P3-5 응답 해석의 계약 고정 없음**: `readAnswer` 의 봉투 모양(`send.go:112-141`)은 mock 에서만 잰다. openapi 200·409 예시와 모양은 일치함을 읽어서 확인 —
  계약 시험(`contract_test.go`)이 요청만 대조한다는 한계 기록.
- **B-P3-6 계좌 선택이 암묵적**: `official.New` 에 `WithAccountSeq` 가 없어 `accounts[0]`(`client.go:261-270`)에서 판다 — 엔진(`engine.go:390`)·CLI 와 같은
  선택이라 결함은 아니나, 010170 이 그 계좌에 있어야 하고 영수증은 계좌를 일부러 적지 않으므로 사전 보유 확인(A-P2-5 3)이 그 공백을 메운다.

### freeze 재개 조건 (3차)

1. A-P1-1: I-R1 처분 — 인스턴스 한정 개방으로 design D3 ③·spec Req 2·tasks 1.5.x 개정, **또는** MCP 에이전트 표면을 명시한 사용자 예외 재확인.
2. A-P1-2: §5 「수용」에 주문 상태 확인을 넣는다.
3. B-P1-1: 변이 하네스를 사본/overlay 로 옮기거나 §5 실행 전제(깨끗한 트리·착지 커밋)를 tasks §5 에 적는다 — §5 실행 전까지.
4. P2 는 같은 개정에서, P3 는 구현 착수 전까지. 위 1~2 는 좁아서 Manager 대조로 닫을 수 있다(전면 Eng 재리뷰 불요 — 단 1 에서 인스턴스 옵션 설계를 택하면
   그 설계 문단만 독립 확인 1회).

### 사용자 결정 (2026-10-10, 2차) — I-R1 처분

3차 재리뷰 보고 후 Manager 권장안(선택지 1)에 사용자 「승인」:

**관문 개방 = 엔진 인스턴스 한정.** `placeIntentSupported` 의 sell+market 허용은 엔진이
만드는 `trading.Service` 인스턴스(`engine.go:401`)에 생성자 옵션으로만 연다. CLI·ops·MCP 가
쓰는 기본 생성 경로는 upstream 과 동일하게 유지(불변식 3 보존, 에이전트 표면 미개방).
2026-10-10 1차 결정이 수용했던 「사람 CLI upstream 동작 변경」은 이 처분으로 **불필요해져
철회**된다 — 사람 측정 수단은 `tools/a087-market-sell-probe/` 가 담당. upstream 동일성은
시험으로 고정한다(task 에 반영).

### 처분 대조표 (3차분 — 2026-10-10 3차 재리뷰 수리, tasks 0.8)

production 편집 0. 도구 커밋 `45167aa7`(`tools/a087-market-sell-probe/` — 하네스·시험 경로·문구, 전송·가드 로직 무변경).
「문서」= 본문 정정, 「task」= 구현 task 추가·개정, 「도구」= `45167aa7`, 「issues」= 미해결 기록, 「n/a」= 적용 안 함 + 사유.
3차 발견 전수 19건(P1 3 · P2 6 · P3 10).

| # | 처분 | 위치 |
| --- | --- | --- |
| A-P1-1 | 문서+task+issues — I-R1 = **엔진 인스턴스 한정 개방**(사용자 결정 2차). ③ 개방은 엔진 `trading.Service` 의 생성자 옵션(모양 제안: 기존 빌더 관례 `WithProtectiveMarketSell()`, 술어가 인자로 받아 `PreviewPlace`·`Place` 가 같은 값 사용 — 코드는 1.5.3). CLI·ops·MCP 기본 경로 upstream 바이트 동일. MCP 표면(ops `place_order` → `internal/mcp/catalog.go`, MCP 서버 인스턴스 `cmd/tossctl/mcp.go`)을 proposal Impact·design 에 기록. 1차 결정의 「사람 CLI 도 열림」 문장 전부 2차 결정으로 정정. **upstream 동일성 고정 시험** 1.5.5 — 기본 인스턴스 거부·`PreviewPlace` 문구 바이트 동일·ops `place_order` 거부 + 옵션 호출 자리 = `engine.go` 하나(구조 시험) | proposal 머리말 5판·「차단은 세 곳이다」·§2·Impact · design D3 머리·③ · spec Req 2 + Scenario 「사람·에이전트 경로는 upstream 그대로」 · tasks 1.5·1.5.0~1.5.5 · issues I-R1 **종결** |
| A-P1-2 | 문서+도구 — §5 「수용」 = HTTP **200** + 사람이 `tossctl orders list/completed` 로 읽은 상태 `PENDING`·`PARTIAL_FILLED`·`FILLED`. 「200 뒤 `REJECTED`」 행(유형 사유 → 거부, 그 밖 → 판정 불가), openapi 상태 enum 10개 각주(표에 없는 상태 = 판정 불가). `201` 표기 제거(openapi POST 응답 키 `200·400·401·409·422·429·500` 재확인). 도구 `accepted` 안내문이 「§5 수용 아님 — 상태 확인」을 말하게 함 | tasks §5 표·각주 · 도구 `receipt.go` `guidanceFor` |
| B-P1-1 | 도구+문서 — `mutants.sh` 를 **저장소 밖 사본**으로: 대상 커밋을 `git worktree add --detach`(훅 끔)로 `${TMPDIR:-/tmp}/a087-mutants-<pid>-*` 에 꺼내 그 사본만 변이. 시작 sha 단언(사본 HEAD == 대상 커밋·사본 깨끗함·`go env GOMOD` == 사본), 무변이 대조군 GREEN 선행, 변이 대상이 사본 안인지 단언, 사본에서 `git checkout` 복원 + sha 대조, 끝에 공유 트리 도구 소스 sha 불변 단언, 전용 `GOCACHE` + `GOFLAGS=-trimpath`, 종료 시 사본·캐시 제거. 인자 없이 돌 때 도구 디렉터리가 더러우면 거절. 부수 발견: 구 시험 둘(`static_test.go`·`contract_test.go`)이 `runtime.Caller` 경로를 써서 `-trimpath` 에서 대조군이 RED(구 커밋 `77a8e119` 에 새 하네스로 실측 — 「no source files found」) → 패키지 작업 디렉터리 상대 경로로 수리. **재측정 13/13 CAUGHT·전부 닿음**(아래 표). §5 실행 전제 1(깨끗한 트리·착지 커밋) 추가 | 도구 `mutants.sh`·`static_test.go`·`contract_test.go` · tasks §5 「실행 전제」 1 |
| A-P2-1 | 문서 — spec Req 2 를 「**비분수** 매수 시장가는 모든 관문이 거부」로 한정 + 「미국 분수(금액 기반) 시장가의 기존 허용은 바꾸지 않는다」, Scenario 「서비스 지원 검사 단독」·「매수 시장가」도 비분수로 | spec Req 2 · proposal §2 |
| A-P2-2 | 문서+task — P2-2 「조기 통과 금지」를 ③ 에 적용(design D3 ③), 1.5.2 RED 에 「`sell+market` + KR 비 KRW」·「US `sell+market` + KRW/USD 외」 행과 조기 `return true` 를 잡는 조건, ② 의 마지막 `Price <= 0` 거부가 limit 전용으로 남아야 함을 design D3 ② 에 명시(검증 필요 — AST 미작성, 1.0.1) | design D3 ②·③ · tasks 1.5.2·1.5.3 |
| A-P2-3 | 문서+issues — D6 「기각 (사용자 결정)」→ 「시장가 방향 사용자 결정(2026-10-10)에 따른 **미채택** — 원안 자체 평가는 미수행」. 「세 관문」 행 비용이 2차 결정으로 줄었음을 기록. 문언 기각 수령은 사람 판단(0.9 에서 묻는다) | design D6 제목·처분·표 · proposal 「StockOS 대조」·Non-goals `flatten` · issues I-R3 |
| A-P2-4 | 문서+task — 슬리피지는 **기존 칸 유도**(`exit_events.observed_price`·`proposed_intent_id` ↔ `fill_snapshots.average_price`·`order_id`), schema·production 코드 추가 없음, 「원장에 남긴다」 철회, 읽기 전용 저널 질의를 `analysis/` 에. intent → `order_id` 결속이 기존 칸으로 안 서면 멈추고 Impact 를 schema 변경으로 고쳐 재리뷰. 비례 원칙 문장 | proposal Impact 「Schema」 · tasks 3.8 |
| A-P2-5 | 문서 — §5 표에 3행: 5.2 가 200 으로 접수(사람이 즉시 상태 확인·취소 여부 결정·5.2 예산 소진 기록), 종목 상태 사유 거부(구분 기준: code·message 가 종목·시장 상태 → 판정 불가, 유형 → 거부, 애매하면 판정 불가), 같은 종목 상주 SELL 조건주문의 매도가능수량 묶음(M-A retain 결정과 교차 — 잔여 조건주문은 취소하지 않으므로 실측 당일 a100 runbook §2 에서 겹침 확인 후 5.1 종목을 바꾼다). 실행 전제 2 = 매도가능수량·상주 조건주문 read-only 스냅숏 | tasks §5 「실행 전제」 2 · 표 |
| B-P2-1 | 문서 — 영수증 디렉터리 **커밋 의도** 명시(Manager 지시 2026-10-10): orderId 는 계좌 개인정보 아님(허용), 토큰·계좌 번호·API 키는 도구가 쓰지 않음(헤더 값 미기록, sentinel 시험), `error.data` 의 수량 값은 계좌 식별 정보 아님. gitignore 하지 않음 | tasks §5 「영수증 디렉터리는 커밋 의도다」 |
| A-P3-1 | 문서 — `proposal.md` 「그 대가가 위의 9분이다」→ 「6회 결정 중 5회 미제출」, design 「건드리지 않는 것」 「a088」→ 무소유. 같은 잔재 「9분 사건」 표기 design 두 곳도 정정(proposal Phase 1 기록 블록 안의 한 곳은 「원문은 기록으로 남긴다」 절이라 유지) | proposal Why 뒤 문단 · design D2b·D6·「건드리지 않는 것」 |
| A-P3-2 | **n/a** — 대안 경로(명시 예외)일 때만의 발견. 2차 결정으로 기본 인스턴스 무변화라 ops `place_order` Summary 「limit or US fractional market」는 사실 그대로 — 무변경을 tasks 1.5.4 에 명시 | tasks 1.5.4 |
| A-P3-3 | task — 3.7 방법: 비시험 SQL 문자열 `rg`(`intents`·`price`) ∪ 그 질의를 감싼 저널 접근자 호출자(CodeGraph) | tasks 3.7 |
| A-P3-4 | task+issues — 0.5 를 0.9 freeze 판정 전에 하는 것을 권장으로 적음(차단 아님), 순서는 0.9 Manager 대조에서 결정 | tasks 0.5 · issues I-R8 |
| B-P3-1 | 도구 — TTY 거절 문구를 「accident guard, not proof of a human; agents are barred by the permission system and invariant 2」로, 같은 뜻의 한글 주석. 추가 마찰 없음 | 도구 `main.go` `execute` |
| B-P3-2 | 도구+문서 — `refused` 안내문에 「refused ≠ MARKET 반증 — error code 로 a087 tasks §5 표」, §5 표 「판정 불가 — 그 밖」 행에 같은 문장 | 도구 `receipt.go` `guidanceFor` · tasks §5 표 |
| B-P3-3 | 문서 — 「5분」은 검증 시점 기준이고 헤더 획득 뒤 전송이 창 끝을 넘을 수 있으나 브로커 10분 창 안(이중 주문 위험 없음). 코드 무변경 | tasks §5 「실행 전제」 4 |
| B-P3-4 | 문서+issues — 도구 잠금 대신 절차: `unknown`/`sending`/`no-answer` 뒤 재미리보기 전 `tossctl orders list/completed` 확인 필수. 도구 가드는 사람이 필요하다고 판단하면 그때 | tasks §5 「실행 전제」 3 · issues I-R7 |
| B-P3-5 | issues — 응답 봉투 해석은 mock 만 잼(요청만 계약 대조) — 한계 기록, 응답 모양 불일치 시 `unknown` 으로 기울어 fail-closed | issues I-R6 |
| B-P3-6 | 문서 — 계좌 선택 = `accounts[0]`(엔진·CLI 와 같음, 결함 아님), 영수증은 계좌를 적지 않으므로 실행 전제 2 의 사전 보유 스냅숏이 「그 계좌에 그 종목」의 유일한 확인 | tasks §5 「실행 전제」 2 |

**2차 대조표의 부분 이행 4건**: P0-1(MCP 표면) → A-P1-1 행 · P2-1(`proposal.md:78`) → A-P3-1 행 · P2-2(③ 적용) → A-P2-2 행 ·
P3-1(`design.md:339`) → A-P3-1 행. 넷 다 이번 개정으로 닫았다.

#### 변이 재측정 (B-P1-1 수리 후, 2026-10-10)

`bash tools/a087-market-sell-probe/mutants.sh` (인자 없음 = HEAD, 도구 디렉터리 깨끗함 확인) — 측정 커밋 `45167aa7`, 사본
`/tmp/a087-mutants-927332-*/tree`, 소요 19.9 s, rc=0. 무변이 대조군 GREEN · 사후 GREEN · 「shared tree: untouched」. 종료 뒤
`git worktree list` = 주 작업 트리 하나, `/tmp/a087-mutants-*` 0 개(정리 확인).

| 변이 | 닿음(파일 sha 앞 12자, 전 → 후) | 판정 | 잡은 시험 |
| --- | --- | --- | --- |
| M1 sell 가드 삭제 | order.go `6af9ec249c4b → c3504dc979c6` | CAUGHT | `TestAssembleRefusesEachGuard/{buy,uppercase_SELL_is_not_the_fixed_value}` |
| M2 수량 상한 완화 | order.go `6af9ec249c4b → 476b04ee329a` | CAUGHT | `TestAssembleRefusesEachGuard/quantity_over_the_cap` · `TestPreviewRefusesGuardedShapes` |
| M2b 상한 상수 3 | order.go `6af9ec249c4b → b515fabffae3` | CAUGHT | `TestAssembleRefusesEachGuard/{negative,zero,over_the_cap}` · `TestPreviewRefusesGuardedShapes` |
| M3 5xx 재전송 | send.go `ed2cbc3e339e → 29983e0f9533` | CAUGHT | `TestNoRetryOnAnyFailure/{500,503}` · `TestOnlySendOnceCallsDoAndItHasNoLoop` · `TestSecondExecuteWithTheSameTokenIsRefused` |
| M4 확인 검사 건너뜀 | main.go `bbe091819b19 → 8bdd6453a542` | CAUGHT | `TestExecuteRefusesAnExpiredToken` · `TestExecuteRefusesMismatchedConfirm` |
| M4b 토큰 비교 우회 | order.go `6af9ec249c4b → 3684a23d97b5` | CAUGHT | `TestConfirmTokenBindsTheWholeBody` · `TestExecuteRefusesMismatchedConfirm` |
| M5 market 가드 삭제 | order.go `6af9ec249c4b → a4cc1d6a9ba3` | CAUGHT | `TestAssembleRefusesEachGuard/US_market_with_a_KR` |
| M6 limit 가드 삭제 | order.go `6af9ec249c4b → be05c8f155e1` | CAUGHT | `TestAssembleRefusesEachGuard/limit` |
| M7 리다이렉트 추종 | send.go `ed2cbc3e339e → d30e07e11b9f` | CAUGHT | `TestNoRetryOnAnyFailure/redirect` · `TestSendClientRefusesRedirectsAndConnectionReuse` |
| M8 배타 생성 제거 | receipt.go `8f981931593e → d37493e8bd91` | CAUGHT | `TestCreatePendingReceiptIsExclusive` |
| M9 만료 건너뜀 | order.go `6af9ec249c4b → 616cf611b0bc` | CAUGHT | `TestConfirmWindow` · `TestExecuteRefusesAnExpiredToken` |
| M10 터미널 검사 건너뜀 | main.go `bbe091819b19 → 589fa51c6eee` | CAUGHT | `TestExecuteNeedsATerminal` |
| M11 심볼 가드 확장 | order.go `6af9ec249c4b → a79bd1f9f4fb` | CAUGHT | `TestAssembleRefusesEachGuard/{five_digits,seven_digits,US_ticker}` · `TestPreviewRefusesGuardedShapes` |

SURVIVED 0 → 시험 보강 없음. 하네스 자체의 대조: 구 커밋 `77a8e119`(구 시험)에 새 하네스를 돌리면 대조군 RED 로 멈춘다(위 부수 발견),
미커밋 도구 편집이 있을 때 인자 없이 돌리면 「REFUSED」로 멈춘다 — 둘 다 실측.

도구 시험 배터리(`45167aa7` 착지 직전 작업 트리에서 -v 로 셈, 착지 뒤 깨끗한 트리에서 vet·gofmt·`-race` 재실행 rc=0): `go vet ./tools/a087-market-sell-probe/` rc=0 · `gofmt -l tools/a087-market-sell-probe/`
빈 출력(양성 대조군 — 일부러 어긋난 파일은 보고됨) · `rtk proxy go test -count=1 -race -v ./tools/a087-market-sell-probe/...` rc=0, 최상위 28 ·
RUN 51 · PASS 51 · FAIL 0 · SKIP 0.

#### freeze 재개 조건 (3차) 상태

1 A-P1-1 — 인스턴스 한정 개방으로 design D3 ③·spec Req 2·tasks 1.5.x 개정 **완료**. 2 A-P1-2 — **완료**. 3 B-P1-1 — **완료**(사본 실행 + §5
실행 전제). 4 P2 6·P3 10 — 전부 처분(위 표). **남은 것**: tasks 0.9 — Manager 대조 + ③ 인스턴스 옵션 설계 문단(design D3 ③) 독립 확인 1회.
그 전 §1~§3 착수 금지는 그대로다.

---

## 0.10 requirement 변경 리뷰 — D7 7판 (2026-10-28, 적대적 Eng 관점)

- **대상**: spec Requirement 1(2026-10-22 개정 + 2026-10-28 7판) · design D7 「처분」·「구간 입력」·「§5.2 처분」 · tasks 2.4~2.6 · 5.2·5.3 · §1 착수 조건 ④.
  입력은 사용자 결정 2026-10-28 Q1~Q4(`issues.md` 「사용자 결정 (2026-10-28)」).
- **리뷰어**: 이 문서를 7판으로 고친 작성자와 같은 세션이다(작성자 겸 리뷰어 지시). 그래서 독립 리뷰가 아니다 — 착수 전 Manager 대조 1회가 이 한계를 메운다.
- **본 것**: HEAD `0c9bfd272` 의 `checkOrderShape`·`Preflight.CheckPlace`(failclosed.go) · `placeIntentSupported`·`Service.Place`·`PreviewPlace`·`guard`(trading/service.go) ·
  `ExitObserver.sellIntent`·`submit`·`isProtective`(exitloop.go) · `Context.ExitObserver`(exitwiring.go) · `NewOrderPath`(engine.go) · `clock` 의 `Location`·`RegularSession`·
  `Session.Contains`·`InRegularSession` · `official.Client.TypedMarketCalendar` · `scheduler` 의 `AdaptOfficialCalendar`·`adaptCalendarDay`·`CalendarSnapshot.ValidityAt` ·
  `strategyScheduleAuthorityLoader.prepare`. 분기 근거는 전부 `tools/logic-map` AST — `analysis/evidence/*.ast.json` 20개, 열거 `analysis/evidence/enumerate-output.md`
  (`enumerate.py` 가 각 `source_sha256` 을 현재 바이트와 대조, stale 0). openapi `docs/migration/openapi.latest.json` 의 `RegularMarketSession`·`IntegratedHour`·
  달력 예시. `trading.NewService(` 비시험 호출 자리(`rg`): `cmd/tossctl/mcp.go:74` · `internal/app/app.go:158` · `internal/app/engine/engine.go:401` — 셋.
  a165 proposal·design 3판·tasks(`bc56522c7`).
- **하지 않은 것**: 시험·게이트·`make sdd-sync` 실행 0, production 편집 0, 브로커·계좌 접근 0. gate ⑤ 함수 단위 창 계산(`check_analysis.py`) 미실행 — 아래 P2-3 은 파일 수만 쟀다.

### 판정: **PASS** — §1~§3 착수 조건 ② 충족

P0 0 · P1 2(둘 다 이 개정에서 문서로 닫음) · P2 4(닫음 1 · Manager 확인 3) · P3 6. 열린 P2 셋 중 P2-2·P2-4 는 **MARKET 칸을 좁히는 쪽이 현재 문서의 기본값**이라
확인 전에도 불변식 위반이 없다. P2-3 은 안전이 아니라 완료 게이트의 작업량 문제다. 착수 전 Manager 가 P2-2·P2-3·P2-4 셋을 처분한다.

### 불변식 판정

| 축 | 판정 | 근거 |
| --- | --- | --- |
| 불변식 3 — 토글 OFF·CLI 경로 = upstream (spec Req 2) | **유지** | ③ 개방은 엔진 인스턴스 옵션뿐이고 `NewService` 비시험 자리 셋 중 옵션을 켜는 것은 `engine.go:401`(`NewOrderPath` 성공 반환 :396) 하나로 설계됐다(1.5.5 (c) 구조 시험이 고정). 7판이 더하는 달력 갱신·판정은 `Context.ExitObserver`(엔진 청산 관측자 생성, B1~B3 이 자동화 검증 전에는 관측자를 만들지 않음) 뒤에만 선다 — 엔진이 꺼지면 달력 GET 0. CLI·ops·MCP 는 달력을 읽지 않는다 |
| 불변식 4 — 손절 즉시성 | **유지(조건부 — P1-2 의 제약을 구현이 지킬 때)** | MARKET 칸(정상일 KRX 접속매매)은 그리드·밴드 거부 클래스를 없앤다(§5.1 M50, n=1). 나머지 칸은 오늘과 같은 LIMIT. 6판에는 판정 입력의 **읽기 시점과 실패 처리**가 없어서 「제출 경로의 달력 GET」(대기)이나 「판정 오류 → `submit` B5 `:1533` 거부 갈래」(미제출)로 구현되면 LIMIT 칸조차 오늘보다 약해질 수 있었다 → P1-2 로 닫음 |
| 불변식 5 — High-risk 경로 | **유지** | 새로 닿는 High-risk 함수: `Context.ExitObserver`(배선) · `sellIntent`·`submit`(시그니처) · `NewOrderPath` · `checkOrderShape` · `placeIntentSupported`·`Place`·`PreviewPlace`. Pre-Edit 1.0·1.5.0·2.0·2.6.0 과 FLM 1.0.1·1.5.1·2.1·2.6.1·2.6.2 가 그 집합을 덮는다(아래 「착수 시 FLM 대상」) |
| 불변식 6 — 손절·익절·사이징은 보수 방향만 | **유지** | 7판은 6판보다 MARKET 칸이 좁다(US 정규장 삭제 · 비정상일 종일 LIMIT · 종가단일가 LIMIT). 익절·사이징·floor(`applyFloor`)·Guardian `IssueReduction` 무변경 |
| LIMIT 칸 = 오늘과 바이트 같은 주문 | **설계상 성립, 시험 요구 추가** | 오늘 `sellIntent`(AST B1~B4, 반환 :1823)는 관측가 → 기준선 → 거부 사다리로 `OrderType "limit"` 리터럴을 만든다. 7판의 LIMIT 갈래는 같은 사다리를 그대로 탄다. ②·③ 의 limit 갈래(② B11~B13, ③ B4·:287)는 바뀌지 않는다(1.1·1.5.2). 엔진은 `PreviewPlace` 에서 `ConfirmToken` 만 쓰고(`gateway.go:396`), `guard`(B1~B4)는 `LiveReady`·`Warnings` 를 보지 않으므로 1.5.4 의 엔진 인스턴스 문구 처리도 LIMIT 주문 바이트에 닿지 않는다. 다만 「유형만 limit」 시험은 이 성질을 못 잡는다 → 2.6.3 에 **`PlaceIntent` 전체 골든 동일** 행을 넣었다(P3-3 범위 주의) |
| 달력 결측 fail-closed | **성립(문서)** | `TypedMarketCalendar` B1·B2·B4 오류 · `AdaptOfficialCalendar` B1~B8 오류 · 오늘 `integrated` null(`adaptCalendarDay` B8 이 **오류 없이** `Regular == nil` 반환 — 휴장일 모양) · KRX 기준 필드 null · 6시간 이상 · 읽은 시각 미래 · 날짜 불일치 · 기동 직후 응답 없음 → 전부 LIMIT(D7 조건 1~5). 2.6.3 RED 가 각 행을, 2.6.5 가 「결측 → market」 등 변이를 잡는다 |
| a165 와의 충돌 | **코드 충돌 낮음 · 착지 뒤 재대조 필요** | 아래 「a165 대조」 |

### 발견

**P1-1 (닫음 — 이 개정) — 달력의 KR 정규장은 KRX·NXT 합집합이라 KRX 조기 폐장을 숨긴다.** openapi `RegularMarketSession`: 「KRX·NXT 정규장의 합집합(가장 이른
시작 ~ 가장 늦은 종료). 종가단일가 구간을 포함」. `adaptCalendarDay` 는 KR 에서 `Integrated.RegularMarket` 만 읽고(B4·B5), Go 형 `official.MarketCalendarSession`
(`typed_calendar_reads.go:11-14`)은 `startTime`·`endTime` 만 있다 — KRX 기준 `singlePriceAuctionStartTime` 을 읽지 않는다. 그래서 Q3 문언을 「달력 창 ∩ `clock` 고정 창」으로
구현하면 KRX 가 13:00 에 닫고 NXT 가 정상 운영하는 날 합집합이 정상 모양이 되어 **Q3 가 막으려던 13:00~15:30 MARKET 누출이 다시 열린다.** 처분: D7 「구간 입력」 조건 4 —
정상 영업일 모양(09:00 · 15:30 · KRX 기준 종가단일가 시작 15:20)이 아니면 종일 LIMIT, 형 확장은 2.6.4. 잔존 `[미측정]`: 브로커가 조기 폐장일에 이 필드를 실제로 바꾸는지(openapi
예시 표본 없음) — `issues.md` I-R4 (라).

**P1-2 (닫음 — 이 개정) — 판정 입력의 즉시성 제약이 없었다.** 6판 D7·2.6.4 는 「입력을 배선한다 · 실패는 LIMIT 갈래」만 적었다. ① 달력을 제출 순간에 읽으면 오늘의 LIMIT 손절이
겪지 않는 왕복이 더해진다. ② 판정 함수가 오류를 반환하면 `sellIntent` 오류 → `submit` B5(`:1533`) → `alertRefused`·`release` 로 **그 주기 보호 청산이 안 나간다**(오늘은
LIMIT 으로 나간다). ③ 갱신을 청산 관측 주기 안에서 하면 GET 지연이 `Run` 의 5초 주기 판정을 늦춘다. 처분: D7 「즉시성 제약」(제출 경로 GET 0 · 판정 실패 = LIMIT ≠ 거부 · 갱신은
관측 주기 밖), spec Req 1 SHALL NOT 두 문장, 2.6.3 RED 행(「판정 실패 행에서 거부 갈래 미도달」·「제출 경로 GET 0」), 2.6.5 변이 둘.

**P2-1 (닫음) — `ValidityAt` 을 그대로 쓰면 매일 장 끝이 LIMIT 으로 강등된다.** AST B3: 정규장 중에 장 개시 뒤 읽은 달력은 `REFRESH_TOO_LATE`. B2 6시간 상한과 겹치면 장 개시 전에
읽은 달력은 늦어도 15:00 전에 낡고, 다시 읽으면 B3 에 걸린다. 안전 약화가 아니라 이득 손실이지만 규칙을 가져올 이유가 청산에 없다. 처분: D7 조건 3(B1·B2 만, B3 미적용) + 2.6.3 행 +
2.6.5 변이.

**P2-2 (열림 — Manager 확인) — 「고정 시각」의 해석.** 문서는 「달력이 정상 영업일이 아니면 종일 LIMIT, 정상일이면 고정 창」으로 읽었다. 대안(교집합)은 조기 폐장일 오전에 MARKET 이득을
더 얻지만 P1-1 의 합집합 문제로 KRX 기준 필드 없이는 안전하지 않다. 현재 문서 해석은 MARKET 칸을 대안보다 좁히기만 한다(보수 방향). 확인 질문: 「비정상 영업일(휴장·조기 폐장·개장 지연)에는
종일 지정가로 둔다 — 권장: 그대로」.

**P2-3 (열림 — Manager 결정, 공정) — gate ⑤ 창이 크다.** 창의 시작은 `base-commit.txt` = `102d4e99`(2026-09-30)이고, base 는 proposal freeze 뒤 불변이다(`check_analysis.py`
착지 판정 docstring 「spec 이 불변을 SHALL 로 요구」). `git diff --name-only 102d4e99 0c9bfd272 -- '*.go'` = **756 파일(비시험 294)**. gate ⑤ 는 창에서 본문이 바뀐 기존 Go 함수마다
번들을 요구하므로(WORKFLOW 「비례 원칙」 단서) 착지 때 a087 이 만들지 않은 형제 함수 다수가 요구될 수 있다 — **함수 수는 미측정**(check_analysis 미실행). a165 착지가 창을 더 키운다.
선례는 a170(창에 든 형제 함수 26 을 a172 아카이브 번들 복사로 덮음). 착수 전 처분: 형제 번들 복사·빌림 방식과 범위를 Manager 가 정한다.

**P2-4 (열림 — Manager 확인) — 종가단일가(15:20~15:30)를 LIMIT 으로 뺐다.** openapi 가 정규장을 「종가단일가 구간을 포함」으로 정의하므로 6판 「KRX 정규장 = MARKET」은 이 구간을 포함했다.
§5.1 은 14:59 접속매매 한 건이라 이 구간의 MARKET 접수를 재지 않았다. 수락된 Q1 원칙(접수 미측정 → LIMIT)을 이 하위 구간에 적용해 D7 표·spec Req 1·2.6.3 을 고쳤다(오늘과 같은
동작이라 약화 없음, 하루 10분의 이득 손실). 이것은 사용자 답의 직접 문언이 아니라 원칙의 적용이다. 확인 질문: 「종가 단일가 10분은 지정가 — 권장: 그대로」.

**P3-1 — US 생산자 0 이 된 게이트 갈래.** ②·③ 의 `sell+market` 예외는 시장 무관 한 갈래라 US 에도 열린다. Q2 로 엔진의 US 시장가 생산자는 0 이다(`sellIntent` US → LIMIT, `flatten` LIMIT,
a112 레인 매수 전용). 별도 US 코드가 아니므로 도달 불가 코드 문제(I-P1 의 기각 사유)는 아니다. 1.1·1.5.2 의 US 행은 단위 행으로 남기고, 1.6 종단 시험은 KR 만 잰다.

**P3-2 — design D3 의 「두 호출자가 같은 값」 근거가 결과를 과장한다.** AST: `Service.Place` B4(`:200`)가 넘긴 `PreviewPlace` 결과에서 `guard`(B1~B4)는 `ConfirmToken` 만 비교하고
`LiveReady` 를 읽지 않는다. 그래서 옵션을 한쪽만 보면 주문이 갈라지는 것이 아니라 **미리보기 표시(`LiveReady`)만** 갈라진다. 설계 결론(술어 인자로 같은 값을 넘긴다)은 그대로 옳다 — 근거
문장만 정정 대상. 1.5.1 FLM 이 이 열거를 다시 적을 때 고친다.

**P3-3 — 「바이트 같은」의 범위.** 성립하는 것은 **주문**(`PlaceIntent` → wire body → `intents` 행)이다. 3.4(화면·알림·`obs` 의 시장가 표시)가 LIMIT 줄에 유형 필드를 더하면 로그·화면 바이트는
달라진다. 3.4 는 LIMIT 줄의 기존 필드를 바꾸지 않는다는 것을 시험 이름으로 고정한다(새 필드 추가는 허용).

**P3-4 — 호스트 시계 가정.** 조건 5 는 엔진 시계를 믿는다. 호스트 시계가 크게 어긋나면 접속매매 밖에 MARKET 이 나갈 수 있다. 엔진 전반의 가정(NTP)이라 새 가드를 제안하지 않고 잔존 위험으로
적는다. 달력의 `fetchedAt` 미래 검사(조건 3)는 역행 하나만 잡는다.

**P3-5 — 달력 GET 의 자원 공유.** 갱신은 OpenAPI 토큰·레이트 제한(a166)을 손절 경로와 공유한다. 하루 수 회 · KR 만이라 경합은 무시할 수준이지만 2.6.4 가 갱신 주기와 실패 시 재시도 간격을
명시한다(토큰 교대 사고 — `--config-dir` 없는 도구가 토큰을 번갈아 무효화한 2026-10-19 실측 — 와 같은 캐시를 써야 한다).

**P3-6 — I-R2 는 판정 입력 없이 무소유로 남는다.** `order-type-not-allowed` 분류는 「§5.2 결과 뒤」였는데 Q4 로 §5.2 가 없다. 정규장 밖 MARKET 생산이 0 이므로 차단 아님.

### a165 대조 (a165 가 먼저 착지할 계획)

a165 3판(`bc56522c7`)의 범위는 취소 409 처리(R3 엔진 · R4 flatten) + 401 근거 시험(R5)이고, 문서상 편집 표면은 `internal/execgw`(dispatch 클로저 — D5) · `internal/official` ·
`internal/flatten` · `internal/app/engine` 의 **clear 경로·배선** · `cmd/tossctl`(flatten 배선) · `internal/journal`(dispatch)이다. 엔진 결과는 `clearTheSymbol` 의 기존 갈래(B3·B11)에 기대고
그 함수 본문을 바꾼다는 문장은 없다. 엔진 게이트웨이 생성은 `buildGateway`(`internal/app/engine/gateway.go:311`)다.

| 표면 | a087 | a165 | 판정 |
| --- | --- | --- | --- |
| `exitloop.go` | `sellIntent`·`submit` 본문(시그니처) | 문서상 본문 편집 없음(clear 경로 시험·인용 좌표 `:1539`·`:1686`) | 코드 충돌 낮음. a165 가 `exitloop.go` 를 한 줄이라도 바꾸면 a087 의 줄 좌표·HEAD AST 가 낡는다 |
| `execgw` | `failclosed.go` `checkOrderShape` | `gateway.go` dispatch 클로저 | 다른 파일. a087 문서가 인용한 `gateway.go:383-384`·`:396`·`:594-595` 좌표는 a165 D5 뒤 옮겨질 수 있다 |
| 엔진 배선 | `NewOrderPath`(engine.go) · `Context.ExitObserver`(exitwiring.go) | `buildGateway`(engine/gateway.go) 추정 | 다른 함수 |
| 행동 | 보호 청산 1차 제출의 유형 | 청산 전 취소 409 → 다음 주기 재취소(오늘은 재시작까지 막힘) | 상호 보완. a165 는 「엔진 청산 매도 접수 변화 없음」(a165 D4) — MARKET/LIMIT 과 독립 |

### a165 착지 뒤 a087 이 다시 할 일

1. 새 main 에서 `python3 -I analysis/evidence/enumerate.py` 를 돌려 stale 을 센다. stale 인 AST 는 그 HEAD 에서 다시 만들고, design·proposal·tasks 의 줄 좌표(`exitloop.go`·`gateway.go`)를 고친다.
2. a165 가 바꾼 함수 중 a087 이 인용하거나 편집하는 것(특히 `exitloop.go`·`execgw` 패키지)을 열거해 D3·D7 주장을 재대조한다.
3. gate ⑤ 창(P2-3)에 a165 의 함수가 더해진다 — Manager 가 정한 방식(아카이브 번들 복사 등)으로 덮을 목록에 넣는다.
4. a165 의 엔진 취소 409 정책 시험과 a087 1.6 종단 시험이 같은 시험 더블을 쓰면 서로의 기대값을 깨지 않는지 확인한다(엔진 패키지 시험 파일 공유).
5. a171(place 재전송)은 청산 매도 접수를 범위에서 뺐다(a165 D4 표) — a171 이 그 범위를 바꾸면 MARKET 보호 청산 재전송(같은 `clientOrderId`)을 a087 D7 표와 대조한다.

### 착수 시 FLM 대상 (gate ⑤ — 본문을 고치는 기존 함수)

`execgw.checkOrderShape`(1.0.1) · `trading.placeIntentSupported`·`Service.PreviewPlace`·`Service.Place`(1.5.1 — 술어 인자 추가로 두 호출 자리 본문이 바뀜) ·
`engine.NewOrderPath`(엔진 인스턴스 옵션) · `ExitObserver.sellIntent`(2.1·2.6.2) · `ExitObserver.submit`(호출부 — 2.3) · `Context.ExitObserver`(달력 배선 — 2.6.1) ·
달력 형을 바꾸며 본문이 바뀌는 기존 함수가 있으면 그것(2.6.1) · 3.x 원장·화면 경로에서 고치는 기존 함수 · 바뀌는 기존 시험 함수(시험 파일 포함). 판정 함수와 달력 보관 갱신은
**새 파일의 새 함수**로 둔다. 창의 형제 함수는 P2-3.

Function Logic Map: not-applicable — 이 리뷰와 7판 개정은 문서만 바꿨다. 분기 주장은 `analysis/evidence/` AST(HEAD `0c9bfd272`, stale 0)와 기존 0.5·3.6 AST 에서만 인용했다.

## 0.10 독립 리뷰 (GPT-5.6 Terra, 2026-10-28)

- **대상**: 7판 proposal 머리말·design D3/D6/D7/D8·order-execution delta Requirement 1·2와 시나리오·tasks 0.10/2.6.x/5.2/5.3·issues 사용자 결정·0.10 기존 리뷰·`analysis/evidence/` AST 20개와 `enumerate-output.md`.
- **독립성**: 7판 작성자와 다른 GPT 세션에서 수행. 현재 HEAD `f8ff5a626`(증거 AST 기준 `0c9bfd272`, `enumerate-output.md` stale 0)을 현재 소스와 다시 대조했다.
- **실행 범위**: 소스·시험 미편집, 시험·gate·브로커/계좌/운영 콘솔·mutating 명령·push 실행 0. ai-memory `a087`/`protective exit` 검색은 빈 결과였고 Hindsight MCP는 이 세션에 노출되지 않아 미실행이다.
- **SDD 적용**: impeccable N/A — 백엔드 주문 형태·문서 리뷰라 UI/UX 표면 0. search-first N/A — 기능·의존성·통합을 만들지 않은 read-only 리뷰이며 저장소·OpenAPI snapshot·AST만 확인했다. grill-me는 Q1~Q4와 Manager P2-2~P2-4가 이미 사람/Manager 결정이라 새 질문 0이다.

### 판정: **FAIL (blocker)** — §1~§3 착수 금지 유지

P0 1 · P1 2 · P2 3 · P3 1. Manager P2-2·P2-3·P2-4 처분은 `issues.md` 사용자 결정 절에 기록했지만 아래 P0/P1을 닫지 않는다.

### P0-1 — `NewOrderPath`는 엔진 전용 인스턴스가 아니다

**근거:** design D3는 `engine.go`의 `NewOrderPath` 생성 자리에서만 `WithProtectiveMarketSell()`을 켜겠다고 제안한다(design.md:211-243, tasks.md:113-142). 그러나 현재 `cmd/tossctl/flatten.go:200`도 `engine.NewOrderPath(...)`를 사용한다. `NewOrderPath`는 `engine.go:401`에서 같은 `*trading.Service`를 만들고, 실제 엔진 조립은 그 뒤 `NewContext`가 `engine.go:438`에서 이 함수를 호출한다. 따라서 제안대로 :401에 옵션을 붙이면 CLI `tossctl flatten`도 시장가 허용 Service를 받는다. `app.go:158`·`mcp.go:74`의 직접 `NewService` 두 자리만 세는 1.5.5(c)는 이 우회 경로를 보지 못한다. 호출 관계는 **손으로 읽음**; `NewOrderPath` 본문 분기·반환은 `analysis/evidence/engine--neworderpath.ast.json`으로 확인했다.

**위험:** 불변식 3의 “엔진 인스턴스만” 경계가 깨진다. flatten이 현재 LIMIT만 만들더라도, 사람이 호출하는 공유 Service에 비분수 MARKET SELL 권한을 심는 설계 자체가 금지한 표면 확장이다. a165 R4도 같은 `cmd/tossctl/flatten.go` 배선을 바꾼다.

**착수 전 처분:** 옵션 부착은 공유 빌더 `NewOrderPath` 밖, 엔진 전용 조립 지점(`NewContext` 이후 등)으로 옮긴다. 1.5.5(c)는 `NewOrderPath`의 모든 비시험 호출자(특히 flatten)를 열거해 옵션 부착 0을 고정하고, 옵션을 받은 객체가 ExitObserver 게이트웨이로만 흘러감을 구조 시험으로 고정한다. a165 착지 뒤 `flatten.go`와 해당 AST를 재대조한다.

### P1-1 — 달력 형 확장만으로 KRX 접속매매를 증명할 수 없다

**근거:** OpenAPI의 KR `regularMarket`은 KRX·NXT의 합집합이고 `startTime`은 가장 이른 시작, `endTime`은 가장 늦은 종료다(`docs/migration/openapi.latest.json:9793-9822`). KRX 기준 필드는 종가단일가 시작 하나뿐이다. KRX 개장 지연 중 NXT가 09:00에 열고 KRX 종가단일가가 15:20이면 design D7의 정상 모양(09:00·15:30·15:20)을 그대로 만족한다. 장중 VI·일시 거래정지 상태도 이 스키마와 typed 형에 없다. 더구나 현 `MarketCalendarSession`은 시작·끝만 보관(`internal/official/typed_calendar_reads.go:11-14`)하고, `adaptCalendarDay`는 `Integrated.RegularMarket`을 `SessionWindow{Open, Close}`로 축약한다(`internal/scheduler/calendar.go:110-133`). 즉 2.6.4가 typed struct에 `singlePriceAuctionStartTime`만 더해도 canonical snapshot까지 전달되지 않는다. 이 분기 사실은 `scheduler--adaptcalendarday.ast.json`으로 확인했고, OpenAPI 계약과 필드 누락은 **손으로 읽음**이다.

**위험:** MARKET이 실제 KRX 접속매매가 아닌 개장 지연·VI·거래정지 시각에 나가 400 거부되거나 단일가에서 의도하지 않은 체결을 할 수 있다. D8도 VI를 미측정 잔존 위험으로 남긴다(design.md:481-483, 497-502). 그래서 “정상 영업일 KRX 접속매매만 MARKET”이라는 현 수용 기준은 구현할 입력이 없다.

**착수 전 처분:** KRX 고유 개장·현재 거래 가능/정지 상태를 권위 있게 주는 브로커 입력을 특정하고, raw 응답→검증 snapshot→판정까지 보존하는 형·RED 표를 설계한다. 그 입력을 못 얻으면 해당 불명 구간은 LIMIT으로 좁혀야 한다. 조기 폐장·개장 지연·KRX 휴장·VI/정지의 실제 응답 표본 또는 권위 문서를 붙이기 전에는 2.6.4만으로 착수 불가다.

### P1-2 — KR 보호 MARKET의 정수 수량이 계약으로 고정되지 않았다

**근거:** OpenAPI는 수량의 기본을 양의 정수로 하고, 소수 수량은 US MARKET SELL에만 허용하며 그 밖은 400이라고 명시한다(`docs/migration/openapi.latest.json:10864-10880`). 그런데 `checkOrderShape`는 `Quantity <= 0`만 검사한다(B12, `analysis/evidence/execgw--checkordershape.ast.json`; 현 소스 `internal/execgw/failclosed.go:91-92`). `sellIntent`는 원장 decimal 문자열을 `ParseFloat`로 받아(`internal/app/engine/exitloop.go:1815-1829`, 2014-2019), `buildOrderCreate`는 shortest decimal로 그대로 전송한다(`internal/official/orders_write.go:136-159`). 투영 Position도 decimal 문자열이다(`internal/journal/position_projection.go:507-510`). 정수성은 보호 진입 readiness에만 별도로 있다(`internal/execgw/protection.go:112-122`), 청산에는 없다. 이 함수 내 분기는 각각의 AST로, 자료형·wire 변환은 **손으로 읽음**이다.

**위험:** KR 위치에서 `1.5` 같은 값이 MARKET으로 나가면 브로커 400이고, `submit`은 B13 기본 거부 처리로 proposal을 해제한다(`analysis/evidence/engine--exitobserver.submit.ast.json` B13~B15; 현 소스 `exitloop.go:1582-1594`). tasks 1.1a의 “브로커 400 위임” 선택지는 이 보호 주문의 무제출을 의도적으로 허용하므로 불충분하다.

**착수 전 처분:** MARKET 자격에 KR 양의 정확 정수를 명시적으로 더하고, 불일치면 오늘의 LIMIT 경로로 기울인다. `checkOrderShape`와 `placeIntentSupported`에는 비분수 MARKET의 정수성 backstop을 두며, decimal KR 입력이 broker/wire에 도달하지 않음을 고정한다. 동시에 KR Position·sellable 수량의 정수 불변을 입구에서 증명하거나 위반 경보를 설계한다.

### P2-1 — 즉시성 제약은 맞지만 구현 실수 지점이 명확하다

design D7의 “제출 경로 GET 0·판정 실패 = LIMIT” 처분은 타당하다. `submit`의 `sellIntent` 오류는 B5에서 alert·release로 끝나므로, 판정 helper가 error를 반환하거나 cache miss를 GET으로 복구하면 오늘 LIMIT보다 한 주기 약해진다(`analysis/evidence/engine--exitobserver.submit.ast.json` B5). 2.6.3의 B5 미도달·GET 호출계수 RED는 이 반례를 겨냥한다. 구현은 오류 없는 `marketEligible` 결과와 동기화된 read-only snapshot을 분리해야 하며, 갱신 goroutine/채널의 경합·초기 cache 없음도 LIMIT으로 시험한다. 차단은 P0/P1 해결 뒤에 이 RED를 먼저 쓰면 닫힌다.

### P2-2 — LIMIT 골든과 변이 축을 더 명시해야 한다

2.6.3의 `PlaceIntent` 전체 골든은 LIMIT 칸을 고정하는 올바른 방향이다. 단 “오늘 `sellIntent` 출력”을 바뀐 함수를 다시 호출해 비교하면 자기비교가 된다. pre-change literal fixture(가격·수량·통화·유형 전체) 또는 독립 baseline renderer로 고정해야 한다. 2.6.5에는 결측·US·시간외·종가단일가·KRX 필드·GET 등이 있으나, **신선도 6시간/미래 fetchedAt/오늘 날짜/주말 검사 제거** 변이가 없다. 2.6.3 행은 있으나 mutation ledger에도 각각 넣어야 fail-open 회귀가 잡힌다.

### P2-3 — a165 교차 계획은 공유 flatten builder를 추가해야 한다

기존 a165 계획의 AST 재생성·gate ⑤ archive bundle 복사는 유지한다. 다만 a165 R4가 CLI flatten 취소 배선과 `buildFlattenWiring`을 소유하고(a165 tasks.md:22, 30), a087의 P0 수정도 이 경계에 의존한다. a165 먼저 착지하면 a087은 `cmd/tossctl/flatten.go:200`의 호출 관계, `NewOrderPath`의 옵션 부착 0, 관련 AST/구조 시험을 다시 확인한 뒤에만 착수한다. bundle 복사는 gate 작업량 처분일 뿐 이 의미 충돌의 대체 증거가 아니다.

### P3-1 — MARKET 칸의 남은 거부·체결 반례

1억원 이상 주문은 `ConfirmHighValueOrder:false`로 400이 될 수 있고(`internal/official/orders_write.go:118-124`, OpenAPI `:11307`), 하한가 매수호가 공백·거래정지에서는 MARKET도 체결을 보장하지 않는다. D8이 이미 유형 무관/잔존 위험으로 적어 둔 범위라 새 blocker로 승격하지 않는다. 다만 P1-1의 VI·정지 무판정과 섞어 “정상 접속매매 MARKET”으로 표현해서는 안 된다.

### 구현 착수 전 고칠 것

1. P0-1: 옵션을 `NewOrderPath` 밖의 엔진 전용 조립으로 이동하고 flatten을 포함한 구조 시험을 명시.
2. P1-1: KRX 고유 개장·정지/VI를 판별할 권위 입력과 snapshot 전달 계약을 정하거나, 불명은 LIMIT으로 축소.
3. P1-2: KR MARKET 정수 수량 불변·fallback/backstop·wire 미도달 시험으로 1.1a의 broker 400 위임 선택지를 제거.

Function Logic Map: not-applicable — 이 독립 리뷰와 Manager 처분 기록은 문서만 바꿨다. 현재 함수 내부 분기 주장은 `analysis/evidence/` AST 20개(열거 stale 0) 또는 위에 표시한 손 읽음 근거에서만 인용했다.

## 0.10 독립 리뷰 처분 (2026-10-28, OpenSpec 작성자 — HEAD `3e9989f61`)

판정: **보완 완료(P0-1 · P1-2 · P2 · P3) · P1-1 은 사람 결정 대기.** 재리뷰는 사람 결정 뒤에 받는다. 문서와 `tools/logic-map` AST 만 다뤘다 — 소스 · 시험 편집 0,
시험 · gate 실행 0, 브로커 · 계좌 · 운영 접근 0. 새 AST 5개(`analysis/evidence/engine--newcontext` · `tossctl--buildflattenwiring` · `execgw--canonicalprotectionquantity` ·
`official--buildordercreate` · `engine--floatof`)를 HEAD 에서 만들었고 `enumerate.py` 로 25개 전부 stale 0 을 확인했다(`enumerate-output.md` 재생성).
기억: ai-memory `a087 OR NewOrderPath` 검색은 세션 기록만 나왔고 관련 결정 페이지는 0 이다. Hindsight 는 이 세션에 없다 — 미실행.

| 발견 | 처분 | 위치 |
| --- | --- | --- |
| P0-1 `NewOrderPath` 는 엔진 전용이 아니다 | **수용 · 닫음.** 비시험 호출 자리 2(`engine.go:438` `NewContext` · `cmd/tossctl/flatten.go:200` `buildFlattenWiring`), 시험 호출 0(문자열 확인 2). 옵션은 `NewOrderPath` 에 두지 않고 `NewContext` B1(`:439`) 뒤에서 `path.Trading` 에만 덧입힌다. 1.5.5 에 (c) 함수 단위 구조 시험 · (d) `NewOrderPath`·flatten 인스턴스 `LiveReady=false` · (e) 엔진 `LiveReady=true`, 변이 「flatten 경로에 옵션 켜기」 2종 + 「엔진에서 옵션 제거」 | design D3 「0.10 독립 리뷰 반영」 · tasks 1.5.0·1.5.3·1.5.5 · spec Req 2 |
| P1-1 KRX 접속매매를 증명할 입력 | **조사 완료 · 사람 결정.** KRX 휴장·조기 폐장 = 달력 KRX 필드(openapi `:9808`, Go 미디코드 `typed_calendar_reads.go:11-14`). 종목별 KRX 정지(`:8815`, Go 미디코드 `stock_reads.go:30-40`)·VI(`/warnings` `:1268`) = 종목별 GET 이라 GET 0 아래 「지금」 불가. **KRX 단독 개장 시각·시장 전체 정지 = 입력 없음.** 그대로 두면 KR 도 LIMIT → Phase 2 의 실제 변경 0. 선택지 (가)(나)(다), 권장 (다) | design D7 「KRX 단독 판정 입력 조사」 · issues D-1 · tasks 0.10 |
| P1-2 KR 정수 수량 | **수용 · 닫음.** 판정 조건 6(같은 float · execgw 정수 술어 하나), 불일치 → 오늘 LIMIT. ②·③ backstop, wire 미도달. backstop 은 도달 불가여야 함(도달하면 B13 해제 루프) — 수량 표 · 종단 「LIMIT 제출이지 해제가 아님」 시험. 1.1a 브로커 400 위임 철회 | design D3 · D7 조건 6 · tasks 1.1a·2.6.3·2.6.5 · spec Req 1·2 |
| P2-1 B5 와 캐시 미준비 | **수용.** 판정은 `(bool, reason)` 총함수, 불변 스냅숏 원자 교체, 캐시 미준비 = LIMIT 사다리, B5 는 오늘의 사유로만 | design D7 「판정 함수의 모양」 · tasks 2.6.3 |
| P2-2 LIMIT 골든 · 변이 축 | **수용.** 골든은 편집 전 HEAD 의 `sellIntent` 출력을 RED 커밋에 literal fixture 로 먼저 커밋하고 바뀐 함수끼리 비교하지 않는다. 변이에 신선도 6시간 · 미래 `fetchedAt` · 오늘 날짜 · 주말 검사 제거 추가 | tasks 2.6.3·2.6.5 |
| P2-3 a165 와 flatten builder | **수용.** a165 착지 뒤 다시 할 일에 6 을 더한다(아래) | 이 절 |
| P3-1 1억 이상 · 하한가 공백 | 유지 — D8 잔존. 「정상 접속매매 MARKET」 표현은 P1-1 결정 전까지 보류 머리말로 덮었다 | design D7 |

### 착수 시 FLM 대상 정정 (위 「착수 시 FLM 대상」 목록)

- `engine.NewOrderPath`(엔진 인스턴스 옵션) → **`engine.NewContext`** 로 바꾼다. `NewOrderPath` 본문은 바이트 그대로다.
- 더한다: `execgw.canonicalProtectionQuantity`(내보내기로 이름·호출자가 바뀌면), `official` 달력 형을 디코드하는 함수 본문이 바뀌면 그것.
  `cmd/tossctl/flatten.go` 는 **편집 대상이 아니다**(구조 시험이 읽기만 한다).

### a165 착지 뒤 a087 이 다시 할 일 (추가)

6. **flatten 경계 재대조** — 새 main 에서 `rg -n 'NewOrderPath\(' --type go` 로 호출 자리를 다시 센다(지금 2). a165 R4 가 `buildFlattenWiring` 을 바꾸면
   `tools/logic-map` 으로 `cmd/tossctl/flatten.go` `buildFlattenWiring` · `engine.go` `NewOrderPath` · `NewContext` AST 를 그 HEAD 에서 다시 만들고, 1.5.5 (c)(d) 가
   새 배선에서도 「옵션 0」을 재는지(새 자리에서 `NewOrderPath` 를 부르거나 다른 빌더로 인스턴스를 받으면 그것도 센다) 확인한 뒤에만 착수한다.
   bundle 복사는 gate 작업량 처분이지 이 의미 충돌의 증거가 아니다.

Function Logic Map: not-applicable — 이 처분은 문서만 바꿨다. 분기 주장은 `analysis/evidence/` AST 25개(HEAD `3e9989f61`, stale 0)에서만 인용했고, 호출 자리 수는 `rg` 와 AST 로 셌다.

## 8판 개정 (2026-10-28, 사람 결정 D-1 (다) 「3은 3」) — **독립 재리뷰 대기**

OpenSpec 작성자 기록이다. 리뷰가 아니다. 문서와 `tools/logic-map` AST 만 다뤘다 — 소스 · 시험 편집 0, 시험 · gate 실행 0, 브로커 · 계좌 · 운영 접근 0.
새 AST 11개를 HEAD `1f4145c77`(origin/main `4b67dd6a9` 위)에서 만들었고 `enumerate.py` 로 36개 전부 stale 0 을 확인했다.

| 0.10 독립 리뷰 발견 | 8판에서 | 재리뷰가 판정할 것 |
| --- | --- | --- |
| P0-1 `NewOrderPath` 는 엔진 전용이 아니다 | **소멸** — `trading` 옵션 없음. 규칙(엔진 전용 상태는 `NewContext` 에서만)은 하한가 판 생성 자리에 승계, tasks 1.2 구조 시험 | 판 생성 · 갱신기 등록이 엔진 밖으로 새지 않는가 |
| P1-1 KRX 접속매매를 증명할 입력이 없다 | **소멸** — 유형이 LIMIT 이라 KRX 상태로 갈래를 나누지 않는다. 고정 창 `InRegularSession` 만 쓰고, 하한가 거부는 되먹임이 1 주기로 묶는다 | design D11 (c) 구간 표와 (e) 불변식 4 「1 주기」 칸이 맞는가 · 15:30 창 끝 |
| P1-2 KR 정수 수량 | **소멸** — 수량 경로가 오늘의 LIMIT 과 같다 | design D11 (f) |
| P2-1 B5 · 캐시 미준비 | **승계** — `exitLimitPrice` 총함수 · 불변 스냅숏 · 판 미준비 = 오늘 LIMIT | D11 (b) 「가격 결정」 |
| P2-2 LIMIT 골든 · 변이 축 | **승계** — 폴백 행 literal 골든 선커밋, 변이 13종(tasks 2.7) | tasks 2.2 · 2.7 |
| P2-3 a165 · flatten builder | **소멸** — 8판은 flatten 경로에 아무것도 붙이지 않는다. a165 재대조는 D11 (g) | D11 (g) |
| P3-1 1억 이상 · 하한가 공백 | 유지 — D8 잔존(8판 읽기 머리말) | — |

새로 판정받을 것: ① 거부 되먹임(B13 만, 출처가 하한가일 때만)이 「하한가 값이 틀린 모든 경우」를 덮는가 ② 하한가 ≤ 관측가 검사 · 같은 날 검사 · 정규장 중에만 읽기가 신선도 근거로 충분한가
③ `exitOrderLineFor` 사본 사다리 제거(결정 값 한 번) ④ 갱신기의 `MARKET_DATA` 공유 ⑤ 사람 확인 4건(`issues.md` D-2). 착수 시 FLM 대상은 design D11 「착수 시 FLM 대상」이 대체한다.

Function Logic Map: not-applicable — 문서 개정만 했다. 분기 주장은 `analysis/evidence/` AST 36개에서만 인용했다.

## 8판 독립 재리뷰 (GPT-5.6 Terra, 2026-10-29)

**결론: FAIL (blocker 3).** D-1(다)의 KR 정규장 하한가 LIMIT 자체는 명확하다. 그러나 8판은 “같은 날”을 수신 시각으로만 라벨링하고, 공유 `MARKET_DATA` 버스트와 비동기 비수용 해제 경로를 누락했다. 따라서 “오늘보다 나빠지는 칸은 하한가 직접 거부 뒤 1주기뿐”은 성립하지 않는다.

| 심각도 | 발견 | 위치 | 근거 | 처분 제안 |
| --- | --- | --- | --- | --- |
| P0 | 응답의 `timestamp`는 디코드 뒤 버려지고 board에는 읽은 KST 날짜만 남는다. 전일 하한가 응답을 오늘 10:00에 받아도 같은 날 항목이 된다. `하한가 ≤ 관측가`는 전일 값이 더 높은 경우만 거르고, 더 낮은 전일 값은 통과한다. 고정 시각 창은 주말만 거르므로 평일 휴장·개장 지연도 이 반례를 막지 않는다. spec의 “같은 KST 날짜에 얻은”을 실제 당일 가격으로 증명할 수 없다. | `design.md:70-97,107-113`; `specs/order-execution/spec.md:17`; `internal/official/price_limits_reads.go:20-62`; `internal/clock/market.go:128-152` | AST: `Client.PriceLimits`, `adaptPriceLimits`, `parseDecimal`; 손으로 읽음: raw `timestamp` 미전파, `parseDecimal` null/오류→0 | 응답 timestamp를 보존·검증하는 엔진 전용 읽기 형을 둔다. CLI `domain.PriceLimits` 출력은 그대로 유지할 수 있다. “오늘 수신한 전일 payload”, null, parse 오류, 기준가 변경일을 분리한 시험과 변이를 추가하기 전 구현 금지. |
| P0 | 갱신기 5개 순차 GET은 관측 `/prices`와 같은 `Context.Official`·`MARKET_DATA` 10/s를 쓴다. 현 클라이언트는 응답 헤더를 기록할 뿐 요청을 예약·제한하지 않는다. 버스트/재시도/다른 소비자와 겹친 429가 `/prices`에 오면 `ObserveOnce`는 판단·제출 전체를 건너뛰고 outage 경로로 간다. 이는 하한가 1회 거부가 아니라 보호 판단 자체의 지연 칸이다. | `design.md:99-101,179-181`; `internal/app/engine/exitwiring.go:333`; `internal/official/client.go:191-207`; `internal/app/engine/exitloop.go:493-500,860-900` | AST: `Client.doRequest`; 손으로 읽음: `Context.ExitObserver`가 `c.Official` 주입, price-read 오류의 주기 조기 반환 | 같은 rate group에서 exit 가격 읽기에 우선 예약을 주고 auxiliary는 남은 예산에서만 실행하거나, 동등한 강제 배분 근거를 설계·시험으로 제시한다. “updater 429 중단”만이 아니라 `/prices` 429에서도 손절 판단이 굶지 않음을 검증한다. |
| P0 | 되먹임은 `submit` B13 직접 비수용에만 있다. B8/미상 결과는 안전상 무장 유지하지만, 뒤에 원장이 `ExitIntentUnaccepted`로 확정하면 `noteHeldProposal`이 발의를 푼다. board의 lower-limit 거부 표시는 없으므로 다음 관측은 같은 하한가를 다시 낸다. 즉 `IN_DOUBT → 확정 비수용 → 해제 → 동일 하한가`가 반복될 수 있다. | `design.md:117-124`; `internal/app/engine/exitloop.go:1547-1592`; `internal/app/engine/exit_held_proposal.go:126-145` | AST: `ExitObserver.submit`(기존 evidence), `ExitObserver.noteHeldProposal` B4; 손으로 읽음: 해제 후 source/rejection feedback 부재 | 가격 출처를 attempt/해제 상관관계까지 보존하고, 직접 B13과 비동기 확정 비수용 모두 같은 날 board를 거부 처리한다. `IN_DOUBT → StateFailedConfirmed → release → 다음 주기 observed fallback` 종단 시험을 추가한다. |
| P1 | board는 “decimal 문자열”을 저장한다고 하나 현재 표면은 `float64`뿐이다. null·파싱 오류 모두 0이어서 결측 사유도 합쳐진다. 새 문자열의 정본 변환·tick/정밀도 계약과 B5 도달 불가 증명이 없다. | `design.md:79-82,107-116`; `internal/official/market_reads.go:15-24`; `internal/official/price_limits_reads.go:55-62`; `internal/app/engine/exitloop.go:1806-1832` | AST: `adaptPriceLimits`, `parseDecimal`, `ExitObserver.sellIntent` | raw decimal과 timestamp를 함께 보존하거나 canonical formatting·허용 tick을 명시하고 시험한다. 결측·형식 오류는 별도 telemetry를 남긴다. |
| P1 | `[미측정]` 표시는 대체로 남아 있으나 “체결될 것이다”, “접수돼 참여할 것이다”처럼 영수증·규정 원문 없이 단정한 문장이 남는다. tasks 5.4는 착수/착지 전 조건을 아직 사람 결정으로만 남겨 착지 gate가 아니다. 거부면 되먹임이 기본 경로가 된다는 사실도 gate로 고정되지 않았다. | `design.md:153-156`; `tasks.md:393-400`; `issues.md` D-2 ② | 손으로 읽음 | 문장을 조건형으로 낮추고, 5.4의 수용·거부·판정 불가 각각에 대한 **착지 전** 사람 처분을 D-2에서 확정해 task check/gate로 만든다. |
| P1 | D11은 a165를 “코드 0, 미착지”로 적지만, 의뢰 시점 정보는 a165가 `clearTheSymbol`·`checkSymbolFree` 변경과 함께 착지 중이다. 같은 `exitloop.go`의 자료·좌표·시험 double 재대조 범위가 문서보다 넓다. 이 worktree에서 main/a165를 열지 않았으므로 충돌 없음은 확인하지 못했다. | `design.md:193-205` | 손으로 읽음 + 의뢰 정보 | a165 착지 SHA에서 `clearTheSymbol`과 `checkSymbolFree`를 포함해 FLM/BTM·AST·공유 double을 다시 대조하고, 그 증거 전 a087 구현 착수 금지. |
| P2 | `exitOrderLineFor` 사본 제거와 `price_source` 추가는 방향이 맞다. 현재 a129 golden helper는 예상 키만 부분 비교하므로 추가 키가 기존 golden을 즉시 깨지는 않는다. 반대로 새 키의 존재·값·모든 제출 결과의 소비 계약을 보장하지도 않는다. | `internal/app/engine/exit_order_log.go:102-114`; `internal/app/engine/a129_exit_order_log_test.go:204-211`; `design.md:128-138`; `tasks.md` 3.1 | AST: `exitOrderLineFor`; 손으로 읽음: a129 helper 부분 비교, `cmd`/`internal` 검색에서 exit-order `price_source` 소비자 미발견 | 복사 제거 후 선택값·source를 한 번 전달하고, a129 golden에 `lower_limit|observed|baseline`의 필수 키/값과 B1·B5·B13·비동기 결과 행을 명시한다. 소비자 0은 구현 시 재검색으로 확인한다. |

### 확인한 것

- D-1(다)와 D11(a): KR 보호 매도만 `[09:00,15:30)`에서 하한가 LIMIT, 나머지는 현재 LIMIT 사다리라는 범위는 일관됐다.
- AST/손으로 읽음: 현재 `sellIntent` B1~B4는 관측가→기준선→거부 사다리와 `OrderType: "limit"`을 유지한다. `NewContext`의 비시험 조립은 `cmd/tossctl/engine_assembly.go:40` 한 곳이고 flatten은 별도 `NewOrderPath`다. v8의 board가 아직 없으므로 P0-1 배선은 설계·task만 확인했다.
- `analysis/evidence/enumerate.py`: `stale=0`. `openspec validate a087-a-protective-exit-is-a-market-order --strict --no-interactive`: 통과.
- `exitOrderLineFor`의 기존 사본과 a129 golden의 부분 비교를 대조했다. 현재 알려진 golden 파손은 없으나 새 contract 검증은 없다.

### 안 본 것·제한

- v8 구현은 없으므로 코드·단위/종단/변이 시험, 전체 gate, 렌더 UI 검증은 실행하지 않았다. 브로커·계좌·운영·mutating 명령도 실행하지 않았다.
- 거래소 체결 규정 원문과 실주문 영수증은 새로 조사하지 않았다. `impeccable: N/A — backend/documentation-only`.
- a165 main/착지 worktree는 금지 범위라 열지 않았다. 위 a165 판단은 의뢰 정보와 이 change의 오래된 D11 기록의 불일치다.
- `grill-me`: D-1은 확정값으로 사용했고 새 사람 결정은 만들지 않았다. `search-first`: 구현·의존성 채택이 아닌 독립 문서/소스 적대 검토라 현재 저장소·AST·OpenSpec 근거만 사용했다.

## 9판 개정 (2026-10-29, Manager 결정 — 8판 독립 재리뷰 FAIL 처분) — **독립 재리뷰 대기**

OpenSpec 작성자 기록이다. 리뷰가 아니다. 문서와 `tools/logic-map` AST 만 다뤘다 — 소스 · 시험 편집 0, 시험 · gate 실행 0, 브로커 · 계좌 · 운영 접근 0.
새 AST 6개(HEAD `5e691ac4e`, 소스 바이트 = origin/main `4b67dd6a9`), `enumerate.py` 42개 stale 0.

| 8판 재리뷰 발견 | 9판 처분 | 위치 |
| --- | --- | --- |
| P0 전일 하한가가 같은 날 항목이 됨 | 엔진 전용 읽기 형(timestamp · raw decimal 보존), 조건 = timestamp KST 날짜 = 그날 거래일 · 09:00 이후, 거름 사유 10종 따로. 「데이터 시각」 뜻 `[미측정]` → 5.4 ②. 전일 값 논증(⑥ 이 위쪽 범위를 막고, 아래쪽은 되먹임) — 제도 근거 `[미인용]` | design D12.1 · spec Req 1 · tasks 2.8 · 5.4 |
| P0 `MARKET_DATA` 버스트 | 가) 관측 주기 끝 토큰(`Run` 한 줄) · 주기당 ≤ 1 · GET 시한 ≤ 2초 < 5초 · 429 세 입력 → 5분 정지. 나)는 다음 주기를 GET 시간만큼 늦춰 기각. 새 압력 0.2/s(2%) | design D12.2 · tasks 2.4 |
| P0 IN_DOUBT → 비수용 → 같은 하한가 반복 | `CONFIRMED` 아닌 모든 결과에서 제출 순간 그날 금지(B8 포함) · 술어 하나 · 재시작 재구성. 꼬리표 수단은 detail 덮어쓰기(`durability.go:627`) 때문에 T3 유도를 기본으로 두고 Manager 결정(D-2 ⑤) | design D12.3 · tasks 2.5 · 2.9 |
| P1 float · 결측 사유 합침 | raw 정수 decimal · 사유별 기록 · B5 도달 불가 시험 | design D12.1 |
| P1 단정 문장 · 5.4 gate 아님 | 조건형으로 · 5.4 착지 전 필수 gate · 처분 셋 | design D11 (c) · D12.4 · tasks 5.4 · issues D-2 ② |
| P1 a165 「코드 0」 | 8판 시점 사실로 낮춤 · a165 archive/push sha 재대조 = tasks 1.0 착수 조건 | design D11 (g) · D12.5 · tasks 1.0 |
| P2 `price_source` 골든 | 필수 키 · 값 셋 · B1/B5/B7/B13/비동기 행 | design D12.6 · tasks 3.1 |

불변식 4 표는 design D12.7 이 대체한다(나빠질 수 있는 칸 셋과 상한).

Function Logic Map: not-applicable — 문서 개정만 했다. 분기 주장은 `analysis/evidence/` AST 42개에서만 인용했다(`a129_exit_order_log_test.go` helper · `doRequest` 의 기록 줄 · `Retrier.Query` 오류 사슬은 손 읽기로 표시).

## 9판 재리뷰 (GPT-5.6 Terra, 2026-10-29)

**결론: FAIL (blocker 1).** D12.1의 raw/timestamp 보존, D12.2의 주기 끝 토큰, D12.6의 `price_source` 골든은 전진했다. 그러나 T3의 “최신 PLACE 상태”는 같은 날의 **단조 금지**를 재시작 뒤 복원하지 못한다. 8판 P0-3은 닫히지 않았다.

| 심각도 | 발견 | 위치 | 근거 | 처분 제안 |
| --- | --- | --- | --- | --- |
| P0 | T3는 가격이 같은 intent의 **최신** PLACE 상태만 본다. runtime 규칙은 하한가 시도가 한 번이라도 `CONFIRMED`가 아니면 그날 금지를 풀지 않는데, `IN_DOUBT → CONFIRMED` 전이는 허용된다. 재시작 뒤 최신 상태만 읽으면 금지가 허용으로 바뀐다. 앞선 하한가 비수용 뒤 관측가가 우연히 같은 가격으로 제출·확정돼도 더 늦은 관측가 intent가 거부 증거를 덮는다. account/market scope도 T3 설명에 명시돼 있지 않다. | `design.md:287-312`; `specs/order-execution/spec.md:17-33`; `internal/journal/lifecycle.go:43-48`; `internal/journal/schema.go:202-224` | AST: `Attempt.transition` B16 및 상태 갱신; 손으로 읽음: `IN_DOUBT → CONFIRMED` 허용, intents가 immutable decimal string·account/market/day 보유 | T3를 최신 상태 판정으로 쓰지 않는다. `account_ref`·`market`·`trading_day`·symbol·보호 청산 scope를 고정하고, 정규화한 같은 가격 후보의 **전이 이력 전체**에서 한 번이라도 비수용/IN_DOUBT가 있으면 하루 금지인 단조 술어로 바꾼다. source를 T3로 구별할 수 없는 우연 일치는 금지로 남긴다. `IN_DOUBT → CONFIRMED → restart`, `하한가 FAILED → 같은 가격 observed CONFIRMED → restart`, 타 계좌/시장 격리 시험을 추가한다. |
| P1 | D12.2의 “어디서든 429 → 5분 정지”는 헤더 없는 429가 `Retrier.Query` 안에서 재시도 후 성공하면 성립하지 않는다. `cycle.Err`는 nil, rate budget은 `Reported=false`라 `Exhausted()`도 false다. ②·③ 모두 멈춤 입력을 놓친다. 헤더 없는 429의 **실제 빈도**는 `[미측정]`이 맞지만, 그 응답을 놓치는 **구조**는 소스상 확정 사실이다. | `design.md:253-286,334-344`; `specs/order-execution/spec.md:31-33`; `internal/execgw/retry.go:347-377`; `internal/official/ratebudget.go:175-177,227-243`; `internal/official/errors.go:54-55` | AST: `Retrier.Query` B10~B12; 손으로 읽음: 재시도 성공은 nil 반환, headerless budget은 미보고 | HTTP 429를 헤더와 무관하게 기록하는 공유 rate-limited latch/epoch를 D12.2 정지 입력으로 쓴다. `headerless 429 → retry success → 5분 GET 0` 시험을 추가한다. 그 전에는 429 정지 SHALL과 D12.7 #4 상한을 주장할 수 없다. |
| P1 | D12.7 #1·#2 상한은 `release` 성공을 가정한다. B13과 비동기 비수용 해제는 `ReleaseUnacceptedExitProposal` 오류를 반환하고, 무장 상태가 남으면 복구 시간은 5초/해소 시간 상한에 없다. 새 하한가 거부가 이 경로를 새로 밟게 하므로 표 상한은 완전하지 않다. | `design.md:334-344`; `internal/app/engine/exitloop.go:1567-1596,1601-1607`; `internal/app/engine/exit_held_proposal.go:126-145` | AST: `ExitObserver.submit` B8~B13, `noteHeldProposal` B4; 손으로 읽음: release 오류 전파 | 표에 journal release 실패·복구 칸과 비유한 상한을 넣거나 기존 관측 불능/지연 contract에 합친다. 직접 거부·IN_DOUBT 해제에서 release 실패 종단 시험을 추가한다. |
| P1 | D12.1은 raw decimal·timestamp 유실을 고쳤지만 timestamp가 효력 날짜인지 응답 시각인지는 아직 모른다. 5.4 착지 전 gate는 맞지만 KST 날짜를 KRX 거래일로 등치한 문구는 평일 휴장·기준가 변경을 증명하지 않는다. P0-1은 **착지 전 gate로 보류**됐을 뿐 현재 사실로 닫힌 것이 아니다. | `design.md:228-251,314-320`; `tasks.md:216-223,422-426`; `internal/clock/market.go:128-152` | 손으로 읽음 | 5.4에서 timestamp가 응답 시각이면 KR off 또는 별도 권위 있는 효력일 검증 중 하나를 사람이 확정해 gate를 닫는다. 구현/배포 완료 주장 금지. |
| P2 | T3 오탐은 시도 당시 관측가와 하한가가 같으면 주문 가격도 같아 즉시 가격 결과는 같다. 다만 재시작 뒤 관측가가 높아지면 금지는 오늘 LIMIT(더 높은 가격)으로 남아 하한가의 공격성을 잃는다. 보수적이고 불변식 4에는 무해하지만 “결과가 같다”는 문장은 그 시점에 한정해야 한다. 원장 읽기 오류 → 항목 없음은 설계대로 fail-closed다. | `design.md:305-312`; `tasks.md:220-229`; `internal/journal/durability.go:113-145` | 손으로 읽음 | T3 오탐을 의도된 보수적 손실로 명시하고, 양의 정수 canonical 비교와 파싱 불가 후보 → 금지/항목 없음의 규칙·시험을 고정한다. |

### 확인한 것

- D12.2 가) 토큰을 `Run`의 `ObserveOnce` 반환 뒤·`Sleep` 앞에 넣으면, AST상 그 주기의 `/prices`·재시도·판정·제출·defer가 끝난 뒤다. GET 시한이 2초 이하로 지켜지면 자기 `/prices`와 시간상 겹침 0이라는 설계는 성립한다. 구현 전이라 실행 증거는 아니다.
- D12.1의 엔진 전용 raw 읽기는 기존 CLI 형을 건드리지 않고 null/형식/timestamp 사유를 분리한다. D12.4의 5.4는 착지 전 필수 gate, D12.6은 a129 `price_source` 필수 키 시험이다.
- `analysis/evidence/enumerate.py`: `stale=0`. `openspec validate a087-a-protective-exit-is-a-market-order --strict --no-interactive`: 통과.

### 안 본 것·제한

- 9판 구현·단위/종단/변이 시험·전체 gate·실주문은 실행하지 않았다. 브로커·계좌·운영·mutating 명령 0.
- a165 main/착지 worktree는 금지 범위라 열지 않았다. D12.5 재대조는 a165 archive/push SHA 뒤 작업이다.
- 거래소 timestamp 의미와 헤더 없는 429의 실제 빈도는 외부 영수증/규정으로 검증하지 않았다. `impeccable: N/A — backend/documentation-only`.
- `grill-me`: D-1·D-2 ⑤는 확정값으로 사용했고 새 사람 결정을 요구하지 않았다. `search-first`: 구현·의존성 채택이 아닌 독립 문서/소스 적대 검토라 현재 저장소·AST·OpenSpec 근거만 사용했다.

## 10판 개정 (2026-10-29, Manager 결정 — 9판 재리뷰 FAIL 처분, KISS) — **독립 재리뷰 대기**

OpenSpec 작성자 기록이다. 리뷰가 아니다. 문서와 `tools/logic-map` AST 만 다뤘다 — 소스 · 시험 편집 0, 시험 · gate 실행 0, 브로커 · 계좌 · 운영 접근 0.
새 AST 1개(`engine--exitobserver.release`, HEAD `525b4fb6c`), `enumerate.py` 43개 stale 0.

| 9판 재리뷰 발견 | 10판 처분 | 위치 |
| --- | --- | --- |
| P0 T3 최신 상태 판정은 재시작 뒤 단조 금지를 복원 못 함 | 가격 · 상태 · 출처 판정을 없앴다 — 하한가는 그날 그 종목의 첫 보호 매도 시도에만. 술어 = 같은 `account_ref` · `market` · `trading_day` · symbol 의 보호 청산 매도 intent 존재(`intents` 는 갱신 · 삭제 없음 → 단조). 원장 오류 → 금지. 첫 시도 뒤는 오늘 LIMIT(보수적 손실). T2/T3 · D-2 ⑤ 대체 | design D13.1 · spec Req 1 · tasks 2.5 · 2.9 |
| P1 헤더 없는 429 를 놓침 | `doRequest` 가 HTTP 429 를 헤더 무관하게 마지막 시각으로 기록(원자 값 · 접근자 하나), 갱신기는 5분 GET 0. 9판 입력 ②③ 대체. `doRequest` 를 Pre-Edit · FLM 에 추가 — a166 · a100 과 같은 함수(대조 기록) | design D13.2 · tasks 2.4 · 2.10 |
| P1 release 실패가 상한에 없음 | 불변식 4 표 #3 — 오늘의 비수용 해제 실패 경로와 같은 상한, 발생 종목당 하루 1회 이하. 직접 거부 · IN_DOUBT 해제 각각의 release 실패 종단 시험 | design D13.5 · tasks 2.5 |
| P1 timestamp 뜻 · KST 날짜 등치 | P0-1 은 5.4 gate 로 **보류**라고 정정. 기본 처분 = 응답 시각이면 KR off(사람 확인). 「KST 날짜 일치」는 평일 휴장 · 기준가 변경을 증명하지 않음 | design D13.3 · issues D-2 ② |
| P2 T3 오탐 · 정본 비교 | 첫 시도 규칙으로 오탐 문장 소멸. 남은 ⑥ 비교는 양의 정수 정본 · 값 비교, 파싱 불가 → 쓰지 않음(시험 행) | design D13.4 · tasks 2.2 |

Function Logic Map: not-applicable — 문서 개정만 했다. 분기 주장은 `analysis/evidence/` AST 43개에서만 인용했다(`intents` 불변 · `exit_events` 결속 쿼리는 schema 주석 손 읽기로 표시).

## 10판 재리뷰 (GPT-5.6 Terra, 2026-10-29)

**결론: FAIL (blocker 1).** D13.1의 첫 시도 단조 규칙은 T3 최신 상태 문제를 제거한다. 그러나 재시작 재구성 술어가 실제 기준선 이탈 action 값 `BASELINE_BREACH_EXIT`가 아니라 존재하지 않는 `BASELINE_BREACH`를 찾는다. 첫 기준선 이탈 하한가 시도 뒤 재시작하면 금지가 풀려 같은 하한가를 다시 낼 수 있으므로, D13.5의 종목당 하루 1회 상한은 성립하지 않는다.

| 심각도 | 발견 | 위치 | 근거 | 처분 제안 |
| --- | --- | --- | --- | --- |
| P0 | 재구성 술어의 `BASELINE_BREACH`는 이벤트에 쓰이지 않는 값이다. 실제 `ActionBaselineBreach` 값은 `BASELINE_BREACH_EXIT`이고, 판단은 `string(proposal.Action)` 그대로 `exit_events.action`에 기록된다. 따라서 첫 기준선 이탈 하한가 시도는 runtime 메모리에서는 금지되지만, 재시작 뒤 D13.1 질의에는 보이지 않는다. 새 board가 다시 공개되어 같은 가격의 직접 거부·비수용이 재시작마다 반복될 수 있다. `STOP_LOSS_LADDER`만 이 술어에 남는다. | `design.md:361-371`; `tasks.md:231-234`; `internal/exitpolicy/ratchet.go:94-104`; `internal/app/engine/exitloop.go:1401-1415,1459-1465` | AST: `ExitObserver.submit` B1이 `isProtective`를 호출하고 B6 뒤 `Place`를 호출함; 손으로 읽음: `isProtective`의 두 enum과 `ExitProposal.Action` 기록값 | D13.1·tasks 2.9를 실제 enum 값 `BASELINE_BREACH_EXIT`로 고치거나 query가 `string(exitpolicy.ActionBaselineBreach)`를 쓰게 명시한다. 기준선 이탈의 CONFIRMED·FAILED_CONFIRMED·IN_DOUBT→비수용 각각에서 restart 뒤 관측가 LIMIT인 종단 시험을 추가한다. |
| P1 | D13.5 #4의 “주기당 ≤1 = 0.2/s”와 “429 뒤 5분 GET 0”은 wire 요청 상한·수명 범위가 빠졌다. 기존 `Client.send`은 401이면 `doRequest`를 최대 두 번 더 호출하므로, 새 raw `PriceLimits`가 표준 official GET 경로를 재사용하면 토큰 하나가 최대 3회 `MARKET_DATA` GET가 된다. 또 D13.2의 원자 `last HTTP 429`는 새 `Client`에서 0이므로 429 직후 엔진 재시작은 남은 5분 안에도 다음 토큰의 GET를 허용한다. 이는 headerless 429 재시도 성공의 **같은 Client** 경로는 고치지만 표의 무조건 상한은 고치지 못한다. | `design.md:274-287,380-393,415-418`; `tasks.md:214-220`; `internal/official/client.go:317-365` | AST: `Client.doRequest` B1 뒤 응답 반환 구조, `Retrier.Query` B10→B5 재시도 성공; 손으로 읽음: `send`의 401 재전송 최대 2회와 process-local 원자 값 | raw read의 실제 wire-call 상한을 명시·시험해 표를 최대 3회/0.6/s로 바로잡거나, 하한가 읽기에는 1회 전송만 허용한다. 429 epoch를 durable하게 보존하거나 기동 뒤 5분 fail-closed warm-up을 둬 restart 시나리오도 5분 GET 0으로 고정한다. 401→401→200, headerless 429→성공→restart 두 종단 시험을 추가한다. |

### 확인한 것

- 보호 범위는 현재 `isProtective`의 기준선 이탈·손절 사다리 두 action뿐이다. 익절(`LADDER_TAKE_PROFIT` 등)은 orderable이어도 이 범위 밖이므로 D13.1이 익절 intent만으로 금지하지 않는 것은 의도와 맞다. action 결속은 flatten의 별도 원장 사슬과 수동 일반 매도를 보호 action으로 오인하지 않는 방향이다(손으로 읽음).
- D13.1의 “`Prepare` 전 로컬 종료이면 재시작 뒤 한 번 더 허용”은 현재 `ReasonSymbolInFlight`에 대해 성립한다. AST상 `Gateway.submit`은 `claimSymbol`·`checkSymbolFree`를 `journal.Prepare`보다 먼저 실행하고, 실제 broker `plan.call`은 `Prepare` 뒤 dispatch 안에만 있다. 그 갈래에 intent가 없고 broker에도 닿지 않는다.
- D13.2는 같은 Client 수명 안에서 headerless 429→`Retrier.Query` 재시도 성공(`cycle.Err == nil`)을 덮는다. 제안한 store가 `c.hc.Do` 성공 직후, status 분류·재시도 전에 실행되므로 뒤 200이 그 시각을 지우지 않는다. 반환값·rate-budget 기록을 보존한다는 범위도 맞다.
- D13.3은 timestamp 의미가 확인될 때까지 5.4 착지 전 gate로 보류한다고 정확히 적었다. 이것은 현재 신선도 증명이 아니며, gate·사람 처분 전 구현/배포 PASS가 아니다. D13.4의 정수 문자열·값 비교와 D12.6의 `price_source` 골든 요구는 문서상 유지된다.
- Manager의 `doRequest` 소유 순서와 a165의 `send` 호출자 재대조 task는 서로 모순되지 않는다. a165의 `clearTheSymbol`·`checkSymbolFree`는 a087 착수 시점 재대조 대상으로 남는다.
- `analysis/evidence/enumerate.py`: `stale=0`. `openspec validate a087-a-protective-exit-is-a-market-order --strict --no-interactive`: 통과.

### 안 본 것·제한

- 10판은 설계 단계다. 구현·단위/종단/변이 시험·전체 gate·실주문은 실행하지 않았다. 브로커·계좌·운영·mutating 명령은 0이다.
- 거래소 timestamp 의미, 하한가 지정가 접수·체결, 다른 프로세스와의 실제 `MARKET_DATA` 경쟁은 검증하지 않았다. `[미측정]`은 tasks 5.4 사람 gate로 남는다.
- `impeccable: N/A — backend/documentation-only`. `grill-me`: D-1·D-2와 두 Manager 결정을 확정값으로 사용했고 새 사람 결정을 만들지 않았다. `search-first`: 구현·의존성 채택이 아닌 독립 문서/소스 적대 검토라 현재 저장소·OpenSpec·AST만 대조했다.

## 11판 개정 (2026-10-29, Manager 결정 — 10판 재리뷰 FAIL 처분) — **독립 재리뷰 대기**

OpenSpec 작성자 기록이다. 리뷰가 아니다. 문서만 다뤘다(새 AST 0 — `engine--isprotective` 등 기존 AST 인용, `send` 의 401 재전송은 손 읽기로 표시). 소스 · 시험 · 브로커 0.

| 발견 | 11판 처분 | 위치 |
| --- | --- | --- |
| P0 재구성 술어의 `BASELINE_BREACH` 오기 | 질의 action 은 `isProtective` 와 같은 Go 상수(`ActionBaselineBreach` = `BASELINE_BREACH_EXIT` · `ActionLadderStop`)에서 유도, 리터럴 금지. 구조 시험(상수 AST 열거 → `isProtective` 참 집합 = 질의 집합) · 행동 시험(기준선 이탈 · 손절 사다리 × 세 결과 → 재시작 → 금지). design · tasks · proposal 의 같은 오기 9곳 정정 | design D14.1 · tasks 2.9 · 2.7 |
| P1 wire 상한 · 재시작 수명 | (a) 401 시 표준 `send` 재전송으로 토큰당 wire GET ≤ 3 — 상한에 그대로 적음, 새 경로 없음 (b) 엔진 시작 뒤 5분 GET 0 — 429 기록 없이 「429 직후 재시작」 닫음. 시험 3종 | design D14.2 · tasks 2.4 |
| (Manager 발견) 예산 기준 | 그룹 10/s → 계좌 공유 3 req/s(사용자 결정 2026-10-22 결정 3): 주기당 ≤ 1 ≈ 6.7%, 401 시 ≤ 3 ≈ 20%. a166 착지 시 그 예산을 탐, a087 은 기다리지 않음 | design D14.3 |

불변식 4 표는 design D14.4 가 대체한다. Function Logic Map: not-applicable — 문서 개정만 했다.

## 11판 재리뷰 (GPT-5.6 Terra, 2026-10-29)

**결론: FAIL (blocker 1).** D14.1은 실제 두 action 상수를 질의에 쓰도록 고쳐 10판 P0의 현재 오기를 닫았고, 시작 정숙 5분도 제출 경로를 멈추지 않는다. 그러나 D14.3의 `0.6/s ≈ 20%`는 5초 평균일 뿐 계좌 공유 **3 req/s**의 한 초 창 상한이 아니다. a087이 a166의 공유 admission 전에 최대 3회 GET burst를 추가하므로 D14.4 #4의 즉시성 상한은 성립하지 않는다.

| 심각도 | 발견 | 위치 | 근거 | 처분 제안 |
| --- | --- | --- | --- | --- |
| P0 | `401 → 401 → 200` 한 토큰은 최대 3 wire GET이나, 세 요청 사이의 pacing/admission이 없다. 더구나 토큰은 `ObserveOnce` 반환 직후, 5초 `Sleep` 전에 보내므로 직전 관측 `/prices`와 **동시 실행 0**이어도 같은 1초 rate 창일 수 있다. 다른 계좌 공유 소비자도 같은 창에 들어온다. `last HTTP 429`와 시작 정숙은 응답을 **받은 뒤** 다음 GET만 멈출 뿐, 이미 초과한 창이나 그 창의 `/prices` 지연을 막지 못한다. 따라서 `≤3/5초 = 0.6/s ≈ 20%`는 계좌 3 req/s에 대한 안전 상한이 아니며, “a087은 a166을 기다리지 않음”은 공유 예산 결정과 충돌한다. D14.4 #4는 새 `/prices` 429·재시도 대기의 상한을 실제로 닫지 못했다. | `design.md:463-477`; `tasks.md:218-224`; `internal/app/engine/exitloop.go:407-417`; `internal/official/client.go:329-365`; `openspec/changes/verify-execution-capability/issues.md:578-579,593-600` | AST: `ExitObserver.Run`은 `ObserveOnce` 뒤 `Sleep` 전 별도 대기 없음; `Client.send` B4 loop 안 `doRequest`와 최초 호출로 최대 3회. 손으로 읽음: loop 조건·계좌 공유 결정 | a087의 하한가 GET을 a166의 실제 계좌 공유 admission 뒤로 gate하거나, 그 전에는 GET 0(오늘 LIMIT)으로 fail-closed 하라. 단순 평균/자체 비동시성으로 대체하지 말고, 같은 1초 창의 `/prices`·401 3회·다른 소비자 합계를 넣은 시험과 admission 거절 시 GET 0 시험을 고정하라. |
| P1 | D14.1의 **현재** 비교는 이름 수가 아니라 실제 `Action` 값을 `isProtective`에 넣어 집합을 비교하므로 방향은 맞다. 다만 “Action 상수 전부를 AST 열거”의 type 판별 규칙이 없다. 현 `tools/logic-map`의 `isProtective` AST는 반환 좌표만 내고 비교 피연산자를 내지 않는다. 식별자 prefix/존재 수로 구현하면 `const ActionFuture = Action("...")`처럼 명시 type 자리가 다른 선언이나 새 파일의 상수를 빼고도 통과할 수 있다. | `design.md:440-445`; `tasks.md:242-244,253`; `internal/exitpolicy/ratchet.go:92-104`; `internal/app/engine/exitloop.go:1463-1465` | AST: `engine--isprotective`는 branch 0·return 좌표 하나라 selector 집합 증거가 아님. 손으로 읽음: 현재 두 selector와 모든 Action 선언 | 구조 시험은 `go/types` 등으로 package 전체의 **평가된 `exitpolicy.Action` 상수값**을 열거하고, 실제 `isProtective` 호출 결과와 질의 집합을 값으로 비교한다고 명시하라. `ActionFuture = Action("...")` 추가 + `isProtective` 편집 변이를 넣어 이름/개수 검사만으로는 통과할 수 없게 하라. |

### 확인한 것

- 손으로 읽음: 현재 `isProtective`의 참 집합은 `ActionBaselineBreach`(`BASELINE_BREACH_EXIT`)와 `ActionLadderStop`뿐이다. `ExitProposal.Action`은 `string(proposal.Action)`으로 기록된다. D14.1의 현행 상수 유도와 기준선·손절 재시작 행동 시험은 10판의 문자열 오기를 닫는다.
- AST: `Gateway.submit`의 `claimSymbol`·`checkSymbolFree`는 `Prepare` 전, 실제 `plan.call`은 `Prepare` 뒤다. `Prepare` 전 로컬 종료를 재시작 뒤 재허용하는 갈래는 브로커에 닿지 않는다는 10판 판단을 유지한다.
- 설계상 시작 정숙은 보조 갱신기의 GET만 5분 0으로 한다. 새 판이 없거나 금지면 오늘 LIMIT이라는 가격 결정 표와 결합하므로, 보호 `submit` 자체를 5분 멈추는 경로는 보이지 않는다. 구현 전 문서 대조이며 실행 증거는 아니다.
- `analysis/evidence/enumerate.py`: `stale=0`. `openspec validate a087-a-protective-exit-is-a-market-order --strict --no-interactive`: 통과.

### 안 본 것·제한

- 11판은 설계 단계다. 구현·단위/종단/변이 시험·전체 gate·실주문은 실행하지 않았다. 브로커·계좌·운영·`mutating` 명령은 0이다.
- 브로커의 실제 rate-window 집계 방식과 다른 프로세스와의 429 빈도는 `[미측정]`이다. 이는 공유 admission 부재라는 정적 사실이나 그 부재를 a166 전까지 허용하는 결론을 해소하지 않는다.
- `impeccable: N/A — backend/documentation-only`. `grill-me`: 확정 D-1·D-2 및 Manager 결정을 재질문하지 않았다. `search-first`: 구현·의존성 채택이 아닌 독립 문서/소스 적대 검토라 현재 저장소·OpenSpec·AST만 대조했다.

## 12판 개정 (2026-10-29, Manager 결정 — 11판 재리뷰 FAIL 처분) — **독립 재리뷰 대기**

OpenSpec 작성자 기록이다. 리뷰가 아니다. 문서만 다뤘다(새 AST 0). 소스 · 시험 · 브로커 0.

| 발견 | 12판 처분 | 위치 |
| --- | --- | --- |
| P0 401 재전송 3건 버스트가 계좌 공유 3 req/s 한 초 창을 넘을 수 있음 | 하한가 GET 은 a166 계좌 공유 admission 이 주입될 때만, wire 요청마다(401 재전송 포함). nil → GET 0 · 판 비어 오늘 LIMIT. 엔진 배선은 a166 전까지 nil — 생산 변화 0(활성화 조건, 사람 확인 D-2 ⑥). `doRequest` 429 기록은 a166 으로 옮겨 a087 `doRequest` 편집 0. 시험 셋 | design D15.1 · spec Req 1 · tasks 2.4 · 2.10 · issues D-2 ⑥ |
| P1 Action 열거 규칙 | `go/types` 로 타입이 `exitpolicy.Action` 인 상수 전부를 모아 `go/constant` 로 평가한 값 집합 비교, 같은 값 두 이름 기록, 비면 실패 | design D15.2 · tasks 2.9 |

불변식 4 표는 design D15.3 이 대체한다(a166 전 전부 0). Function Logic Map: not-applicable — 문서 개정만 했다.

## 12판 재리뷰 (GPT-5.6 Terra, 2026-10-29)

**결론: FAIL (blocker 1).** D15.1의 nil 기본값은 하한가 GET 0·빈 판·기존 오늘 LIMIT 사다리로 fail-closed이며, D15.2의 평가된 `Action` 값 집합은 이름 존재 검사 함정을 닫는다. 그러나 활성화 뒤에도 401 재전송 **매 wire** admission을 보장하는 a166 계약이 아직 없다. a166은 `doRequest`와 `send`를 모두 삽입 후보로만 남겼고, `send` 하나에 admission을 두면 첫 승인 뒤 401 재전송 두 wire가 통과한다. 이 상태에서 admission을 주입해 활성화하면 계좌 공유 3 req/s 보호가 다시 깨질 수 있다.

| 심각도 | 발견 | 위치 | 근거 | 처분 제안 |
| --- | --- | --- | --- | --- |
| P0 | D15는 “각 wire”를 요구하지만 이를 만족하는 a166의 확정 인터페이스·삽입 자리·실제 관문 종단 시험을 묶지 않았다. 현 `send`는 최초 `doRequest` 뒤 401 loop에서 다시 `doRequest`를 최대 두 번 부른다. 따라서 admission을 `send` 입구에 한 번만 두는 a166 구현도 후보 문서와 양립하지만, 401→401→200에서 한 승인으로 3 wire를 보낸다. a087의 갱신기 래퍼나 시험 더블은 unexported 재전송 loop의 둘째·셋째 요청을 대신 막지 못한다. | `design.md:498-510`; `tasks.md:1472-1474`; `openspec/changes/a166-account-shared-rate-limit/proposal.md:43-46,76`; `internal/official/client.go:191-206,320-365` | AST: `Client.send` B4의 최초·loop `doRequest` 호출, `Client.doRequest`의 유일한 `c.hc.Do` 자리. 손으로 읽음: a166은 proposal/tasks만 있고 design/spec/구현이 없으며 두 후보 중 하나를 고르지 않았다. | D-2 ⑥ 활성화 조건을 a166의 구체 SHA·공개 capability·통합 시험에 결속하라. admission은 `doRequest`의 각 `c.hc.Do` 직전(또는 매 `doRequest`마다 반드시 도는 동등 관문)에 한 번이어야 하며, 둘째 admission 거절이면 셋째 `hc.Do` 0을 **실제 a166 관문**으로 증명해야 한다. 그 전까지는 nil을 유지한다. |
| P2 | “생산 동작 변화 0”은 broker wire와 `PlaceIntent` 가격으로 한정해야 한다. D11/tasks 3.1은 a129 출력에 새 `price_source` 키를 추가하므로 nil admission에서도 관측 로그 바이트는 바뀐다. 판이 비면 `lower_limit`은 될 수 없고 값은 `observed` 또는 `baseline`이어야 한다. | `design.md:502-503`; `design.md:144-149`; `tasks.md:1568-1571`; `internal/app/engine/exit_order_log.go:101-114` | AST: 현 `exitOrderLineFor` B1은 `observed`가 비면 `Baseline`으로만 내리고, 현재 줄에는 아직 `price_source`가 없다. 손으로 읽음: tasks 3.1은 세 값을 필수 golden으로 요구한다. | D15.1의 “0”을 주문 가격·wire 동작으로 명시하고, nil/빈 판 golden에 `price_source ∈ {observed, baseline}` 및 `lower_limit` 부재를 명시해 로그 계약도 고정하라. |

### 확인한 것

- AST: 현 `exitOrderLineFor` B1과 `sellIntent`의 기존 fallback은 `observed`를 먼저 쓰고 비면 `Baseline`을 쓴다. D11·tasks 2.3의 pre-edit literal `PlaceIntent` golden과 D15 nil GET 0을 결합하면, board nil의 broker 주문 가격·유형·수량은 오늘 LIMIT 사다리와 같아야 한다. 이는 구현 전 문서 대조이며 새 `exitLimitPrice`의 AST는 아직 없다.
- AST: `Client.doRequest`의 유일한 transport 호출은 `c.hc.Do`; `send`는 최초 1회와 401 loop 최대 2회에서 이를 다시 부른다. 그러므로 request별 admission의 검증 위치는 outer `send`가 아니라 매 transport 전송을 포괄하는 관문이어야 한다. D15가 a087의 `doRequest` 편집을 0으로 둔 것 자체는 a166 단일 소유와 맞다.
- D15.2의 `go/types` 타입 동일성 + `go/constant` 평가값 집합은 파일명·식별자명·`Action("…")` 표기 차이를 보지 않는다. 값 집합과 `isProtective(Proposal{Action: Action(v)})` 참 집합을 비교하고 양쪽 공집합을 실패시키므로 단순 존재/개수 검사의 허점을 닫는다. 같은 문자열 값의 alias는 런타임 `Action`에서 구별 불가라 값→이름 기록만 하는 처분도 맞다.
- `analysis/evidence/enumerate.py`: `stale=0`. `openspec validate a087-a-protective-exit-is-a-market-order --strict --no-interactive`: 통과.

### 안 본 것·제한

- 12판은 설계 단계다. a087·a166 구현, 단위/종단/변이 시험, 전체 gate, 실제 공유 admission, 브로커·계좌·실주문은 실행하지 않았다. 운영·`mutating` 명령은 0이다.
- a166의 wire 관문, 429 수명, 공정성·1초 창 배분은 아직 정의/측정되지 않았다. D15의 nil fail-closed는 이를 숨기지 않지만, P0 활성화 결속을 대신하지 않는다.
- `impeccable: N/A — backend/documentation-only`. `grill-me`: D-1·D-2 및 Manager 결정을 확정값으로 사용했고 새 사람 결정을 만들지 않았다. `search-first`: 구현·의존성 채택이 아닌 독립 문서/소스 적대 검토라 현재 저장소·a166 OpenSpec·AST만 대조했다.

## 13판 개정 (2026-10-29, Manager 결정 — 12판 재리뷰 FAIL 처분) — **독립 재리뷰 대기**

OpenSpec 작성자 기록이다. 리뷰가 아니다. 문서만 다뤘다(새 AST 0). 소스 · 시험 · 브로커 0.

| 발견 | 13판 처분 | 위치 |
| --- | --- | --- |
| P0 a166 의 `send` 단위 admission 후보는 401 재전송 두 wire 를 우회 | 활성화 조건 = a166 이 `doRequest` 의 `c.hc.Do` 직전에서 wire 요청마다 admission 강제(401 재전송 포함, `send` 단위 불가) **그리고** 그 통합 시험이 a166 착지 sha 에 있음. 엔진 배선을 non-nil 로 바꾸는 커밋은 그 시험 이름 · sha 를 인용하는 별도 착지(tasks 5.5). a166 쪽 요구는 issues D-2 ⑦ | design D15.4 · tasks 2.4 · 5.5 · issues D-2 ⑦ |

불변식 4 표(D15.3)는 그대로다(a166 전 전부 0). Function Logic Map: not-applicable — 문서 개정만 했다.

## 13판 재리뷰 (GPT-5.6 Terra, 2026-10-29)

**결론: proposal-freeze 재리뷰 PASS — Manager 승인 대기.** 12판 P0는 D15.4가 “a166 착지”가 아니라 `doRequest`의 매 wire admission·실제 `official.Client` 통합 시험·그 sha를 인용하는 **별도 5.5 활성화 커밋**으로 바꿔 닫았다. 현재 a166은 아직 이 계약을 갖지 않지만, 그 사이 a087의 admission은 nil이고 하한가 GET은 0이다. 따라서 미완성 a166은 현재 보호 매도를 약화시키지 않는다.

| 심각도 | 발견 | 위치 | 근거 | 처분 제안 |
| --- | --- | --- | --- | --- |
| P1 | “비시험 코드의 admission 주입 자리 0”은 목표만 있고, helper 반환값·다른 필드명·옵션 closure를 거친 non-nil 값까지 포함해 어떻게 세는지 정하지 않았다. `admit` 식별자나 특정 필드의 존재 수만 세면 helper 안의 non-nil 주입을 놓치며, 반대로 이름을 바꾼 정상 구현을 오판할 수 있다. | `tasks.md:232-235,268`; `design.md:505,520-524` | 손으로 읽음: 새 갱신기/필드의 구체 API는 아직 없고 task는 “구조 시험 0”만 적는다. AST: 현 `NewContext`와 `Context.ExitObserver`는 새 admission 값을 아직 갖지 않아 현재 source에서 이름 독립 census를 검증할 대상은 없다. | 2.0 Pre-Edit에서 sole refresher constructor와 admission capability의 타입·생성 경로를 고정하고, 2.4 구조 시험은 비시험 Go 파일 전체의 그 constructor/capability 입력을 열거해 **상수 `nil`만** 허용하라. 직접 인자·식별자·selector·helper 반환·옵션 closure를 모두 추적 불가로 분류해 실패시키고, helper로 non-nil을 넘기는 변이를 추가하라. 5.5에서만 이 단언을 실제 a166-owned capability로 바꾼다. |
| P2 | “생산 동작 변화 0”은 여전히 broker wire·`PlaceIntent` 가격의 무변화로 한정해야 한다. a129은 새 `price_source` 출력 키를 받으므로 nil admission에서도 로그 바이트는 바뀐다. 빈 판에서는 `lower_limit`이 아니라 `observed` 또는 `baseline`이어야 한다. | `design.md:505-506`; `design.md:144-149`; `tasks.md:1568-1571` | AST: 현 `exitOrderLineFor` B1은 관측가가 비면 기준선만 쓰며, 현재 출력에는 `price_source`가 없다. 손으로 읽음: tasks 3.1은 세 출처 값을 golden으로 요구한다. | D15의 “0” 범위를 주문 가격·wire로 명시하고, nil/빈 판 golden에 `price_source ∈ {observed, baseline}` 및 `lower_limit` 부재를 적어 둔다. |

### 확인한 것

- AST: `Client.send`은 최초 `doRequest` 뒤 B4 401 loop에서 최대 두 번 더 `doRequest`를 부른다. `Client.doRequest`의 유일한 transport 호출은 `c.hc.Do`다. D15.4의 `c.hc.Do` 직전·둘째 거절 뒤 셋째 wire 0 요구는 이 실제 분기와 맞고, `send` 입구 한 번의 admission은 명시적으로 배제됐다.
- 손으로 읽음: 현재 a166은 proposal·tasks·초안 spec delta뿐이며 `doRequest` 삽입/공개 capability/통합 시험을 아직 계약화하지 않았다. 그러나 D15.4·tasks 5.5·issues D-2 ⑦은 그 사실을 활성화 전제로 바꾸고, a166 sha·시험 이름이 없는 커밋에는 non-nil 배선을 허용하지 않는다. 현행 nil은 board 없음 → 오늘 LIMIT이라 fail-closed다.
- D15.2의 `go/types` 타입 동일성 + `go/constant` 평가값 집합은 action 이름·파일·선언 표기 차이를 보지 않고, `isProtective` 참 값 집합과 질의 집합 및 두 공집합 실패를 비교한다. 11판 P1 존재 검사 함정은 다시 열리지 않았다.
- 첫 보호 시도 단조 규칙·원장 오류 fail-closed·시작 정숙·가격 사다리/`price_source`의 기존 시험 요구도 D15.4로 완화되지 않았다. D15.3은 admission 전 1~4를 0으로 유지한다.
- `analysis/evidence/enumerate.py`: `stale=0`. `openspec validate a087-a-protective-exit-is-a-market-order --strict --no-interactive`: 통과.

### a166 착수 시 문서로 옮길 요구

- `official.Client.doRequest`의 매 `c.hc.Do` 직전에 계좌 공유 admission을 한 번 적용한다. `send` 단위 admission은 금지하고 401 재전송도 같은 관문을 지나야 한다.
- a166-owned, account-keyed admission capability와 a087이 쓰는 `*official.Client` 인스턴스가 동일함을 계약화한다. a087의 별도 preflight가 실제 토큰을 또 소비하거나, 다른/dummy admission이 활성화를 가능하게 해서는 안 된다.
- 실제 `official.Client` 통합 시험에 401 → 401 → 200의 admission 3회, 둘째 거절 뒤 셋째 `c.hc.Do` 0, 첫 거절의 `c.hc.Do` 0을 고정한다. a087 5.5는 이 시험 이름·착지 sha를 인용한다.
- 헤더 없는 HTTP 429도 a166 admission 거절로 표현하는 수명·재시작 처분을 계약·시험으로 정한다. 이는 a087에서 제거한 `doRequest` 429 기록의 소유권이다.

### 안 본 것·제한

- 13판은 문서 단계다. a087·a166 구현, unit/통합/변이 시험, 전체 gate, 실제 공유 admission, 브로커·계좌·실주문은 실행하지 않았다. 운영·`mutating` 명령은 0이다.
- a166의 공유 키·공정성·429 수명·wire 관문은 현재 정의/실측 전이다. 이것들은 5.5 전 a087 nil 기본값을 바꾸지 않지만, 위 a166 요구와 사람 승인 없이는 활성화할 수 없다.
- `impeccable: N/A — backend/documentation-only`. `grill-me`: D-1·D-2 및 Manager 결정을 확정값으로 사용했고 새 사람 결정을 만들지 않았다. `search-first`: 구현·의존성 채택이 아닌 독립 문서/소스 적대 검토라 현재 저장소·a166 OpenSpec·AST만 대조했다. `ai-memory-retrieval`: 이 세션에는 `memory_query` 등 MCP가 없어 현재 source와 대화 내 결정만 검증에 썼다.

## 구현 리뷰 (GPT-5.6 Terra, 2026-10-29)

**결론: 조건부 PASS — a087 구현 범위의 보호 매도 가격 변경은 통과함.** admission은 production에서 literal `nil`로 고정되어 a166 전 하한가 GET이 0이며, 휴면 보호 매도는 기존 LIMIT `PlaceIntent`·wire body를 유지함. a166 매 wire admission과 사람 실측/활성화는 아직 통과 판정 대상이 아님.

### 구현·적대 검토 발견과 처분

| 심각도 | 발견 | 처분·증거 |
| --- | --- | --- |
| P0 | 하한가가 아닌 fallback 보호 매도는 첫 시도 금지를 남기지 않아 같은 날 다음 하한가 시도가 가능했음 | `Place` 반환 직후 모든 보호 action에 `markProtectiveAttempt`를 적용함. RED는 `analysis/red-output.txt`, GREEN은 `TestA087FallbackProtectivePlaceStillClosesFirstAttempt`로 확인함. |
| P0 | positions/attempt 원장 읽기가 주기를 넘기면 다음 관측과 하한가 GET가 겹칠 수 있었음 | 사전 원장 읽기에 `interval - readLimit` deadline을 두고, 만료 시 reader 시작 전 반환함. `TestA087ExpiredPrefetchContextSkipsLowerLimitWire`·`TestA087ExpiredAttemptReadSkipsLowerLimitWire`가 각각 변이를 잡음. |
| P1 | 실제 broker wire마다 admission을 강제하는 구현은 없음 | 의도된 a166 소유 범위임. a087 production wiring은 literal `nil`; 구조 시험이 non-nil 흐름을 거부함. 5.5는 a166의 `doRequest`별 실제 `official.Client` 통합 시험·착지 SHA와 사람 승인이 있어야만 가능함. |
| P2 | cycle마다 positions 전체 조회와 종목별 journal EXISTS, snapshot 복사는 보유 종목 수에 따라 비용이 증가함 | schema v36 유지·새 auxiliary 한 종목/cycle 계약을 우선하여 현 change에서 재설계하지 않음. 실측 부하/인덱스 근거 없음; 활성화 전 용량 점검 위험으로 남김. |

### 확인 결과

- 불변식 4 판정: **조건부 충족**. nil admission에서 `Run`은 취소까지 대기해 하한가 GET 0이고, `TestA087DormantBoardKeepsLegacyPlaceIntent`는 기존 LIMIT 가격 `67900`, 직렬화 wire body, journal price가 동일함을 고정함. `TestA087RunSendsLowerLimitTokenBetweenObserveAndSleep`는 token이 관측 뒤·sleep 앞에만 보내지고 비차단임을 고정함. 새 로그 `price_source`는 D15가 허용한 별도 관측 정보이며 broker wire에는 없음.
- 가격·시점: KR 평일 `[09:00, 15:30)`의 첫 보호 매도만 유효한 당일 하한가를 사용하며, 정수 canonical·당일 timestamp·하한가 ≤ 관측가가 아니면 기존 관측가/기준선 LIMIT으로 fallback함.
- 재시작: journal SELECT-only 존재 질의가 account/market/trading day/symbol/protective action으로 scope를 고정하며 상태와 가격을 읽지 않음. 오류는 하한가 사용 금지임.
- concurrency: snapshot은 atomic read와 writer lock을 쓰며 `-race` A087 focused tests를 통과함.
- gstack 독립 적대 경로에서 위 P0 두 건을 찾고 수정 후 재검증함. gstack 자체 업그레이드 제안은 범위 밖이라 수행하지 않음. PR/Greptile 입력은 없었음.
- `openspec validate a087-a-protective-exit-is-a-market-order --strict --no-interactive` 통과. 영향 패키지 `go vet`, 일반·`tossos_testseams` `go test`, A087 focused 일반·tagged `-race`를 통과함.
- `check_analysis.py --change a087-a-protective-exit-is-a-market-order`는 이 change의 `base-commit.txt`가 오래되어 base→working tree의 기존 함수 378개 증거를 요구해 실패함. 이번에 편집한 7개 FLM/BTM의 stale 오류는 재생성 후 없음. base 재고정·`--record-landing`은 금지 지시대로 하지 않음.

### 범위·미실행·남은 위험

- `doRequest` 편집, schema 변경, 실제 broker/계좌/주문, 운영 토글, 사람 실측 5.4, 활성화 5.5, 전체 `go test ./...`, `make gate`는 실행하지 않음.
- a166이 매 `c.hc.Do` 직전 admission을 401 재전송까지 강제하고 실제 client 통합 시험으로 증명하기 전에는 non-nil admission 배선을 금지함.
- timestamp가 실제 효력일인지와 KRX 휴장/기준가 규칙은 5.4 사람 gate 전 미검증임. 이 구현 PASS는 주문 활성화 또는 실제 체결 PASS가 아님.
- `impeccable: N/A — backend/documentation-only`. 화면·접근성 변경 없음.

## 구현 Pre-Edit · a165 재대조 (2026-10-29, HEAD `a284a2b0a`)

### a165 재대조 — 1.0

- a165 착지 `5e2facaee`는 archive 이동만 포함하며, 현재 `main`의 `clearTheSymbol`·`checkSymbolFree`·공유 `fakeSubmitter`를 다시 읽었다.
- `checkSymbolFree`는 노출 증가 주문만 a165 latch를 읽고, exit/flatten/reconcile 축소 주문은 그 입력을 읽지 않는다. a087은 `ExitObserver.submit`의 LIMIT 가격만 고치므로 이 경계를 넓히지 않는다.
- `clearTheSymbol`과 `cancel_refusal*.go`의 취소·정리 계약은 a087에서 호출하거나 수정하지 않는다. `NewOrderPath` 호출부·`buildFlattenWiring`도 무변경 범위다.
- `analysis/evidence/tossctl--buildflattenwiring.ast.json`을 현 HEAD로 재생성했고 `enumerate.py`는 `stale=0`이다. 공유 `fakeSubmitter`는 기존 시험 보조물로 재사용하며, a165와 충돌하는 state를 추가하지 않는다.

### Pre-Edit Gate — 2.0

- change id / task id: `a087-a-protective-exit-is-a-market-order`, 1.0·2.0·2.1
- 대상 심볼: `ExitObserver.submit`, `ExitObserver.Run`, `exitOrderLineFor`, `exitOrderLine.args`, `NewContext`, `Context.ExitObserver`, `engineRuntime`; 새 파일은 engine board·official raw reader·journal read-only reader.
- CodeGraph definition/callers/callees/impact: `submit`은 `sellIntent`와 `Gateway.Place` 사이의 LIMIT 가격 인자만 받는다. `Run`은 `ObserveOnce` 뒤·`Sleep` 앞에만 token을 보낸다. `NewContext`가 board를 한 번 만들고 `Context.ExitObserver`가 caller 값과 무관하게 덮는다. `engineRuntime`은 Auxiliary에만 새 executor를 넣는다.
- CodeGraphContext 후보와 evidence reconciliation: 현재 source·AST·a165 archive를 우선으로 대조했다. 오래된 `buildFlattenWiring` AST 하나만 재생성했고 나머지 43개 증거는 fresh다.
- 기존 동작 파악 근거: Pre-Edit HEAD `a284a2b0a`, `exitloop_test.go`의 real journal/Guardian + `fakeSubmitter`, a129 log golden, cmd a129 runtime wiring structure test.
- Function Logic Map / Branch Test Map: `analysis/function-logic/`의 7개 번들(각 `ast.json`·FLM·BTM·risk report). `official.Client.doRequest`는 D15.1에 따라 제외하며 변경 0이다.
- upstream 상속 테스트 영향: yes. LIMIT order gate·a129 line whitelist·runtime auxiliary 목록이 영향권이며, new tests와 기존 focused tests로 회귀를 막는다.
- 실패 테스트 선행 작성: yes. RED commit은 board price table, nil admission GET 0/literal PlaceIntent, first attempt/restart, wire admission double, raw reader, journal reconstruction, price_source를 포함한다.
- 설정·DB·journal 변경과 rollback: config·schema 변경 없음(v36 유지). journal은 SELECT-only function 추가이며 오류는 lower-limit 사용 금지로 fallback한다. rollback은 board admission을 nil로 유지하거나 code revert이며 주문 type·quantity는 변경하지 않는다.

### SDD 적용 기록

- `search-first`: 기존 `PriceLimits`는 raw null/timestamp를 유실하므로 재사용하지 않는다. 공식 client의 read path·engine Context/private `official`·`ManualBuyRefresher`·표준 `math/big.Rat`를 확인했다. 새 외부 의존성은 채택하지 않는다.
- `grill-me`: D-1(다), D-2 ①·⑥, a166 전 nil dormant가 확정되어 새 사람 결정·가정이 없다.
- `impeccable: N/A — backend/documentation-only`; UI·접근성·화면 상태 변경 없음.

## 추가 검증 Pre-Edit (2026-10-30)

- 범위: `exitLowerLimitBoard.refresh`의 전일 판 항목이 당일 갱신을 막는 결함. 이 함수는 새 함수이며 기존 FLM의 적용 대상은 아니지만 현재 AST `/tmp/a087-terra-refresh-ast.json` B7과 원본을 대조했다. 동일 날짜 유효 항목만 건너뛰고 오래된 항목은 기존 검증과 원장 조회를 거쳐 교체한다. 실패 시 기존 항목은 가격 선택에서 거부된다.
- search-first: 저장소 `recordValidNow`, 기존 board 및 Go context/time을 재사용한다. 새 의존성·라이선스 변경 없음. 외부 패키지 탐색 불필요. ai-memory 해당 질의 결과 없음.
- 별도 미해결: 지연 토큰이 다음 관측과 겹치는 문제는 deadline만으로 시작 경합을 막을 수 없어 Manager 설계 판단 전 수정하지 않는다. 생산 admission nil 유지.

### 추가 검증 결과 · 미해결

- P1: `exit_lower_limit.go:210`의 존재만 확인하는 분기는 전일 판을 당일에도 건너뛰어 당일 GET을 막았다. `recordValidNow`로 당일 유효 항목만 건너뛰도록 수정. RED `analysis/20261030-daily-refresh/red.log`, GREEN race `analysis/20261030-daily-refresh/green-race.log`; 전후 AST는 같은 폴더에 있다. 기존 분기를 되돌린 원본이 RED이므로 그 변이의 검출을 확인했다.
- P1 활성화 차단: D12.2(`design.md:265-284`)는 관측 `/prices`와 GET 겹침 0을 요구한다. 손으로 읽음: `exit_lower_limit.go:142-149,176-180,191,221`은 토큰 생성 시점이 아닌 소비 시점부터 timeout을 주므로 지연 토큰이 다음 관측과 겹칠 수 있다. 동기화 방식은 Manager 설계 판단을 기다린다. production admission nil 유지, 5.5 활성화 금지.
- `analysis/evidence/enumerate.py`의 이전 설계 증거는 소스 파일 hash 기준 stale 14; 이는 현재 구현 FLM 번들과 별도다. 기존 base→착지 창의 378개 형제 함수 증거 누락은 해결되지 않았다. base 변경이나 landing 기록은 하지 않았다.
- gstack 범위 checklist와 독립 Terra 수동 리뷰를 수행했다. 독립 리뷰는 수정 분기의 새 안전 결함을 발견하지 않았고 원장 기시도·오류 조합 시험 공백을 지적했다. 해당 조합은 `TestA087PreviousDayRecordFailsClosedOnAttemptOrJournalError`의 GET 0·관측가 fallback 시험으로 확인했다 (`analysis/20261030-daily-refresh/combination-race.log`). Claude·Codex 독립 음성은 실행하지 않아 gstack skill 전체 절차 및 최종 gate는 미완료다. 5.4 사람 실측과 5.5 활성화 미실행, archive·운영 완료 아님. `impeccable: N/A — backend/documentation-only`.

### 2026-10-31 실제 게이트 종단 시험

`TestA087RealGatewayLowerLimitAndDormantFallback`은 테스트 전용 판 주입과 nil 판을 각각 실제 RiskGuardian → Gateway → Trading → official.Client → httptest 수신까지 연결한다. KR 정규장 하한가 50400, 판 없음 67900, 정규장 밖 67900의 LIMIT SELL wire를 검증했다. `go test -race -count=1 -run '^TestA087RealGatewayLowerLimitAndDormantFallback$' ./internal/app/engine` 통과, 영수증 `analysis/20261030-daily-refresh/real-gateway-race.log`. 실제 주문·운영 설정 변경 없음. 지연 토큰 겹침과 5.4·5.5는 여전히 차단이며 task 전체 완료를 주장하지 않는다. 별도 독립 검토 미완료.

### 2026-10-31 관문·생성 위치 시험

- 1.1: `TestA087EngineAndDefaultServiceRejectNonfractionalMarket`는 실제 엔진 Context와 기본 trading Service의 KR/US buy/sell 시장가가 `LiveReady=false`임을 확인한다. `LiveReady`는 운영 토글과 별도로 지원 여부를 나타낸다. `TestA087NonfractionalMarketShapeRejected`는 동일 네 행의 execgw 사유를 확인하고, 기존 riskcalc 시장가 진입 거부 두 시험도 통과했다. 브리지는 `_test.go` 전용이다.
- 1.2: `TestA087ProductionBoardAdmissionFlowIsOnlyLiteralNil`은 go/types로 engine 생산 Go 파일 전체의 생성자 참조가 `NewContext` 한 곳뿐임을 검사한다. 생성자는 비공개이므로 다른 패키지는 직접 호출할 수 없다. `git archive HEAD` 격리 사본의 정상 시험 PASS 후 생성자를 `NewContext`에서 `NewOrderPath`로 이동한 변이는 해당 위치 오류로 FAIL했다. 영수증은 `analysis/20261030-daily-refresh/a087-12-{baseline,relocation-mutant}-20261031.log`.
- Manager가 세 신규 파일과 `PreviewPlace`의 판단 순서를 독립 수동 검토하여 blocking finding 없음. 전 범위 gstack 인증은 아님. UI 미변경: `impeccable: N/A — backend/documentation-only`.

### 휴면 자원 사용 Pre-Edit (2026-10-31)

`exitLowerLimitBoard.cycleComplete`와 `markProtectiveAttempt`는 a087 신규 함수다. 현재 nil admission에서도 토큰과 일자별 시도 map을 채운다. 두 함수에서 admission nil일 때 즉시 반환해 휴면 상태의 자원 누적을 없앤다. non-nil fixture에서 기존 동작과 첫 시도 금지를 유지한다. `a087_export_test.go` 및 `TestA087FirstProtectiveAttemptIsMonotonic`은 하한가를 인위 주입하므로 활성 fixture에만 non-nil test admission을 명시한다. 기존 시험 함수 변경 전 원본과 실행 분기를 확인했다. Go 기존 `nil` 검사 및 repository 시험 더블 재사용, 새 의존성 없음. 활성 겹침 문제는 별도 미해결, 생산 admission은 nil 유지.

### 2026-10-31 휴면 자원 수정과 1.6 최종 근거

- `cycleComplete`·`markProtectiveAttempt`는 nil admission이면 즉시 반환한다. 기존에는 휴면 상태에서 토큰과 일자별 map이 누적됐다. 각 함수 선행·후행 AST는 `analysis/20261030-daily-refresh/{cycleComplete,markProtectiveAttempt}-{pre,post}.ast.json`. 선행 RED는 토큰 1·시도 1, 후행 일반·`tossos_testseams` A087 집중 race 각각 42.464초·43.769초 GREEN이다. 활성 fixture는 test admission을 명시해 기존 첫 시도 금지 동작을 시험한다. nil 생산 배선은 그대로다.
- 함수 분기표: 두 함수의 B1은 `board == nil || admission == nil`이면 return; non-nil일 때 `cycleComplete`는 채널 select 송신 또는 default, `markProtectiveAttempt`는 거래일 오류면 return, 성공 시 lock/map 기록이다. RED/GREEN과 기존 `TestA087FirstProtectiveAttemptIsMonotonic`, A087 활성 wire 시험이 이를 확인한다. 신규 함수 본문에 한정한 변경이며 관측 겹침 문제는 미해결이다.
- 1.6: `TestA087RealGatewayLowerLimitAndDormantFallback`은 실 Guardian·Gateway·Trading·official.Client와 httptest를 사용한다. 하한가 50400, 판 없음·창 밖 67900 세 행의 실제 전송 body 전체를 골든과 비교한다(동적 `clientOrderId`는 비어 있지 않음을 별도 확인). 원본 `analysis/20261030-daily-refresh/real-gateway-full-wire-race.log` GREEN. Manager가 시험과 영수증을 수동 검토했다. 실제 broker·live 검증은 아님.
- `price_source` 로그는 휴면 경로에도 추가되는 관측 정보다. 사용자 승인된 dormant 착지 검토와 별개로 5.4·5.5, 겹침 0, gate ⑤ 형제 증거, 전체 gstack 절차, archive는 미완료다.

## 2026-10-31 D16 문서 정합 검토

Terra 독립 문서 검토에서 보호 우선·대기 정책과 a166 gate의 P1 거절 경계를 수정 후 PASS로 확인했다. a087 scoped OpenSpec 1.4.1 strict 검증과 PM tracker 검사, diff whitespace 검사를 수행했다. 문서 범위 판정이며 코드 시험·gstack 전체 검토·전체 gate 및 운영 검증은 수행하지 않았다. 5.4·5.5 미완료와 admission nil을 유지한다.
