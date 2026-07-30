# Proposal: size-us-guardian-tier

## Why

콘솔의 USD 1회 주문 금액 상한은 **$300**이다. 그 숫자는 결정된 것이 아니라 상속된
것이다 — `us-small-live`는 StockOS `risk_profiles.py`의 `_US_SMALL_LIVE`를 글자
그대로 옮긴 행이고, 상한은 "그 통화에 등록된 티어의 필드별 최대"이므로 StockOS의
미국 스모크 규모가 그대로 TossOS의 천장이 됐다.

그 천장이 TossOS가 실제로 재려는 종목과 정확히 겹친다. 이 저장소 자신의 실측:

> `verify-execution-capability/measurements.md` M49 (2026-07-30 US 정규장)
> **TSLA** — 틱 간격 0.18초/1초, 스프레드 **299.88–299.94 = 0.0200%**, 1주 **$300**

그리고 `internal/risk/chain.go`의 notional 검사는 `exceedsInclusiveLimit` —
경계 포함 허용이다. 즉 TSLA 한 주는 **막히지 않는다. 6~12센트 남기고 통과한다.**

문제는 불가능이 아니라 **헤드룸 0**이다. 세 가지가 겹친다.

1. 대기 중인 실측(`verify-execution-capability` 3.3)은 SINGLE+**MARKET**이다.
   시장가 주문은 상한을 검사하는 시점에 체결가를 모른다.
2. 6센트는 이 종목의 스프레드(0.02%)보다도 작다. 한 틱이면 넘는다.
3. 상한은 등록 티어의 최대이므로 콘솔에서 올릴 수 없다. 운영자가 할 수 있는
   유일한 일은 `config.json`을 손으로 여는 것이고, 그때 인터록에는 상한이 없으므로
   백스톱이 아예 사라진 채로 아무 값이나 적히게 된다.

즉 지금 구조는 "안전하게 낮다"가 아니라 **"백스톱을 우회해야만 쓸 수 있다"**이다.
그것은 백스톱이 가장 나쁘게 실패하는 방식이다.

두 번째 공백은 **승인된 USD 집합이 없다는 것**이다. `risk-management`의
`정책 수치의 provenance`가 승인한 다섯 숫자는 KRW 전용이고(주문당 1,000,000 KRW
외), 같은 요구가 "모든 한도·정책 수치는 코드에 출처(StockOS 파일·검증 상태)와
함께 기록되어야 한다(SHALL)"고 말한다. USD 축에는 그 출처 기록이 통째로 없다 —
있는 것은 다른 제품의 스모크 프로파일 두 줄뿐이다.

## What Changes

1. **미국 단일 종목 티어를 실측 근거와 함께 등재한다** — `us-single-name`
   (`$500` 주문 / `$1,500` 노출 / `$50` 일일 손실 / 비율 1% / 수량 100주 / USD).
   각 수치의 유도는 design D1에 적고, 코드 주석에 M49 실측과 승인된 KRW 집합
   양쪽을 인용한다. 프리셋 카드로 제시되며 **권장 기본값은 여전히
   `kr-small-live`이다**.

2. **`risk-management`의 provenance 요구가 TossOS 실측 출처를 인정하게 한다** —
   현재 문장은 출처를 "StockOS 파일·검증 상태"로 열거한다. TossOS가 스스로 잰
   숫자는 그 열거에 들어갈 자리가 없어서, 실측 근거를 가진 수치일수록 요구를
   만족시킬 수 없는 상태다. 승인된 USD 집합도 같은 문장에 기록한다.

3. **상한 이동을 명시적 계약으로 만든다** — USD 상한 세 필드가 올라간다는 사실과
   그 상한이 **콘솔 쓰기 경로 전용**이라는 사실(기동 인터록은 상한을 보지 않는다)을
   스펙과 테스트로 고정한다.

## Impact

- `internal/config/limits.go` — 티어 한 줄 추가, 파일 헤더의 provenance 서술 확장
- `internal/config/limits_test.go` — 전사 검사·상한 검사 갱신 + 유도 검증 신규
- `openspec/specs/risk-management` — `정책 수치의 provenance` 개정 (Requirement 수준)
- `openspec/specs/operator-console` — 티어 provenance 문구를 인용 권위와 일치시킴
- 콘솔 화면·핸들러·writer·audit 배선 — **무수정** (레지스트리를 읽을 뿐이다)
- 엔진·인터록·파싱 — **무수정**

### 상한 이동 (USD만)

| 필드 | 이전 | 이후 |
|---|---|---|
| `max_order_quantity` | 100 | 100 (불변) |
| `max_order_notional` | 300 | **500** |
| `max_total_exposure` | 1,000 | **1,500** |
| `max_daily_loss_amount` | 50 | 50 (불변) |
| `max_daily_loss_ratio` | 0.01 | 0.01 (불변) |

다섯 중 둘만 움직인다. KRW 상한은 다섯 필드 전부 불변이다.

## 범위 밖 (명시)

- **US 자동 진입을 켜지 않는다.** 이 change는 콘솔이 기록할 수 있는 값의 천장만
  옮긴다. 게이트 ON 기동은 여전히 인터록 6조항이 막고, 6번 ProtectionReady는
  보호주문 change 이전에 충족될 수 없다.
- **어떤 설정도 적용하지 않는다.** 파일은 운영자가 프리셋을 누를 때만 바뀐다.
- **`us-smoke`·`us-small-live`를 지우거나 고치지 않는다.** StockOS 전사 행이므로
  감사 대조 대상으로 남는다.
- **수량 상한 100주를 티어별로 나누지 않는다.** fat-finger 백스톱이고 금액 축이
  이미 먼저 조인다 (`console-sets-guardian-limits` design D4 유지).
- **FX 환산을 코드에 넣지 않는다.** KRW 등가 논증은 이 문서의 §0.9 근거이지
  런타임이 계산할 값이 아니다. 통화 간 정규화는 riskcalc의 FX staleness 경계가
  있는 자리이고, Guardian 체인은 오늘도 통화 불일치를 계산하지 않고 거부한다.
