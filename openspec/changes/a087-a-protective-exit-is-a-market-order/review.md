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
