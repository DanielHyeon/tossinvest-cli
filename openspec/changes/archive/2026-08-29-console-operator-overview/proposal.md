# Change: console-operator-overview

## Why

사용자 요청(2026-07-28): `http://localhost:9000/dashboard`가 필요하고 **현재 TossOS에는 없다**.
참조는 StockOS의 같은 이름 화면이다.

지금 콘솔의 `/`는 대시보드라는 이름을 달고 있지만 보여주는 것은 **검증 콘솔**이다 —
soak·attestation·바이너리 판본·검증 run·엔진 기동/정지. 계좌가 지금 어떤 상태인지, 오늘
무엇이 일어났는지는 어느 화면에도 모여 있지 않다. `/positions`는 보유만, `/history`는 이미
끝난 것만 답한다. 운영자는 세 화면을 오가며 머릿속에서 합쳐야 하고, 그 합침은 매번 다시 한다.

### 정정 — 초안의 동기는 기록과 달랐다 (2026-07-28 리뷰)

초안은 "2026-07-27 승인 창 셋이 소진됐고 그중 둘은 **잔여물을 화면에서 볼 수 없었기
때문**"이라고 썼다. 기록을 다시 읽으면 그렇지 않다.

| 측정 | 실제 원인 |
|---|---|
| M11 | [이어하기] 기본값이 무동작이었다 → change `console-click-approval` |
| M18 | 조건주문이 프로세스 종료를 넘어 존속 — 창 소진과 무관 |
| M22 | 잔여물 교착. 그러나 잔여물은 run 종료 메시지와 상한 오류에 **이름이 찍혀 있었다** |
| M23 | 세 창 소진. 만료 문구 + 시작 컨트롤이 렌더되지 않은 것 |

그리고 잔여물은 이미 `/verify`에 렌더된다 —
[templates.go:413](../../../internal/console/templates.go#L413)이 "이전 실행이 남긴 객체 N건 …
그것이 노출 상한을 채우고 있는 동안에는 아래 단계들이 아무것도 보낼 수 없다"를 출력한다.

**동기는 유지되지만 근거는 바뀐다.** 이 화면이 필요한 이유는 "볼 수 없던 것을 보이게 한다"가
아니라 **"흩어진 것을 한 자리에 모은다"**다. 근거를 정확히 적어 두지 않으면 다음 사람이
없는 문제를 고치는 코드를 쓴다.

### 이 change가 `/orders`를 포함하지 않는 이유

초안은 `/dashboard`와 `/orders`를 한 change로 묶었다. 리뷰가 `/orders`의 데이터가 **존재하지
않는다**는 것을 확인했다 — `adaptOrder`는 종목명·시장을 채우지 않고
([orders_reads.go:155](../../../internal/official/orders_reads.go#L155)), `parseDecimal`이
부재·해석불가·진짜 0을 콘솔이 보기 전에 하나로 뭉개며
([market_reads.go:15](../../../internal/official/market_reads.go#L15)), 조건주문은 다른
엔드포인트에 있고, 발주 주체 조인에 필요한 `mutation_attempts`는 `journal.ReadOnly`의 어떤
메서드로도 읽히지 않는다.

즉 `/orders`는 `internal/official`과 `internal/journal` — 둘 다 **High-risk 표면** — 을 바꿔야
한다. 위험 등급이 다른 작업을 한 change에 묶으면 완전히 뒷받침되는 화면이 원장 읽기 표면
변경 뒤에서 기다린다. `/orders`는 별도 change `console-orders-screen`이다.

## What Changes

- **`/dashboard`(GET, 읽기 전용)** — 여섯 패널: 엔진 실행 여부와 기동 거부 사유, 계좌
  평가액·평가손익·보유 종목 수(관리/관리 외, **시장별**), 오늘 실현손익·왕복 건수·승패
  (**시장별**), 살아 있는 주문 건수, 검증 상태·잔여물·Guardian 한도, 최근 exit 이벤트.
- **`/`는 그대로 둔다** — 검증 승인 창을 보고 있는 탭이 다른 화면으로 갈아타서는 안 된다.
- **"못 읽었다"와 "없다"의 구분** — 값을 얻지 못한 패널은 0이 아니라 미측정과 **사유**를
  말한다. 사유는 다섯이고 각각 다른 대응을 요구한다.
- **미체결 패널은 지금 미측정이다** — 주문 조회 seam은 `console-orders-screen`이 배선한다.
  자리를 비워 두는 것이 아니라 **미측정 사유 `주문 조회 미배선`으로 렌더한다.** 이 change의
  규칙을 이 change 자신에게 먼저 적용하는 자리다.
- **라우트 정적 가드의 확장** — 현행 `registeredRoutes`는 `console.go` **한 파일만** 읽는다
  ([static_test.go:78](../../../internal/console/static_test.go#L78)). 새 파일에서 라우트를
  등록하면 라우트 표 검사 전체가 그 라우트를 보지 못한다. 패키지 전체를 걷도록 넓힌다.
- **주입 능력 정적 가드의 확장** — 현행 브로커 인터페이스 검사는 `holdings.go` 한 파일만 보고,
  주입 seam 일곱 중 **다섯은 인터페이스가 아니라 func 타입**이라 검사 대상이 아니다
  (`StartVerify`·`StartEngine`·`StopEngine`·`Relaunch`·`RestartSoak`). 검사 단위를 "인터페이스"가
  아니라 **"`Options`가 받는 능력"**으로 바꾼다.

## Impact

- **Specs**: `operator-console` — `콘솔 안전 불변식`·`read-only 불변식` MODIFIED,
  `운영 개요 가시성` ADDED.
- **Code**: `internal/console`(신규 `overview.go` + templates + static_test 확장),
  **Guardian 한도 전용 읽기 seam 1개 추가**, `cmd/tossctl` 배선.
- **Code(무변경)**: `internal/journal`·`internal/official`·`internal/config`. 이 change가 쓰는
  원장 값은 전부 기존 `ReadOnly` 메서드로 읽힌다 — `AccountTradeTrips`(시장·청산시각·비용차감
  실현손익), `AccountExitEvents`, `LivePositionExits`.
- **주문 능력 무변경**: 화면은 GET이고 CSRF 게이트 밖이며, 콘솔의 상태변경 행위 목록은
  **한 건도 늘지 않는다**.

## Risks

| 위험 | 완화 |
|---|---|
| 가드를 넓히면 기존 seam(`Handoff{Mint,Consume}`·`AdoptionSettings{Load,Save}`)에서 실패한다 | allowlist를 **seam별**로 두고 `Options` 필드에서 유도 — 타입 이름·파일 이름에 고정하지 않는다 |
| Guardian 한도 seam이 설정 seam의 "Load·Save 두 메서드뿐" 규칙을 깬다 | 세 번째 메서드가 아니라 **별개 seam**. spec 문장을 두 seam으로 고쳐 쓴다 |
| "오늘"의 경계가 시장마다 다르다 | 시장별 현지 자정(`clock.Market.Location()`), 화면이 **어느 경계를 썼는지 출력** |
| 계좌 합계가 KRW와 USD를 더한다 | 계좌 패널도 **시장별 분리**. 통화를 섞은 한 숫자를 만들지 않는다 |
| 브로커 0콜 규칙이 계좌 패널을 영원히 빈 채로 둔다 | 캐시 **peek** 접근자 신설(갱신하지 않는 읽기) + 값을 채우는 화면으로 가는 링크를 spec이 요구 |
| 넓힌 가드가 실제로는 아무것도 못 잡는다 | 변이 검증 필수 — 새 파일에 라우트·mutation seam을 넣어 FAIL 확인, 제거 후 PASS 확인 |
| Guardian "오늘 소진분"은 어디에도 기록이 없다 | 오늘 실현손익 대 `max_daily_loss_amount` 한 축만 산출. 나머지 축은 **구조적으로 미측정**이라고 spec이 명시 |
