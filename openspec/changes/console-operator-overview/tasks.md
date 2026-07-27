# Tasks: console-operator-overview

각 `[T]`는 RED → GREEN → REFACTOR → VERIFY를 거친다. 읽기 전용 화면 하나이므로 실계좌
mutating 단계가 없다 — 검증은 fake seam과 httptest로 한다.

> 선행: proposal-freeze 리뷰 2라운드 반영 완료(review.md). 초안은 `/orders`를 함께 담고 있었고
> 리뷰가 그쪽 데이터의 부재를 확인해 `console-orders-screen`으로 분리했다.
> 파일 표면: `internal/console`·`cmd/tossctl`. `internal/journal`·`internal/official`·
> `internal/config`는 **건드리지 않는다** — 이 화면이 쓰는 값은 전부 기존 `ReadOnly` 메서드로
> 읽힌다. 활성 change `add-candidate-discovery`(`internal/candidate`)와 겹치지 않는다.
> 단 `add-candidate-discovery` 5.5가 같은 라우트 표에 `/signals`를 얹으므로, §1의 가드 확장이
> 먼저 들어가는 편이 그쪽에도 이득이다.

## 1. 가드 먼저 — 새 파일이 기존 불변식을 무디게 하지 않는다

새 화면을 새 파일에 만드는 순간 기존 정적 검사 둘이 그 파일을 보지 못한다. **가드를 먼저
넓히지 않으면 이 change가 만드는 것은 화면이 아니라 사각지대다.**

- [ ] 1.1 [T] `registeredRoutes`를 **패키지 전체 파일**로 확장(D2). RED: `overview.go`에서
      라우트를 등록하면 `TestNoRouteNamesAnAccountMutation`·`TestEveryStateChangingRouteAlso
      GoesThroughTheCSRFGate`·`TestTheDashboardScreensAreReads`가 **전부 침묵으로 통과한다**
      → GREEN. **변이 검증**: `overview.go`에 `/verify/steal`을 등록 → FAIL 확인, 제거 → PASS
      확인. 두 결과를 보고에 남긴다.
- [ ] 1.2 [T] `Handle`도 `HandleFunc`과 같이 인식하고, 인자 수로 호출을 건너뛰지 않는다.
      RED: `mux.Handle("/x", h)`로 등록한 라우트가 표에 없다 → GREEN.
- [ ] 1.3 [T] 라우트 수 하한을 실제 수로 올린다(현행 `>= 13`, 실제 16, 이 change 후 17).
      RED: 하한이 실제보다 낮으면 파싱이 절반에서 멈춰도 카나리아가 죽지 않는다 → GREEN.
- [ ] 1.4 [T] 주입 능력 정적 검사를 **`Options` 필드 단위**로 재작성(D3). 세 구멍을 동시에 닫는다.
      - 파일 하드코딩(`packageFiles(t)["holdings.go"]`) 제거
      - **func 타입 seam**을 검사 범위에 포함 — 현행 주입 seam 일곱 중 다섯이 func 타입이라
        인터페이스만 보는 검사는 그것들을 전혀 보지 못한다
      - allowlist를 **seam별**로 — 하나의 allowlist를 전 인터페이스에 적용하면
        `Handoff{Mint,Consume}`·`AdoptionSettings{Load,Save}`에서 실패한다(후자는 spec이
        요구하는 것이다). 즉 "모든 인터페이스에 하나의 allowlist"는 그 자체로 모순이다
      - allowlist에 없는 **새 `Options` 필드는 그 자체로 실패**
      - 임베디드 인터페이스 거부 유지 + 인라인 `interface{...}` 리터럴도 걷는다
      **변이 검증**: ① 새 파일의 인터페이스 seam에 `PlaceOrder` 추가 → FAIL,
      ② func 타입 `PlaceOrderFunc`를 `Options`에 추가 → FAIL, ③ allowlist 없는 필드 추가 →
      FAIL. 셋 다 제거 후 PASS. 세 결과를 보고에 남긴다.
- [ ] 1.5 [T] 메서드 패턴(`"GET /dashboard"`)을 쓰지 않음을 확인. RED: 쓰면 추출기가 경로를
      `GET /dashboard`로 읽어 **모든 경로 대조 가드가 조용히 어긋난다** → GREEN(평범한 경로).

## 2. 읽기 seam

- [ ] 2.1 [T] `holdingsCache.peek(now)` — **갱신하지 않는 읽기**(D4). RED: `get()`은
      `!hold && TTL 경과`면 무조건 갱신하므로 0콜이 불가능하고, 우회로 `hold=true`를 쓰면
      검증이 돌지 않는데도 `HeldReason`이 "검증 중 — 갱신 보류"로 렌더된다 — **지어낸 사유**
      → GREEN(`never_fetched`를 별도 사유로).
- [ ] 2.2 [T] `holdingsSnapshot.Stale()`의 문서 주석 정정 — "It is only ever true alongside
      Held: an unheld request refreshes instead"는 `peek` 도입으로 거짓이 된다. 주석에 기대는
      템플릿 분기가 있으면 함께 고친다.
- [ ] 2.3 [T] `c.positions(ctx)`를 fetch와 원장 조인으로 분리. RED: 그대로 쓰면 관리/관리 외
      집계가 `holdings.get`을 불러 2.1의 0콜을 깬다 → GREEN.
- [ ] 2.4 [T] Guardian 한도 읽기 seam — `GateLimits()` **한 메서드**, 편입 설정 seam과 **별개**
      (D8). RED: `AdoptionSettings`에 세 번째 메서드를 더하면
      `TestTheSettingsSeamDeclaresExactlyLoadAndSave`가 실패한다 — 초안 spec은 같은 파일에서
      "Load·Save 둘뿐"과 "한도를 보인다"를 동시에 요구하고 있었다 → GREEN(별개 seam).
- [ ] 2.5 [T] seam 미배선 빌드에서 해당 패널만 `seam_unwired`로 렌더되고 나머지는 영향받지
      않는다. RED: nil seam이 패닉 → GREEN.

## 3. 미측정의 자료형

- [ ] 3.1 [T] `(값, 측정됨, 사유)` 자료형(D5). RED: 실패가 0으로 표현되어 "값이 0"과 구분되지
      않는다 → GREEN.
- [ ] 3.2 [T] 사유 **다섯**을 각각 보존 — `verify_suspended`·`broker_read_failed`·
      `journal_unreadable`·`seam_unwired`·`never_fetched`. RED: 하나로 뭉치면 운영자가 기다릴지
      고칠지 배선할지 알 수 없다 → GREEN. 초안은 셋이라고 썼는데 코드는 이미 넷을 모델링하고
      있었다(`holdings.go`의 `Wired`).
- [ ] 3.3 [T] 값이 없는 항목을 **조용히 생략하지 않는다** — 라벨 없이 사라지는 것은 미측정
      표시가 아니라 은폐다(D10, StockOS `addDetail()` 패턴은 가져오지 않는다).

## 4. `/dashboard` 화면

- [ ] 4.1 [T] 라우트 등록 — GET, `session0` 안, CSRF 밖. `/`는 손대지 않는다(D1).
      RED: `/`가 리다이렉트되면 검증 승인 창을 보고 있던 탭이 갈아탄다 → GREEN.
      `TestTheDashboardScreensAreReads`의 `want` 맵에 `/dashboard` 추가.
- [ ] 4.2 [T] **브로커 0콜**. RED: 개요가 캐시를 스스로 갱신해 가장 오래 열려 있는 화면이
      호출을 만든다 → GREEN(`peek`). 콜드 캐시는 0이 아니라 "아직 읽지 않음" +
      **값을 채우는 화면으로 가는 링크**. RED: 링크가 없으면 그 숫자는 영원히 빈다.
- [ ] 4.3 [T] 엔진 패널 — 실행 여부 + 기동 거부 사유. 사유는 **프로세스 로컬**이다
      (`c.engineNote`는 이 콘솔에서 시작/정지를 누른 경우에만 채워진다). 갓 연 화면에 사유가
      없는 것이 정상 — 지속 저장된 사유를 찾아 헤매지 않는다.
- [ ] 4.4 [T] 계좌 패널 — 평가액·평가손익·보유 수(관리/관리 외), **시장별**.
      RED: 시장을 가로질러 더하면 KRW와 USD를 더한 숫자가 나온다(`domain.Position`에는 통화
      필드가 없고 `MarketType`뿐) → GREEN(시장별 행, 가로 합계 없음).
- [ ] 4.5 [T] 오늘 패널 — 실현손익·왕복 건수·승패, **시장별**. `AccountTradeTrips`(Market·
      ClosedAt·비용차감 실현손익), 동결 값만 사용하고 fills 재계산 없음.
      거래가 없는 시장은 0이 아니라 "거래 없음".
- [ ] 4.6 [T] **"오늘"의 경계 = 시장별 현지 자정**(D6). RED: `closed_at`은 UTC이고 UTC 자정은
      **KST 09:00 — KR 장 시작 한 시간 뒤**에 떨어져 그날 아침 체결이 어제로 간다 → GREEN
      (`clock.Market.Location()`). 화면이 **어느 경계를 썼는지 출력**한다 — 경계를 고르는 것과
      고른 것을 감추는 것은 다른 문제다.
- [ ] 4.7 [T] 미체결 패널 — 이 change에서는 **`seam_unwired` 미측정**이다(D7).
      RED: 임시로 0을 넣으면 "미체결 없음"으로 읽히고, 그것이 D5가 금지하는 바로 그것이다
      → GREEN(미측정). seam은 `console-orders-screen`이 배선한다.
- [ ] 4.8 [T] 안전 패널 — 검증 상태 + 잔여물. **두 시장 모두 읽는다**. RED: `c.snapshot()`은
      KR만 읽고 KR/US 증거 기록은 별개 파일이라, 한 시장만 보이는 잔여물 패널이 다른 시장의
      잔여물을 감춘다 — 그것이 이 패널이 존재하는 이유다 → GREEN.
      잔여물이 다음 검증의 노출 상한을 먹는다는 사실을 명시.
- [ ] 4.9 [T] Guardian 한도 — **읽히는 값만**, 소진분은 **한 축만**(D8).
      실현손익 대 `max_daily_loss_amount`만 산출하고 개방 노출·주문 수량·주문 금액은
      **구조적으로 미측정**이라고 말한다. RED: holdings에서 무언가를 계산하면 측정된 한도
      소진분처럼 렌더되고, 있지도 않은 보호를 믿게 만든다 → GREEN.
- [ ] 4.10 [T] 최근 exit 이벤트 N건. `AccountExitEvents`(RO). **체결은 넣지 않는다** —
      `journal.ReadOnly`에 fills 접근자가 없고, 만드는 것은 원장 읽기 표면 변경이라
      `console-orders-screen`의 일이다.
- [ ] 4.11 [T] **패널 단위 격리** — 한 출처가 없어도 나머지가 렌더된다. RED: journal 부재가
      화면 전체를 비운다 → GREEN.
- [ ] 4.12 [T] meta refresh 주기 = `holdingsTTL`. 상수에서 파생시켜 표류를 막는다
      (`positionsPage.RefreshSeconds`가 선례).
- [ ] 4.13 [T] 화면에 POST 폼이 없고 확인 문자열 입력란이 없음을 렌더 결과로 고정
      (D11 — 사용자 지시 2026-07-27).

## 5. 배선과 내비게이션

- [ ] 5.1 [T] `cmd/tossctl`이 `GateLimits` seam을 주입. 주입 없이도 콘솔이 뜬다(2.5).
- [ ] 5.2 [T] 템플릿 등록 — `pages.go`의 `Parse` 체인에 개요 템플릿 추가. 페이지 구조체는
      `Nav`·`Refresh`·`RefreshSeconds` 셋을 다 제공해야 한다(`head`가 셋 다 읽는다;
      `historyPage`가 `RefreshSeconds`를 생략할 수 있는 것은 `Refresh`가 false이기 때문이다).
- [ ] 5.3 [T] 내비게이션 라벨(D1) — `/`는 `검증 콘솔`(`Nav: "verify-console"`),
      `/dashboard`는 `개요`(`Nav: "overview"`). RED: 현행 `head`는 `/`를 `대시보드`·
      `Nav == "dashboard"`로 표시하므로 두 화면이 같은 이름을 갖는다 → GREEN.
- [ ] 5.4 [T] `handleDashboard`의 404 본문 "화면은 … 여섯뿐이다"를 새 목록으로 고친다.
      고치지 않으면 404 문구가 거짓말이 된다. (`/report.json`은 원래 그 목록에 없다.)
- [ ] 5.5 [T] 기존 콘솔 테스트 전부 통과(`go test ./internal/console/...`).

## 6. 게이트

- [ ] 6.1 Function Logic Map + `check_analysis.py` — 기존 함수 내부 로직을 바꾸는 것은
      `registeredRoutes`, `TestTheConsoleBrokerInterfaceDeclaresNothingButReads`,
      `holdingsCache`, `c.positions` 넷이며 앞의 둘은 High-risk 가드이므로 면제 없음.
- [ ] 6.2 PM registry allowlist + fixture 등록
- [ ] 6.3 `make sdd-sync && make sdd-check && make gate CHANGE=console-operator-overview`
- [ ] 6.4 독립 리뷰(별도 컨텍스트) — 특히 §1의 가드 확장이 **실제로 좁아졌는지**를
      변이 결과로 확인
