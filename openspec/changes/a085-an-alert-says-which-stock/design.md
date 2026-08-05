# a085 · Design

## D1 — 종목명은 이미 도착해 있다

공식 `GET /api/v1/holdings`의 항목에 `name`이 있고 `official.apiHoldingsItem.Name`으로
파싱된다. float 경로(`adaptHoldings` → `domain.Position.Name`)는 그것을 채우고 콘솔이
쓴다. 엔진이 쓰는 raw 경로만 버린다.

```
apiHoldingsItem.Name  ──✓──>  domain.Position.Name     ──>  콘솔 화면
                      ──✗──>  official.RawHolding      ──>  reconcile.RawHolding
                                                       ──>  reconcile.Holding  ──>  엔진 알림
```

그러므로 이 change는 **새 데이터를 가져오지 않는다.** 이미 대금을 지불한 응답에서
버리던 필드를 줍는다. §0.4 예산은 그대로다.

**별도 메타데이터 API를 쓰지 않는다.** 콘솔에는 `InstrumentNameReader`가 있지만 그것은
*과거* 거래의 이름을 위한 것이고, 요청 예산과 lease를 따로 쓴다. 알림이 필요한 이름은
**지금 보유 중인 종목**의 것뿐이고, 그것은 대사 스냅샷에 전부 들어 있다.

## D2 — 비교 타입에 표시 필드를 넣는 것

`reconcile.Holding`은 수량 불일치를 판정하는 비교 타입이다. 거기에 표시 전용 필드를
넣는 것은 설명이 필요하다.

선례가 있다. `CostBasisRaw`가 같은 성격이다.

> Only the adoption record consumes it, and only to store it. Nothing compares or
> computes with it — AveragePrice is what the comparison uses.

`Name`도 같다. `Compare`는 심볼과 수량만 본다. 이름은 비교에 참여하지 않고, 비어 있어도
비교는 그대로 동작한다. 대안 — 이름만 나르는 두 번째 스냅샷 타입 — 은 같은 응답을
두 모양으로 들고 다니게 만들고, 두 모양이 어긋날 수 있는 자리를 새로 만든다.

**테스트로 고정한다:** 이름이 다른 두 holding은 여전히 같은 것으로 비교된다.

## D3 — registry의 수명과 빈 값

엔진은 심볼→이름 map을 대사 주기마다 갱신한다.

```
갱신    안정화된 스냅샷의 holdings로 덮어쓴다 (주기 60초)
읽기    알림 조립 시점
없음    코드만 쓴다
```

**보유가 사라진 종목의 이름을 지우지 않는다.** 청산 직후에 나가는 알림 — "외부에서
청산되었다", "판정이 거부되었다" — 이 이름을 잃는 것이 가장 나쁜 시점이다. map은
프로세스 수명 동안 누적되고, 보유 종목 수의 상한이 곧 크기의 상한이다.

**이름을 추정하지 않는다.** registry에 없으면 코드만 쓴다. `042660`이 `한화오션`이라는
것을 다른 출처에서 유추해 붙이면, 틀렸을 때 운영자가 엉뚱한 종목을 본다.

## D4 — 무엇을 한국어로 바꾸고 무엇을 두는가

```
바꾼다    Notification.Title, Notification.Body        사람이 읽는다
둔다      obs.Event.Fields (payload JSON)              기계가 읽는다
둔다      구조화 로그 필드 키와 값                      기계가 읽는다
둔다      EventType, ReasonCode, 원장 cause 문자열      기계가 읽는다
둔다      Go 에러 문자열                                개발자가 읽는다
```

경계는 "누가 읽는가"다. `alert_outbox.title`/`body`는 ntfy로 그대로 나가고 사람이 읽는다.
`payload`는 JSON이고 사후 조사가 파싱한다. 둘을 같이 번역하면 원장 질의가 언어에
의존하게 된다.

에러 문자열을 번역하지 않는 것도 같은 이유다. 알림 본문이 에러를 포함할 때는 한국어
설명 뒤에 원문 에러를 그대로 붙인다 — 운영자가 그것을 그대로 검색하거나 붙여넣을 수
있어야 한다.

## D5 — 격리 표시: 왜 감추는 것이 fail-closed가 아닌가

`operatorview`의 계약은 이렇게 적혀 있다.

> Stale and unknown evidence keeps provenance but hides actionable values behind
> em dashes.

이 규칙은 옳고, 격리가 그 규칙의 어느 쪽인지가 잘못 분류되어 있다.

```
unknown                      근거가 없다. 보여줄 값이 없다.              → 감춘다
stale (엔진 미실행, 관측 노후)  값이 있지만 얼마나 틀렸는지 모른다.        → 감춘다
격리                          값이 있고, 정확히 그 값을 붙들고 있으며,     → 보여준다
                             그것이 갱신되지 않는다는 것까지 알려져 있다.
```

세 번째는 앞의 둘과 다르다. 격리된 포지션의 보호선은 "오래된 값"이 아니라 **현재 유효한
값**이다 — 엔진이 그것을 갱신하지 않을 뿐, 그 선이 이 포지션에 대해 저장된 마지막이자
유일한 보호선이다. 감추면 운영자는 손절이 어디 있는지 알 수 없고, 그것은 판정이 멈춘
포지션에 대해 가장 알아야 할 하나다.

관리 화면이 이미 그렇게 판단했다는 것이 증거다 — a079의 해제 화면은 "유지 보호선"을
논거의 중심에 놓는다.

> That is why this screen argues from what is *kept* rather than what is fixed:
> the entry, the initial stop and the stored protection line all survive the
> release untouched.

포지션 화면만 그 값을 감춘다.

## D6 — 얼어 있음을 어떻게 구별시키는가

값을 보여주되 살아 있는 보호선과 같은 모양이면 안 된다. 화면은 이미 상태 축을 가지고
있다 — `Status`/`StatusText`가 `fresh`/`stale`/`unknown`이고 템플릿이 그것으로 분기한다.

격리는 `stale`로 남기되(판정이 멈췄다는 사실은 그대로다), 값이 채워진 stale이 된다.
템플릿은 stale일 때 값 옆에 "갱신 안 됨"을 붙이고, 기존의 "이 포지션은 지금 exit 판정
대상이 아니다 — 손절도 평가되지 않는다" 경고를 유지한다.

이 방식은 다른 stale 사유에 영향을 주지 않는다. 그것들은 계속 값을 채우지 않으므로
`—`로 남는다. 분기는 "사유가 격리인가"가 아니라 "값이 있는가"가 되고, 값을 채우는
사유는 격리뿐이다.

## D7 — RED이 되어야 하는 테스트

```
1. reconcile:    raw 경로가 holdings 응답의 이름을 Holding까지 나른다
2. reconcile:    이름이 달라도 수량 비교 결과는 같다 (비교 오염 없음)
3. engine:       registry가 대사 주기마다 갱신되고 보유가 사라져도 이름을 잊지 않는다
4. engine:       알림 제목·본문이 한국어이고 `이름(코드)`로 종목을 부른다
5. engine:       이름을 모르면 코드만 쓰고 추가 요청을 하지 않는다
6. engine:       payload와 구조화 로그 필드는 영문·원문 그대로다
7. operatorview: 격리 사유의 exit line이 저장값을 채우고 stale로 남는다
8. operatorview: 다른 stale 사유는 지금과 같이 값을 채우지 않는다
9. console:      격리된 포지션 행에 보호선이 그려지고 갱신 안 됨 표시와 경고가 함께 있다
```

1·4·7·9가 현재 코드에서 **실패**한다.
