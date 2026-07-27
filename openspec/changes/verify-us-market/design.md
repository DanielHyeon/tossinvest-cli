# Design: verify-us-market

## A1. 왜 기록을 시장별로 나누는가 (그리고 왜 step id를 나누지 않는가)

증거 기록은 StepID로 색인된다. `settled`는 "이 단계에 terminal 판정이 있으면 다시 실행하지
않는다"이고 `RedoSet`은 그 반대다. 한 파일에 KR과 US를 섞으면 US에서 통과한
`conditional-register`가 KR의 같은 단계를 settled로 만든다 — **측정하지 않은 시장의 능력을
측정한 것으로 만드는** 최악의 실패다.

세 가지 방법이 있었다.

| 방법 | 결과 | 판정 |
|---|---|---|
| step id를 시장별로 복제(`conditional-register-us`) | 카탈로그가 두 배, 모든 의존 관계·라벨·리포트가 두 배 | 기각 |
| 엔트리에 `market` 필드 추가 + settled/RedoSet/progress를 (step, market)로 | 스키마·resume·digest·리포트가 전부 영향 — M3 resume 증거가 걸린 기계를 건드린다 | 기각(지금은) |
| **기록 파일을 시장별로 분리** | 기존 기계 전부 무변경, KR 파일 바이트 무접촉 | **채택** |

capability는 (계좌, 시장)의 속성이므로 파일 분리는 임시방편이 아니라 그 사실의 표현이다.
리포트·attestation도 시장별로 읽으면 되고, 2c는 시장별 ProtectiveCapability를 그대로 받는다.

기본 경로: KR `capability-verify.jsonl`(무변경), US `capability-verify-us.jsonl`.
`--record`로 명시하면 그 값이 이긴다(기존 플래그 의미 보존).

## A2. preflight의 시장 판정

현재:

```go
if step.Mutates && MarketOf(r.mutationSymbol(step)) != MarketKR { skip }
```

이 줄은 두 가지를 동시에 말하고 있었다 — "이 도구는 KR만 안다"와 "이 run은 KR이다".
전자는 더 이상 참이 아니므로 후자만 남긴다:

```go
if step.Mutates && !SameMarket(MarketOf(r.mutationSymbol(step)), r.market) { skip }
```

의미는 보존된다: **run의 시장과 다른 심볼에는 주문을 내지 않는다.** KR run에서 US 심볼이
섞여 들어오면 종전과 똑같이 건너뛴다. 사유 문구는 시장을 이름으로 말한다.

## A3. US 정정이 가격 전용인 이유는 브로커 계약이다

`OrderModifyRequest.quantity`: "**KR 주식: 필수.** … US 주식: 전달 불가. 제공 시
`400 us-modify-quantity-not-supported`."

그래서 `amendOrder`는 시장에 따라 `Quantity`를 채우거나 비운다. 이것은 우리가 US에서 덜
측정하는 것이 아니라, **브로커가 US에서 정의한 정정이 가격 정정**이라는 사실을 그대로
시험하는 것이다. 승인 목록의 문구도 그에 맞춰 "US는 정정에 수량을 보내지 않는다"로 바뀐다 —
사람이 승인하는 목록은 실제로 나갈 요청과 같아야 한다.

측정 가치는 그대로다: 정정이 **새 conditionalOrderId/orderId를 발급하는지**가 2c 귀속
규칙의 입력이고, 그 질문은 US에서도 동일하다.

## A4. 가격: 새 산식 없음

이미 있는 것을 확인만 하고 쓴다.

- `PriceLimits`는 US에서 0/0을 반환한다(`internal/official/price_limits_reads.go`: "US
  stocks have no limits; UpperLimit and LowerLimit will be 0").
- `FarBuyLimit`/`FarSellLimit`는 `lowerLimit > 0` / `upperLimit > 0`일 때만 클램프한다.
  밴드가 없으면 오프셋이 가격을 정하고, 결과가 여전히 시장 반대편인지 확인하는 검사는
  남는다.
- `TickSize`는 US를 이미 구현한다($1 미만 0.0001, 이상 0.01) — `OrderCreateRequest.price`의
  절삭 규칙과 같은 표.

따라서 이 change는 pricing.go에 **산식을 추가하지 않는다**. 바꾸는 것은 주석 하나(`MarketKR`가
"유일한 시장"이라는 문장)와 그것을 근거로 삼던 preflight뿐이다.

경계 위험 하나는 남는다: 밴드가 없는 시장에서 20% 떨어진 지정가를 브로커가 "비합리적
가격"으로 거부할 수 있는지는 **[미측정]**이다. 거부되면 그것이 M-번호 실측이 되고, 그때
오프셋을 조정한다. 거부는 주문이 나가지 않았다는 뜻이므로 안전 방향이다.

## A5. 세션 advisory는 기존 달력을 쓴다

`internal/clock`이 이미 시장별 정규장을 안다(KR 09:00–15:30, US 09:30–16:00, IANA 존으로
DST 처리, 휴일표는 의도적으로 없음). verifylive에 두 번째 US 달력을 쓰면 order-execution
"시간 규율"이 요구하는 판정이 저장소에 둘이 된다.

문구의 정직성 규칙은 유지한다:

- KR: 실측된 `order-hours-closed` 422를 계속 인용한다(M1).
- US: **아직 실측이 없다.** "US 정규장 밖이다. 휴장 시 브로커 응답은 이 계좌에서 아직
  관측되지 않았다([미측정])"로 쓴다. 관측하지 않은 것을 KR 근거로 단언하지 않는다.

advisory는 계속 advisory다 — 어느 시장에서도 시작을 막지 않는다.

## A6. 심볼 선택

| | KR | US |
|---|---|---|
| 매수측 프로브 | `005930`(상수, 무변경) | 계좌의 사용 가능한 US 보유 종목 |
| 보유 필요 단계 | 사용 가능한 KR 보유 | 사용 가능한 US 보유 |

US 프로브 심볼을 상수로 박지 않는 이유: 계좌마다 보유가 다르고, 매수 프로브와 매도·조건주문
단계가 **같은 심볼**이면 노출 심볼이 하나로 줄고 opposite-pending(422) 관측도 자연스럽다.
`--symbol`/`--holding-symbol`로 덮어쓸 수 있다(기존 플래그).

사용 가능 판정은 기존 그대로다: 수량 ≥ `MinQuantity`(1주). 소수점 보유(TSLA 0.000154주)는
자동으로 제외된다 — 이 도구는 소수점 주문을 내지 않는다.

## A7. 콘솔

`/verify?market=us`. 선택된 시장의 기록으로 진행률·이어하기·재측정·단계 목록·리포트를
렌더하고, 시작 폼이 그 시장을 함께 보낸다. 승인 화면은 무변경이다(클릭 1회).

한 프로세스는 여전히 검증을 한 번만 수행한다(`errProcessSpent`) — 조건주문 존속 측정의
프로세스 경계 규칙이다. 따라서 KR과 US를 이어서 측정하려면 사이에 [콘솔 재시작]이 필요하고,
화면이 그렇게 안내한다.

## A8. 이 change가 하지 않는 것

엔진의 US 매매를 켜지 않는다. `ProtectionReady`는 여전히 미충족 상수이고, 무엇을 자동
매매해도 되는지는 2c의 2.6(미검증 시장·유형 금지 목록)과 §0.7 게이트가 정한다. 이 change의
산출물은 **그 판단에 넣을 US 실측**이다.
