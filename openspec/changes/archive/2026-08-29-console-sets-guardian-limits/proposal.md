# Proposal: console-sets-guardian-limits

## Why

개요 화면의 Guardian 한도 다섯 칸이 전부 **미설정**이다.

그것은 화면의 결함이 아니라 정직한 보고다. `config.DefaultFile()`이 `Engine`을
zero value로 두고, `mergeEngine`은 파일에 적힌 것만 옮긴다. 그래서 파일에 숫자를
손으로 적기 전까지 다섯 칸은 0이고, 0은 이 코드베이스에서 "무제한"이 아니라
"아무도 정하지 않았다"이며 게이트 ON 기동을 거부하는 조건이다.

그런데 그 다섯 숫자는 **이미 승인돼 있다.** `risk-management` 스펙의
`정책 수치의 provenance`:

> 모든 한도·정책 수치는 코드에 출처(StockOS 파일·검증 상태)와 함께 기록되어야
> 하며(SHALL), 사용자 미확정 시 보수 기본값 전체 집합을 사용한다(SHALL — 인터록
> 5필드를 전부 충족): 주문당 notional 1,000,000 KRW·주문당 수량 100주·총 노출
> 10,000,000 KRW·일일 손실 100,000 KRW·일일 손실 자본비 1%·통화 KRW.

그리고 그 네 개의 금액·비율은 StockOS `packages/trading/stockos_trading/
risk_profiles.py`의 `_KR_SMALL_LIVE`와 글자 그대로 같다(`max_order_krw=1_000_000`,
`max_open_exposure_krw=10_000_000`, `max_daily_loss_krw=100_000`,
`max_daily_loss_pct=0.01`). 따라갈 정책은 인용까지 끝나 있고 구현만 없다.

두 번째 공백은 **바꿀 방법**이다. 오늘 이 다섯 숫자를 바꾸는 유일한 경로는
`config.json`을 손으로 여는 것이다. 그 파일은 편입 설정이 `LoadRawEngineAdoption`
+ surgical write로 이미 떠난 자리이고, 같은 이유가 여기에도 그대로 적용된다 —
사람이 손으로 하는 read-modify-write는 언급하지 않은 값을 조용히 떨군다. 다만
한도에서 떨어진 값은 목록 한 줄이 아니라 **인터록 5필드 중 하나**이고, 하나가
0이면 게이트는 기동을 거부한다. 즉 오타의 결과가 "일부만 무제한"이 아니라
"엔진이 안 뜬다"인 것은 다행이지만, 그 진단을 사람이 파일에서 찾아야 한다.

## What Changes

1. **보수 기본값 집합을 코드에 등재하고 클릭 한 번으로 적용 가능하게 한다** —
   `risk-management`의 승인된 다섯 숫자를 티어 레지스트리에 출처와 함께 적고,
   한도 화면의 **권장 프리셋**으로 제시한다. 클릭이 그 값을 파일에 명시적으로
   기록한다. 암묵 기본값은 두지 않는다(design D1) — 화면이 보여주는 숫자와 엔진이
   보는 숫자가 갈라지기 때문이고, 이 저장소는 이미 손절폭 기본값에서 같은 판단을
   내려 두었다("the engine still never runs on an implicit number").

2. **StockOS 티어 레지스트리를 이식한다** — `risk_profiles.py`의
   `(market, profile) → RiskProfile` 표를 TossOS의 다섯 필드로 옮긴다. KRW 두 티어,
   USD 두 티어. `paper_demo`는 이식하지 않는다(KIS 데모 계좌 전용 시드).

3. **상한 백스톱을 이식한다** — StockOS ADR §2.6의
   `real_safety_gate_ceiling_violations`. 저장하려는 값이 그 통화에 등록된 가장
   느슨한 티어를 넘으면 거부한다. **콘솔은 등록된 티어 위로 올릴 수 없다.**

4. **콘솔에 편집 표면을 연다** — 프리셋 적용은 클릭 한 번(사용자 결정 2026-07-30),
   개별 값 기입은 고급 접힘 안. 대상은 다섯 한도와 한도 통화뿐이다.

5. **`automation_gate.enabled`와 kill switch는 콘솔 밖에 그대로 둔다** — 그리고
   그것을 규율이 아니라 **구조**로 만든다. 새 writer는 여섯 개 키만 splice하고
   `enabled` 바이트에 손대지 않는다. 콘솔에서 게이트를 켜는 경로는 존재하지
   않는다.

## Impact

- 스펙: `operator-console` 개정 — `Guardian 한도는 콘솔에서 편집할 수 없다
  (SHALL NOT)`를 `한도는 편집할 수 있고 게이트 ON·kill switch는 여전히 불가`로.
  `risk-management`는 개정하지 않는다.
- 코드: `internal/config`(티어 레지스트리·상한·한도 전용 writer),
  `internal/console`(seam·핸들러·한도 화면), `cmd/tossctl`(seam 배선·audit).
  파싱(`mergeEngine`)·엔진 배선·인터록·개요 패널은 **건드리지 않는다**.
- 위험도: **High-risk** — Guardian 한도는 §0.5가 열거한 High-risk 경로다.

## 범위 밖 (명시)

`risk-management`의 "사용자 미확정 시 보수 기본값 전체 집합을 **사용한다**"는
암묵 기본값을 명령하며, 그것을 지키려면 화면뿐 아니라 `runtime_wiring`·인터록까지
같은 주입을 배선해야 한다(design D1·§0.1). 그것은 게이트 ON 기동 조건을 바꾸는
일이므로 별도 change로 분리한다. 이 change는 그 요구사항의 **전반부**(수치를
출처와 함께 코드에 등재)를 구현하고, 후반부는 미구현으로 남긴다.
