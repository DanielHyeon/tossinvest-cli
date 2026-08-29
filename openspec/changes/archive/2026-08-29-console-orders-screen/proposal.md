# Change: console-orders-screen

## Why

사용자 요청(2026-07-28): `http://localhost:9000/orders`가 필요하고 **현재 TossOS에는 없다**.

`/positions`는 보유만, `/history`는 이미 끝난 왕복만 답한다. **그 사이 — 낸 주문이 아직 살아
있는가 — 는 어느 화면에도 없다.** 잔여물은 `/verify`의 검증 화면에 나오지만 그것은 검증 run이
있을 때의 이야기이고, 엔진이 낸 주문이나 사람이 앱에서 낸 주문은 거기 없다.

이 change는 `console-operator-overview`에서 분리됐다. 분리 이유는 **데이터가 없기 때문**이다 —
초안은 한 change에 두 화면을 담았고, proposal-freeze 리뷰가 `/orders`의 약속 여섯이 실제로는
뒷받침되지 않으며 셋은 구현자가 숫자를 지어내야 함을 확인했다.

| 초안이 약속한 것 | 실제 |
|---|---|
| 종목명·시장 열 | `adaptOrder`가 채우지 않는다([orders_reads.go:155](../../../internal/official/orders_reads.go#L155)). `currency`는 디코드 후 버려진다 |
| "브로커가 안 준 값은 0이 아니라 —" | `parseDecimal("")→0`([market_reads.go:15](../../../internal/official/market_reads.go#L15)). 부재·해석불가·진짜 0이 콘솔이 보기 전에 하나가 된다 |
| 미체결 건수 | **조건주문이 안 보인다.** 다른 엔드포인트다([conditional_reads.go:48](../../../internal/official/conditional_reads.go#L48)) |
| 발주 주체 조인 | `mutation_attempts`를 읽는 `journal.ReadOnly` 메서드가 없고 `readOnlyTables`에도 없다 |
| 정확한 건수 | `Orders`가 `NextCursor`·`HasNext`를 버린다 — 잘린 첫 페이지가 확신에 찬 짧은 숫자가 된다 |

### 조건주문이 가장 위험하다

`verifylive`의 정리 로직은 **일반 주문과 조건주문을 둘 다 잔여물로 센다** —
[cleanup.go:141](../../../internal/verifylive/cleanup.go#L141)이 양쪽에 대해 "그것이 노출
상한을 채우고 있는 동안에는 아래 단계들이 아무것도 보낼 수 없다"고 말한다. 그리고 M18은
**조건주문이 프로세스 종료를 넘어 존속한다**는 측정이다 — 이 제품에서 조건주문은 예외가 아니라
지속되는 산출물이다.

그러므로 `Client.Orders` 하나만 부르는 `/orders`는 **조건주문 잔여물이 살아서 다음 검증을 막고
있는데 "미체결 0건"을 측정된 값으로 렌더한다.** 이 change가 막으려는 실패를 이 change가
저지르는 것이다.

## What Changes

- **`/orders`(GET, 읽기 전용)** — 미체결 주문과 오늘의 주문 이력. 시각·심볼·매수/매도·상태·
  주문수량·체결수량·주문가·평균체결가·**주문번호**·발주 주체. 조건주문을 **함께** 센다.
- **원문 보존 주문 읽기** — `RawHolding`의 선례를 그대로 따라, 브로커의 decimal 문자열을
  보존하는 주문 읽기를 `internal/official`에 **가산**한다. `Orders`/`adaptOrder`는 손대지
  않는다. `parseDecimal`을 거친 값으로는 "0이 아니라 —"가 원리적으로 불가능하다.
- **`journal.ReadOnly`에 발주 주체 읽기 가산** — `mutation_attempts.broker_order_id`.
  `readOnlyTables`에도 등록해, 테이블이 없으면 질의마다 하나씩 실패하는 대신 **열 때 한 번
  깨끗하게 거부**된다(그 목록이 존재하는 이유가 그것이다).
- **페이지 경계를 숨기지 않는다** — `HasNext`가 참이면 건수는 "N건 이상"이지 N건이 아니다.
- **라우트 가드의 정밀화** — 현행 가드는 경로에 `order`가 들어가면 실패시킨다. `/orders`는
  주문을 내는 라우트가 아니라 **주문 기록을 읽는 라우트**다. 가드를 끄지 않고 **정확 경로
  1건**만 예외로 열되, 그 예외가 실제로 읽기임을 검사가 확인할 수 있게 **읽기 전용 wrapper**를
  도입한다(현행 검사는 HTTP 메서드를 볼 수 없어 "GET"이 "CSRF 보호가 없다"로 퇴화한다).
- **주문 조회 seam 배선** — `console-operator-overview`가 `seam_unwired` 미측정으로 남겨 둔
  미체결 패널이 이 change로 실제 값을 갖는다.

## Impact

- **Specs**: `operator-console` — `콘솔 안전 불변식`·`rate budget 보호` MODIFIED,
  `주문 가시성` ADDED.
- **Code(High-risk)**: `internal/journal`(읽기 전용 접근자 가산 — 원장 표면),
  `internal/official`(원문 보존 주문 읽기 가산 — 계좌 게이트웨이).
  둘 다 **가산이며 기존 서명·반환·동작을 바꾸지 않는다.**
- **Code**: `internal/console`(신규 `orders.go`·`orders_reader.go` + templates + static_test),
  `cmd/tossctl` 배선.
- **선행 change**: `console-operator-overview`(라우트·seam 정적 검사 확장, 미측정 자료형).
  그 확장 없이 이 change의 새 파일은 라우트 표 검사에 보이지 않는다.
- **주문 능력 무변경**: GET이고 CSRF 게이트 밖이며, 콘솔의 상태변경 행위 목록은 늘지 않는다.

## Risks

| 위험 | 완화 |
|---|---|
| `/orders` 예외가 "주문 라우트 부재" 가드를 무디게 한다 | 정확 경로 1건, 접두 일치 아님, 대소문자·후행 슬래시 포함 바이트 일치. 파생 경로가 실패함을 테스트가 직접 확인 |
| "읽기라서 열어준다"가 "보호가 없어서 열어준다"로 퇴화한다 | 읽기 전용 wrapper 도입 — GET/HEAD 외 405, 같은 AST 검사가 인식 |
| `journal.ReadOnly`에 메서드를 더하면 쓰기 표면이 열린다 | 새 메서드도 `mode=ro` 연결만 쓰고, `TestTheReadOnlyHandleHasNoWriteMethods`의 열거를 갱신 |
| 원문 보존 읽기가 기존 `Orders` 호출자를 바꾼다 | 가산 전용. `RawHolding`이 같은 패턴으로 이미 존재한다 |
| 조건주문 조회가 rate budget을 먹는다 | 화면당 2콜/TTL(주문 + 조건주문). 검증 중 갱신 보류 상속. 서버측 폴러 없음 |
| 조건주문 조회만 실패하면 총계가 조용히 낮아진다 | 부분 실패는 **미측정**으로 렌더한다 — "0건"과 합쳐 세지 않는다 |
| 페이지가 잘려 건수가 짧아진다 | `HasNext` 노출. 잘렸으면 숫자 대신 "N건 이상" |
