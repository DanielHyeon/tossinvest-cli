# Review: console-orders-screen

## 라운드 1 — proposal-freeze (2026-07-28)

이 change는 초안 `console-dashboard-and-orders`의 적대적 리뷰 2종에서 **분리되어 태어났다**.
그 리뷰의 판정은 "초안대로는 구현 불가"였고, 불가의 대부분이 `/orders` 쪽이었다.
전체 findings와 처분은 `../console-operator-overview/review.md` 라운드 1에 있다.

### 분리 근거 — 데이터가 없다

| 초안이 약속한 것 | 검증된 사실 |
|---|---|
| 종목명·시장 열 | `adaptOrder`가 채우지 않는다(`orders_reads.go:155`). `currency`는 디코드 후 버려진다 |
| "0이 아니라 —" | `parseDecimal("")→0`(`market_reads.go:15`). 부재·해석불가·진짜 0이 콘솔이 보기 전에 하나가 된다 |
| 미체결 건수 | 조건주문이 다른 엔드포인트에 있다(`conditional_reads.go:48`) |
| 발주 주체 조인 | `mutation_attempts`를 읽는 `ReadOnly` 메서드가 없고 `readOnlyTables`에도 없다 |
| 정확한 건수 | `Orders`가 `NextCursor`·`HasNext`를 버린다 |
| market 필터 | 시장 값 자체가 없으므로 필터가 성립하지 않는다 |

셋(종목명·시장, 소수 부재, 발주 주체)은 구현자가 **숫자나 라벨을 지어내야** 하는 것이고,
그것이 이 change가 막으려는 실패다. 그러므로 `internal/official`과 `internal/journal`을
가산으로 바꾸는 것이 전제가 된다 — 둘 다 High-risk 표면이므로 별도 change로 분리했다.

### 리뷰가 찾은 것 중 이 change가 소유하는 P0

**P0-1. 조건주문 맹점.** 리뷰가 `verifylive/cleanup.go:141`을 인용해, 일반 주문과 조건주문이
**둘 다 노출 상한을 채우는 잔여물**임을 확인했다. 그리고 M18은 조건주문이 프로세스 종료를
넘어 존속한다는 측정이다 — 이 제품에서 조건주문은 예외가 아니라 지속되는 산출물이다.
→ `Client.Orders` 하나만 부르는 화면은 **조건주문 잔여물이 다음 검증을 막고 있는데 "미체결
0건"을 측정된 값으로 렌더한다.** 화면당 2콜로 계약을 고치고, **부분 실패는 합산 금지**를
spec 문장으로 넣었다. 확신에 찬 0은 도달 불가능해야 한다.

**P0-2. "GET"이 검사 불가능하다.** 초안 spec은 "예외 경로가 CSRF 게이트 밖 GET임을 같은 검사가
확인한다"고 썼다. `route`가 나르는 사실은 `{Path, Session, CSRFGated}`뿐이고 패키지 전체에
메서드 검사는 `mutating()` 안의 한 줄뿐이다(`console.go:533`). **"GET"은 "CSRF 보호가 없다"로
퇴화하고, 예외가 보호되지 않았다는 이유로 부여된다.** 그 상태에서 `POST /orders`는 세션
쿠키만으로 CSRF 없이 통과한다.
→ 두 리뷰가 서로 다른 해법을 제시했다. 리뷰 1은 Go 1.22+ 메서드 패턴(`"GET /orders"`),
리뷰 2는 그것이 추출기의 경로 대조를 어긋나게 한다는 반대 근거를 들었다(추출기가 리터럴을
그대로 경로로 읽으므로 경로가 `GET /orders`가 된다). **둘 다 맞다** — 그래서 리뷰 1의 대안인
`reading(next)` wrapper를 채택했다. 추출기를 바꾸지 않고 같은 사실을 주고, 405를 런타임
성질로 만든다.

**P0-3. `OrdersReader{Orders}`가 금지 동사 그 자체다.** 메서드 이름 금지 목록에 `"order"`와
`"conditional"`이 있다(`static_test.go:530`). 검사를 패키지 전체로 넓히면 **이 change가
추가하는 seam에서 실패한다.** 자연스러운 해제는 목록에서 `"order"`를 지우는 것이고 그것이
정확히 D3이 경고한 실패다.
→ 선행 change의 **seam별 allowlist**에 `OrdersReader→{Orders}`를 등록하는 방식으로 푼다.
금지 목록은 손대지 않는다.

**P0-4. 원장 접근자 부재를 초안이 몰랐고, 인용도 틀렸다.** 초안 design은 테이블을 `attempts`라
부르고 `internal/journal/dispatch.go`를 인용했다. 테이블은 `mutation_attempts`이고 DDL은
`schema.go`에 있으며 `dispatch.go`에는 관련 내용이 0건이다. 더 중요한 것은 `readOnlyTables`에
그 테이블이 없다는 것 — 등록하지 않으면 **열기는 성공하고 질의만 하나씩 실패하며 0행을
돌려주고, 0행은 "전부 수동 주문"으로 읽힌다.** 그 목록이 존재하는 이유가 정확히 그것이다.
→ 접근자 가산 + `readOnlyTables` 등록 + `TestTheReadOnlyHandleHasNoWriteMethods` 열거 갱신.

### P1 — 반영한 것

- **페이지네이션 유실**(`orders_reads.go:106`) — `HasNext`가 참이면 건수는 "N건 이상"이다.
- **정확 일치의 범위** — 바이트 일치여야 한다. `ToLower`면 `/Orders`, `TrimSuffix`면 `/orders/`가
  통과하고 **후행 슬래시 패턴은 서브트리 패턴이라 하위 경로를 그 핸들러로 라우팅한다.**
- **두 판정 루프** — `actVerbs`가 `accountVerbs`를 포함하므로 `/orders`가 양쪽에 걸린다.
  `consoleStateChanging`에 넣어 조용히 만드는 것은 틀린 수리다(CSRF 게이트가 요구된다).
- **테스트 가능성** — 판정부를 순수 함수로 뽑지 않으면 "예외 경로가 allowlist에 없다"는
  **아무것도 재지 않는 테스트**가 된다. 디스크의 소스를 파싱하는 검사에 가짜 라우트를
  등록할 수 없기 때문이다.
- **필터가 캐시를 쪼갠다** — 브로커 파라미터로 넘기면 TTL당 2콜이 4콜이 된다. in-process로.
- **필터 + 미측정** — `0/—건`은 "0건이 일치"로 읽힌다. 미측정이면 필터를 비활성으로.
- **`cmd/tossctl` 경로가 세 갈래다** — `verifylive.Broker`에는 `Orders`가 없다. 두 번째
  `*official.Client`를 만들면 계정 seq 해석이 중복되고 429 위험이 는다. task에 명시.
- **`SubmittedAt` vs `OrderDate`** — 전자가 실제 순간(`*time.Time`), 후자는 앞 10자.
  시각 열은 전자를 쓰고 파싱 실패 처리를 명시.

### 설계 판단 — 리뷰 이후 직접 확인한 것

리뷰 2가 `internal/brokerstate.ParseOfficialOrder`를 언급했다. 직접 확인한 결과:

- `internal/brokerstate`의 모듈 내부 의존성은 **`internal/domain` 하나뿐**이라 콘솔이
  import할 수 있고 금지 목록에도 없다.
- 이미 `State`/`FailClosed`/`Reason`/`Detail`로 **사유 코드가 붙은 UNKNOWN**을 주문에 대해
  모델링하고 있다 — 이 change가 원하는 규율이 그쪽에 이미 있다.
- 그러나 그쪽 `parseDecimal(*string)`도 `nil`과 `""`를 **둘 다 `0`**으로 돌려준다.
  **소수의 부재는 거기서도 무너진다.**

→ 상태 판정은 `brokerstate`에서 가져올 수 있지만 소수의 부재는 원문 보존 읽기가 필요하다.
D1에 그대로 적었다. 리뷰의 제안을 절반만 채택한 것이고, 절반인 이유를 근거와 함께 남긴다.

### 확인 마찰

**없음.** 화면은 GET뿐이고 폼이 없다. 취소가 필요하면 그것은 콘솔이 아니라 `tossctl`의 일이며
그 경계는 `콘솔 안전 불변식`이 소유한다. 2026-07-27 사용자 지시 준수.

## 착수 조건

1. `console-operator-overview`의 §1(라우트·주입 능력 정적 검사 확장)이 **먼저 들어가야 한다.**
   그 확장 없이 이 change의 새 파일에 등록된 라우트는 라우트 표 검사에 보이지 않는다.
2. 라운드 1의 P0 4건·P1 8건을 proposal·design·tasks·spec에 반영 완료.
   `openspec validate console-orders-screen --strict` 통과.
