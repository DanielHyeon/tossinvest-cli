# Function Logic Map: `lazyOrders.Orders`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (revision=current, L521–585, 분기 9개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `583772c4` — base에는 이 함수가 없었다. 현재 본문을 diff hunk가 덮으므로 evidence가 요구된다 (revision=current)

한 번의 /orders 새로고침이 쓰는 브로커 예산 **3콜**이 전부 이 함수 안에 있다.

- **넘어가는 것**: `console.OrdersReading` — 미체결/종결/조건주문 세 목록과 각 목록의 truncated 플래그·에러 문자열. 값은 전부 브로커가 보낸 **원문 문자열**이다.
- **넘어가지 않는 것**: `plain`/`conditional` 두 method value조차 밖으로 나가지 않고, 이제는 필드로 남지도 않는다 — 호출마다 공유 클라이언트에서 꺼내 쓰고 버린다. `broker`는 지역 변수로만 존재한다.
- **계좌 해석**: `l.shared.resolve()`가 콘솔 프로세스당 **한 번**만 `verifyBrokerFactory`를 부른다. 포지션 화면도 같은 resolver를 쓰므로 두 화면을 모두 여는 세션의 `/api/v1/accounts`는 1회다(이전에는 seam마다 1회씩 2회였다). 두 번째 `*official.Client`를 여기서 만들면 그 해석이 또 중복되고, 그 호출이 2026-07-26에 429를 세 번 받아 실행 3스텝을 잃게 한 것이다(measurements.md M4).
- **타임아웃**: 이 함수는 자기 데드라인을 세우지 않는다. `ctx`는 요청의 것이고, 브라우저가 떠난 요청은 세 콜 도중에 취소된다. HTTP 레벨 재시도·백오프는 `internal/official` 쪽 정책이다.
- **실패가 화면에 남기는 것**: 세 콜 중 하나가 실패하면 그 목록만 `*Error`로 표시되고 나머지 둘은 살아남는다. 함수 전체 에러는 **읽을 것이 아예 없을 때**(자격증명 없음 B1, raw 읽기 없는 브로커 B2)만 반환한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 요청 context | internal/console 핸들러 | 취소되면 진행 중 콜이 에러로 끝나고 그 목록만 미측정으로 남는다 |
| `l.shared` | non-nil `*consoleBroker` | `runConsole`이 만든 공유 resolver 1개 | 해석 실패는 그대로 반환(B1). 실패는 캐시되지 않는다 |
| 브로커의 raw 읽기 | `consoleOrdersReader` 단언 성공 | `*official.Client` | 실패하면 `%T`를 담은 에러(B2) — 원문 소수 보존이 불가능하다는 설명 |
| `orderGroupOpen`/`orderGroupClosed` | "OPEN" / "CLOSED" | openapi `status` (required) | 잘못 보내면 400이거나 계좌 전 이력 1페이지 — 둘 다 잔여물이 숨는다 |
| `consoleOrdersPageLimit` | 100 (CLOSED에만) | openapi 최대 100 | OPEN에 붙이면 자를 수 없는 목록에 자른다는 표시를 다는 셈 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `l.shared.resolve()` 실패(자격증명 없음 등) | 없음 | `console.OrdersReading{}, err` | `TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient`의 factory 경로 |
| B2 | 브로커가 `consoleOrdersReader`가 아님 | 없음 | `%T`를 담은 에러 | `TestTheConsoleIsHandedOneCapabilityAndNotABroker`(형상) + 컴파일 타임 형 계약 |
| B3/B4 | OPEN 콜 실패 / 성공 | 실패: `out.OpenError`. 성공: `out.OpenTruncated`, `out.Open` | 에러 반환 없음 | `TestTheOrdersSeamCarriesEachListsOutcomeSeparately`, `TestOneRefreshAsksTheOpenGroupAndTheClosedGroupSeparatelyAndTheLiveOneWhole` |
| B5/B6 | CLOSED 콜 실패 / 성공 | 실패: `out.ClosedError`. 성공: `out.ClosedTruncated`, `out.Closed` | 에러 반환 없음 | 동일 2종 (`hasNext`가 `ClosedTruncated`로 살아남는지 포함) |
| B7/B8 | 조건주문 콜 실패 / 성공 | 실패: `out.ConditionalError`. 성공: `out.ConditionalTruncated` + 레코드 변환 | 에러 반환 없음 | `TestTheOrdersSeamCarriesEachListsOutcomeSeparately` (429 주입) |
| B9 | 조건주문 레코드 루프 | `out.Conditional` append (원문 보존) | — | 동일 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `l.shared.resolve` | 콘솔이 공유하는 라이브 클라이언트 — 이 파일이 `official.New`를 직접 부르지 않는 이유 | 첫 호출만 `verifyBrokerFactory`(= `/api/v1/accounts` 1회), 락 안에서 1회. 에러는 그대로 반환 | `TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient`의 소스 절반이 `official.New(` 금지와 `l.shared.resolve()` 경유를 검사 |
| `plain` (= `Client.OrdersRaw`) | OPEN 1콜 + CLOSED 1콜 | 각 콜의 에러는 해당 목록에만 귀속 | `TestOneRefreshAsks...` 가 wire 3건을 순서대로 검사 |
| `conditional` (= `Client.ConditionalOrdersRaw`) | 살아있는 조건주문 = 노출 상한을 채우는 잔여물 | `verifylive.ConditionalStatusOpen`, limit 100 | 동일 테스트 3번째 wire |
| `consoleOrderRecords` | raw 주문을 콘솔 레코드로 옮긴다 | 변환 없음 — 문자열 그대로 | `TestTheOrdersSeamCarriesEachListsOutcomeSeparately`의 null price 검사 |

## State mutations and fallbacks

- **이 타입은 아무것도 변이하지 않는다**. 예전의 `l.plain`/`l.conditional` 캐시는 사라졌고, 클라이언트 캐시는 `consoleBroker` 한 곳에 있다. 실패는 여전히 캐시되지 않는다 — `tossctl openapi login` 이후 다음 새로고침이 다시 시도한다.
- 세 콜의 결과는 **합산되지 않는다**. 하나라도 실패하면 콘솔은 그 부분을 미측정으로 렌더하고 건수 합계를 거부한다. 세 실패를 하나의 에러로 접으면 답한 부분까지 화면에서 사라진다.
- OPEN에는 limit/cursor를 보내지 않는다. API가 무시하므로 보내는 순간 "자를 수 있는 콜"이라는 표시가 wire에 남고, 미체결 건수의 유일한 주장(짧을 수 없다)이 무너진다.
- **계좌 해석 1회의 범위**: 콘솔 프로세스당 1회이며 화면당이 아니다(`TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce`). 검증 실행만 자기 해석을 따로 하고, 그 이유는 `consoleVerifyStarter`의 주석에 적혀 있다.

## Safety conclusion

- Safe edit boundary: 세 콜의 파라미터와 결과 귀속. 지역 변수 `reads`/`broker`를 필드에 담는 변경은 경계 자체를 무너뜨린다.
- High-risk impact: yes (주문 경로) — 주문 가능 브로커에서 읽기 두 개만 잘라내는 지점이다. `reads`를 버리고 `broker`를 필드에 보관하는 한 번의 편집이면 콘솔이 `PlaceOrder`에 도달한다. 또한 `status` 인자를 잃으면 살아있는 잔여물이 건수에서 사라지고, 그 잔여물을 못 보는 것이 이 화면이 막으려는 실패다.
