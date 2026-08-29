# Review: console-operator-overview

## 라운드 1 — proposal-freeze 적대적 리뷰 (2026-07-28)

초안 change 이름은 `console-dashboard-and-orders`였고 `/dashboard`와 `/orders`를 한 change에
담고 있었다. 별도 컨텍스트 2종(안전 불변식 렌즈 / 구현가능성 렌즈)으로 검토했다.

**판정: 초안대로는 구현 불가.** 두 리뷰가 독립적으로 같은 결론에 도달했다 — 스펙이 고정한
값 여섯이 실제로는 뒷받침되는 데이터가 없고, 그중 셋은 구현자가 **숫자를 지어내야** 한다.

### 구조 결정 — change를 둘로 쪼갠다

`/orders`가 요구하는 데이터가 존재하지 않는다는 것이 확인됐다.

| 필요한 것 | 실제 |
|---|---|
| 종목명·시장 | `adaptOrder`가 채우지 않는다(`orders_reads.go:155`). `currency`는 디코드 후 **버려진다** |
| "0이 아니라 —" | `parseDecimal("")→0`(`market_reads.go:15`). 부재·해석불가·진짜 0이 콘솔이 보기 전에 하나가 된다 |
| 조건주문 | 다른 엔드포인트(`conditional_reads.go`). `Client.Orders`는 `/api/v1/orders`만 본다 |
| 발주 주체 조인 | `mutation_attempts`를 읽는 `ReadOnly` 메서드가 없고 `readOnlyTables`에도 없다 |
| 페이지네이션 | `NextCursor`·`HasNext`가 버려진다 |

즉 `/orders`는 `internal/official`과 `internal/journal` — 둘 다 **High-risk 표면** — 을
바꿔야 한다. 위험 등급이 다른 작업을 묶으면 완전히 뒷받침되는 화면이 원장 읽기 표면 변경
뒤에서 기다린다. → `console-operator-overview`(이 change)와 `console-orders-screen`으로 분리.

### P0 — 이 change에 남은 것

**P0-1. 초안의 동기가 기록과 달랐다.** "2026-07-27 승인 창 셋 소진, 둘은 잔여물을 화면에서
볼 수 없었기 때문"이라고 썼다. 기록은 M11=이어하기 무동작, M18=조건주문 존속, M22=잔여물
교착이지만 **잔여물에는 이름이 찍혀 있었고**, M23=만료 문구+시작 컨트롤 미렌더다. 그리고
잔여물은 이미 `/verify`에 렌더된다(`templates.go:413`).
→ **없는 문제를 고치는 코드를 쓸 뻔했다.** proposal의 Why를 정정하고 근거를 "볼 수 없던 것을
보이게 한다"에서 **"흩어진 것을 한 자리에 모은다"**로 바꿨다.

**P0-2. 라우트 예외의 "GET" 전제가 검사 불가능했다.** 초안 spec은 "예외 경로가 CSRF 게이트 밖
GET임을 같은 검사가 확인한다"고 썼다. `registeredRoutes`가 만드는 사실은 `{Path, Session,
CSRFGated}` 셋뿐이고 **HTTP 메서드는 볼 수 없다** — 패키지 전체에 메서드 검사는 `mutating()`
안의 한 줄뿐이다(`console.go:533`). 즉 "GET"은 "`mutating`으로 감싸지 않았다" = **"CSRF 보호가
없다"**로 퇴화한다. 예외가 **보호되지 않았다는 이유로** 부여된다.
→ `/orders`가 이 change에서 빠지면서 예외 자체가 사라졌다. `console-orders-screen`이
추출기 개선과 함께 다룬다. 메서드 패턴(`"GET /x"`)은 추출기가 리터럴을 그대로 경로로 읽어
**모든 경로 대조를 어긋나게** 하므로 이 change에서는 쓰지 않기로 명시(task 1.5).

**P0-3. `OrdersReader{Orders}`가 가드의 금지 동사 그 자체였다.** 메서드 이름 금지 목록에
`"order"`와 `"conditional"`이 있다(`static_test.go:530`). 검사를 패키지 전체로 넓히면
**이 change가 추가하는 바로 그 seam에서 실패한다.** 자연스러운 해제는 목록에서 `"order"`를
지우는 것이고 그것이 정확히 D2가 경고한 실패다.
→ D3을 재작성. 검사 단위를 "인터페이스"에서 **"`Options`가 받는 능력"**으로 바꿔 seam별
allowlist를 쓴다.

**P0-4. Guardian 한도가 spec 자기모순이었다.** 같은 파일이 "설정 seam은 Load·Save 두 메서드뿐
(정적 검사)"과 "대시보드가 Guardian 한도를 보인다"를 동시에 요구했다. 콘솔의 유일한 config
seam은 편입 블록만 돌려주고 세 번째 메서드는 `TestTheSettingsSeamDeclaresExactlyLoadAndSave`가
막는다.
→ D8: 세 번째 메서드가 아니라 **별개 seam**. 좁은 인터페이스 규율은 seam을 늘려서 지킨다.

**P0-5. "오늘 소진분"은 어디에도 기록이 없다.** `OpenExposure`·`DailyRealizedLoss`는 엔진이
in-process로 계산하는 값이고(`risk/input.go:159`) 원장 23개 테이블 어디에도 없다. 구현자는
holdings에서 무언가를 계산할 것이고 그것은 **측정된 한도 소진분처럼 렌더된다.**
→ 한 축(실현손익 대 `max_daily_loss_amount`)만 산출, 나머지는 **구조적으로 미측정**.

**P0-6. "오늘"의 시간대가 정해져 있지 않았다.** `closed_at`은 UTC이고 UTC 자정은
**KST 09:00 — KR 장 시작 한 시간 뒤**에 떨어진다. 그날 아침 체결이 어제로 간다. ET로는
US 세션을 반으로 자른다. 저장소는 두 시장의 위치를 이미 정의한다(`clock/market.go:44`).
→ D6: 시장별 현지 자정 + **어느 경계를 썼는지 화면이 출력**.

### P1 — 반영한 것

**P1-1. 라우트 추출기가 `console.go` 한 파일만 읽는다**(`static_test.go:78`). 새 파일에서
등록한 라우트는 라우트 표 검사 **전부에서 침묵으로 통과한다**. D3이 인터페이스 검사에서
지적하는 것과 **같은 결함이 라우트 쪽에 있었고 초안은 그쪽만 고치고 있었다.**
`add-candidate-discovery` 5.5의 `/signals`가 곧 두 번째 사례가 된다. → task 1.1~1.3.

**P1-2. func 타입 seam은 인터페이스 검사에 잡히지 않는다.** 주입 seam 일곱 중 **다섯이 func
타입**이다(`StartVerify`·`StartEngine`·`StopEngine`·`Relaunch`·`RestartSoak`).
`type PlaceOrderFunc func(...)`를 새 파일에 두면 인터페이스도 없고 금지 import도 없어 통과한다.
→ D3의 `Options` 필드 단위 검사가 이것을 처음으로 범위에 넣는다.

**P1-3. 하나의 allowlist를 전 인터페이스에 적용하면 기존 seam에서 실패한다** —
`Handoff{Mint,Consume}`, `AdoptionSettings{Load,Save}`. 후자는 spec이 명시적으로 요구하는
것이다. 초안 문장은 한 문단 안에서 자기모순이었다. → seam별 allowlist.

**P1-4. 초안 D9의 기계적 보증 주장이 틀렸다.** "경로 allowlist가 주문 티켓 이식을 막는다"고
썼는데, 티켓은 금지 동사가 없는 경로(`/ticket`)로 POST할 수 있고 그 경로는 CSRF 게이트에
걸릴 때만 실패한다 — 원하는 것의 반대다. → 기계적 보증은 **`Options` 능력 열거**가 한다.

**P1-5. 0콜 규칙이 헤드라인 숫자를 영원히 비운다.** 캐시는 요청 시 lazy로 채워지므로,
`/positions`를 열지 않으면 계좌 패널은 계속 "아직 읽지 않음"이다. 그리고 갱신 없이 읽는
접근자가 **없다** — `get()`은 `!hold && TTL 경과`면 무조건 갱신한다(`holdings.go:165`).
우회로 `hold=true`를 쓰면 검증이 돌지 않는데 **"검증 중 — 갱신 보류"라는 지어낸 사유**가
렌더된다. → `peek` 신설 + **링크를 spec이 요구**(산문에만 두면 링크 없는 구현도 계약을 만족).
`Stale()`의 문서 주석은 `peek` 도입으로 거짓이 되므로 같은 change에서 고친다.

**P1-6. 미측정 사유가 셋이 아니라 다섯이었다.** 코드는 이미 넷을 모델링하고 있었고
(`holdings.go:83`의 `Wired`) D4가 다섯째("아직 읽지 않음")를 만든다. → `seam_unwired`,
`never_fetched` 추가.

**P1-7. MODIFIED가 `console-adoption-controls`의 조항 9개를 떨어뜨렸다.** 헤더는 "그대로
포함한다"고 썼지만 축자가 아니었고, MODIFIED는 블록 전체를 갈아치우므로 **나중에 아카이브되는
쪽이 이긴다.** 둘은 HEAD 대비 회귀였다 — "엔진 프로세스가 주문 능력을 갖는지는 §0.7…",
"— 콘솔이 인터록을 우회할 수 없다". → 두 요구사항 본문을 `console-adoption-controls`의
delta에서 **바이트 그대로** 가져오고 이 change의 문장만 뒤에 붙이도록 재작성. 보존 확인:
안전 불변식 1010자, read-only 1127자 원문 일치.

**P1-8. 계좌 합계가 통화를 섞는다.** `domain.Position`에는 통화 필드가 없고 `MarketType`뿐이며
공식 클라이언트에는 계좌 요약 읽기가 없다. → 계좌 패널도 시장별 분리. **합계를 만들지 않는
것이 합계를 잘못 만드는 것보다 낫다.**

**P1-9. 안전 패널이 KR만 읽는다.** `c.snapshot()`은 KR 전용이고 KR/US 증거 기록은 별개
파일이다. 한 시장만 보이는 잔여물 패널은 다른 시장의 잔여물을 감추는데, **그것이 이 패널이
존재하는 이유 그 자체다.** → 두 시장 모두 읽는다.

### P2 — 반영한 것

- 라우트 수 하한이 `>= 13`인데 실제 16 — 카나리아가 느슨하다. → 실제 수를 따라가게.
- `head`가 `/`를 `대시보드`·`Nav == "dashboard"`로 표시 — 두 화면이 같은 이름을 갖는다.
  → `/` = `검증 콘솔`, `/dashboard` = `개요`.
- 템플릿은 디렉터리가 아니라 `Parse` 체인이고, `Refresh`가 true면 `RefreshSeconds`가 필수다.
- `TestTheDashboardScreensAreReads`의 `want` 맵 갱신이 task에 없었다.
- 엔진 기동 거부 사유는 **프로세스 로컬**이다(`c.engineNote`) — 갓 연 화면에 없는 것이 정상.
- 체결은 `journal.ReadOnly`로 읽을 수 없다 → 최근 활동에서 제외, orders change로 이관.

### 확인 마찰

**없음.** 두 리뷰 모두 확인 문자열·2단계 클릭·추가 승인이 없음을 확인했다. 화면은 GET뿐이고
폼이 없다. task 4.13은 마찰이 아니라 **마찰에 대한 가드**다. 2026-07-27 사용자 지시 준수.

### 리뷰가 유지하라고 한 것

D1(`/` 그대로 — 탭 갈아타기 근거가 맞다), D5/D10의 미측정 자료형 규칙, 패널 단위 격리,
폼 없음 입장, StockOS SPA 이식 거부, TTL 상수에서 refresh 주기 파생
(`portfolio_pages.go:41`이 선례), 시장별 분리(`AccountTradeTrips`가 이미 `Market`을 준다).

## 착수 조건

라운드 1의 P0 6건·P1 9건을 proposal·design·tasks·spec에 반영 완료.
`openspec validate console-operator-overview --strict` 통과. 구현 착수 가능.
