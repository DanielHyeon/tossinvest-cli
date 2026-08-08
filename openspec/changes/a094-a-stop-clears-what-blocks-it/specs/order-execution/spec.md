# order-execution — a094 delta

> **MODIFIED는 요구를 통째로 치환한다.** 따라서 이 delta는 기존 「IN_DOUBT 해소」 요구의
> **본문 전문과 시나리오 6개를 그대로 재현**하고, 그 뒤에 a094의 조항과 시나리오를
> 덧붙인다. 1판은 기존 본문을 「종전 조항 (변경 없음)」이라는 산문 참조로 대체했고,
> 그러면 재생 계약·부재 판정 N회·해소 불능 규칙과 시나리오 6개가 정본에서 **사라진다**
> (1라운드 차단 1). `openspec validate --strict`는 구조만 보고 보존을 보지 않으므로
> 이 손실을 잡지 못한다 — 선례는 `archive/2026-07-26-extend-execution-contract`이며
> 그 delta는 전문을 재현했다.

## MODIFIED Requirements

### Requirement: IN_DOUBT 해소
IN_DOUBT 해소의 목적은 **정체 회수**다 — 주문이 접수됐는지, 접수됐다면 어떤 식별자인지 확정하는 것이며, 정체를 모르는 채 다시 주문을 내는 것은 금지된다(SHALL NOT).

공식 API의 주문 생성은 클라이언트 제공 멱등키를 지원한다(openapi `clientOrderId`: "동일 값으로 재요청 시 이전 주문 결과를 그대로 재반환합니다", 유효 10분). 해소 순서(SHALL):

1. **멱등 재생**: RECORDED 단계에 영속된 동일 키·동일 wire body의 재요청 응답에서 식별자를 회수한다. 다음은 **진입점 자신의 의무**다(SHALL — 호출자 전제 금지): 대상 attempt가 IN_DOUBT 상태인지 확인, 능력 attestation 플래그 확인 `[미측정 — 2b 전 비활성]`, **재생 1회마다** `elapsed(전송 시작) < TTL − margin`(margin 기본 60초, 설정 주입 — 경과 근거가 로컬 시계이므로 마진 없는 경계 사용 금지(SHALL NOT)) 재검사, 회수 상한(기본 2회)·최소 간격 준수, 저장된 wire body 외의 본문을 구성·전송할 수 없는 형태(SHALL — 진입점 입력은 attempt 식별자뿐). 재생은 새 attempt를 만들지 않고 nonce를 소비하지 않으며 해소 기록(횟수·시각)으로 남는다(SHALL).

   재생 응답 분류에 dispatch 분류기를 사용해서는 안 된다(SHALL NOT — 재생의 422는 의미가 다르다): 2xx+식별자 → 회수·CONFIRMED; `409 request-in-progress`(openapi) → 원 요청 처리 중, 대기 후 재시도(상한 미소비); `422 idempotency-key-conflict`(openapi) → 원 주문에 대해 아무것도 증명하지 않으므로 FAILED_CONFIRMED 금지(SHALL NOT), 프로그램 오류로 UNRESOLVED + critical 알림; 응답 유실 → 기록 후 상한 도달 시 조회 폴백.
2. **조회 대조 (폴백)**: 창 초과, 미검증, 키 충돌, 멱등키를 받지 않는 mutation(취소·정정)은 journal fingerprint로 미체결·종결 양쪽 목록을 pagination 완주하며 대조한다. `status` 파라미터는 그룹 라벨이고 `PARTIAL_FILLED`는 양쪽 그룹에 모두 속하므로(openapi), 유일성 판정 전에 **orderId 기준 dedup**을 수행해야 한다(SHALL — dedup 없이는 부분 체결 주문이 이중 매칭되어 영구 주차된다). 조회 응답은 멱등키를 싣지 않으므로 키로 매칭할 수 없다(SHALL NOT).
3. **부재 판정**: 최소 관찰 기간에 걸친 연속 N회(기본 3회) 안정화 조회 + 매수가능금액·보유수량 delta 교차 확인 후에만. 관측 창 동안 같은 심볼에 다른 mutation이 전송되었다면 delta 교차 확인은 무효이며 자동 FAILED_CONFIRMED는 금지된다(SHALL NOT — 부재를 증명할 수 없으므로 UNRESOLVED로).
4. **해소 불능**: UNRESOLVED_IN_DOUBT로 해당 심볼 신규 진입 영구 차단, 운영자 해소만 허용.

재시작 복구가 호출 순서를 소유한다(SHALL): 미종결 attempt 순회 → IN_DOUBT이고 재생 적격이면 Gateway의 해소 전용 재생 진입점 → 부적격·실패 시 조회 폴백. 해소 절차 자체는 mutator를 갖지 않는다(P1 불변식 유지 — 재생 진입점은 Gateway 소속). 조회 대조의 유일 매칭을 위해 심볼당 in-flight mutation 1개 제한은 **모든 safety class에** 유지된다(SHALL — 동시 반대 방향 제출은 브로커도 `422 opposite-pending-order-exists`로 거부한다(openapi)). §0.3은 동시성이 아니라 해소의 우선·유계성(재생이 이를 단축)과 수동 flatten의 순서화된 saga로 달성한다.

#### Scenario: 멱등 재생으로 정체 회수
- **WHEN** 능력 검증 완료 상태에서 제출 응답이 유실되고 TTL−margin 이내에 해소가 시작되면
- **THEN** 저장된 wire body의 재요청 응답에서 orderId를 회수해 CONFIRMED로 종결하며 두 번째 주문은 생성되지 않는다

#### Scenario: 두 번째 재생의 시간 재검사
- **WHEN** 첫 재생 후 재시도 시점에 경과 시간이 TTL−margin을 넘었으면
- **THEN** 두 번째 재생은 수행되지 않고 조회 폴백으로 전환된다

#### Scenario: 재생 중 409
- **WHEN** 재생 요청이 409 request-in-progress를 받으면
- **THEN** 상한을 소비하지 않고 대기 후 재시도하거나 조회 폴백으로 전환한다

#### Scenario: 부분 체결 주문의 양쪽 목록 출현
- **WHEN** 조회 폴백이 같은 orderId를 미체결·종결 목록 양쪽에서 발견하면
- **THEN** dedup 후 단일 매칭으로 CONFIRMED 종결되며 이중 매칭 주차가 발생하지 않는다

#### Scenario: 관측 창 오염
- **WHEN** 부재 판정 창 동안 같은 심볼의 다른 mutation이 전송되었으면
- **THEN** delta 교차 확인이 무효화되어 자동 FAILED_CONFIRMED 없이 UNRESOLVED로 처리된다

#### Scenario: 해소 불능
- **WHEN** 관찰 기간 내 존재도 부재도 증명되지 않으면
- **THEN** UNRESOLVED_IN_DOUBT로 표기되어 해당 심볼 신규 진입이 영구 차단되고 운영자 알림이 발송된다

**브로커가 이름을 준 거절은 code가 분류한다.**

브로커 응답이 **요청 자체를 서술하는 code**를 실었으면 그 code로 분류해야 하며, HTTP status로 분류해서는 안 된다(SHALL NOT). code가 확정 거절을 뜻하면 그 attempt는 IN_DOUBT로 가지 않고 종결해야 한다(SHALL).

`opposite-pending-order-exists`는 확정 거절이다(SHALL) — 브로커가 반대 방향 미체결 주문의 존재를 이유로 요청을 거절했고, 그 상태에서 주문은 접수되지 않는다.

이 판정은 응답 본문의 **`code` 필드 값**으로만 성립해야 한다(SHALL). status로 걸어서는 안 되고(SHALL NOT), message 문구로 걸어서도 안 되며(SHALL NOT), **본문 전체에 대한 부분문자열 일치로 걸어서도 안 된다**(SHALL NOT — 다른 필드가 그 문자열을 담을 수 있다). 브로커는 같은 code에 대해 계약과 다른 status와 message를 낼 수 있고, 실제로 냈다.

**code 필드의 자리와 표기는 하나가 아니다.** 실측된 본문은 두 모양을 갖는다 — 최상위 `{"code": "TRADE_AUTH_REQUIRED", …}`와 `{"error": {"code": "opposite-pending-order-exists", …}}`. 따라서 판정은 **최상위 `code`와 `error.code`를 모두 읽어야 하고(SHALL)**, 값 비교는 **대소문자를 무시한 전체 일치**여야 한다(SHALL — 부분문자열 일치는 위에서 금지했다). 어느 자리에서도 code를 얻지 못하거나 본문이 JSON이 아니면 **분류하지 않고 종전 경로로 보내야 한다**(SHALL — 모르는 것을 확정으로 바꾸지 않는다).

HTTP status 자체의 확정 거절 목록에 409를 추가해서는 안 된다(SHALL NOT — `409 request-in-progress`는 원 요청이 실행됐을 수 있으므로 확정 거절로 다루면 살아 있는 주문을 은퇴시킨다).

**이 분류를 재생 응답에 적용해서는 안 된다(SHALL NOT).** 재생에서 같은 code는 재생 요청이 거절됐다는 것만 말하고 원 주문에 대해서는 아무것도 증명하지 않는다 — 위 §1의 `422 idempotency-key-conflict` 규칙과 같은 구조다.

**종결된 mutation attempt는 그것이 무장한 보호 발의를 풀어야 한다.**

보호 청산의 발의는 그것을 실행하려던 attempt가 **CONFIRMED가 아닌 종결 상태**에 이르면 해제되어야 한다(SHALL). 대상 상태는 접수되지 않음이 확정된 경우, 전송되지 않은 경우, 그리고 판정 불능으로 park된 경우다.

**attempt가 CONFIRMED로 종결되면 발의를 해제해서는 안 된다**(SHALL NOT — 그 주문은 실제로 나갔고, 발의를 끝내는 것은 체결 경로의 일이다).

**발의를 attempt 종결 전에 해제해서는 안 된다**(SHALL NOT — 주문이 살아 있을 수 있는 동안 해제하면 다음 관측이 그 위에 두 번째 매도를 얹는다). 이 요구는 그 규칙을 **뒤집지 않고 완성한다**: 종결 전에는 무장을 유지하고, 종결 후에는 반드시 푼다.

**근거**: 오늘 모호한 결과를 받은 발의는 무장된 채 남고, 그 상태에서 청산 정책은 손절 조건이 성립해도 **빈 전이**를 돌려준다. 그 결과 해당 포지션은 손절도 익절도 제안되지 않는 상태로 **영구히** 머문다. 재시작 복구는 attempt만 순회하므로 이 상태를 해소하지 못한다. 실측(2026-08-07): 열린 포지션 다섯 중 둘이 이 상태였고, 그중 하나는 사건 발생 이후 계속 무보호였다.

발의 해제는 **보호 기준을 바꾸어서는 안 된다**(SHALL NOT — 진입가·최초 손절·유효 손절 중 어느 것도 이 경로에서 쓰이지 않는다).

발의 해제는 **멱등이어야 한다**(SHALL — 이미 해제된 발의에 대한 재호출은 아무것도 하지 않는다).

**저장된 모호한 attempt는 같은 증거로 다시 분류할 수 있어야 한다.**

이미 원장에 기록된 미종결 attempt 중 그 응답 본문이 **확정 거절 code**를 싣고 있는 것은 종결로 재분류되어야 한다(SHALL). 재분류는 기동 시 1회여야 하며(SHALL) 주기적으로 수행해서는 안 된다(SHALL NOT — 새로 생기는 응답은 전송 시점에 분류된다).

재분류는 **응답 본문의 code로만 판단해야 한다**(SHALL). 저장된 기록에 엔진 자신이 덧붙인 설명 문구를 매칭 대상으로 삼아서는 안 된다(SHALL NOT). 확정 거절 code가 없는 attempt는 그대로 두어야 한다(SHALL NOT 재분류 — 모호한 것은 모호한 채 둔다).

재분류는 **attempt의 상태만 바꾸어야 한다**(SHALL NOT 발의 해제 — 그것은 위 요구가 하는 일이며, 두 쓰기를 한 절차에 섞으면 어느 쪽이 실패했는지 말할 수 없다).

**세션 중 해소는 이 change의 범위가 아니다.**

미종결 attempt를 엔진이 도는 동안 해소 절차에 넣는 것은 **별도 change로 분리한다**. 그 change는 해소 진입점의 배선, 주기와 동시 실행 상한의 실측, 그리고 park가 계정 전역 진입을 차단한다는 운영 결과의 공시를 **모두** 포함해야 한다(SHALL). 그 셋 없이 세션 중 해소를 도입해서는 안 된다(SHALL NOT).

재시작 복구의 재시작 규칙(RECORDED → NOT_DISPATCHED, DISPATCH_STARTED → IN_DOUBT)을 세션 중에 적용해서는 안 된다(SHALL NOT — 세션 중 그 두 상태는 진행 중인 전송이며, 그것을 재시작 규칙으로 종결시키는 것은 원장 위조다).

#### Scenario: 종결된 attempt가 발의를 푼다
- **WHEN** 보호 발의를 실행하려던 attempt가 접수되지 않음으로 확정되면
- **THEN** 그 발의가 해제되어 다음 관측이 같은 조건에서 다시 제안할 수 있다

#### Scenario: 접수된 주문의 발의는 유지된다
- **WHEN** attempt가 CONFIRMED로 종결되면
- **THEN** 발의는 해제되지 않는다

#### Scenario: 판정 불능으로 park된 attempt
- **WHEN** attempt가 존재도 부재도 증명되지 않아 park되면
- **THEN** 그 발의도 해제되어 그 포지션이 무보호로 영구히 남지 않는다

#### Scenario: 발의 해제는 손절 가격을 바꾸지 않는다
- **WHEN** 발의가 해제된다
- **THEN** 진입가·최초 손절·유효 손절은 종전 값 그대로다

#### Scenario: 이미 해제된 발의
- **WHEN** 이미 해제된 발의에 대해 해제가 다시 요청되면
- **THEN** 아무것도 바뀌지 않고 오류도 발생하지 않는다

#### Scenario: 저장된 모호한 attempt의 재분류
- **WHEN** 기동 시, 저장된 미종결 attempt의 응답 본문에 확정 거절 code가 있으면
- **THEN** 그 attempt는 종결로 재분류되고, code가 없는 attempt는 그대로 남는다

**부재 판정의 증거 모델은 매도에 대해 성립해야 한다.**

부재 판정의 매수가능금액·보유수량 delta 교차 확인은 **매수 예약**을 증거로 삼는다. 접수됐으나 미체결인 매도는 두 값을 모두 움직이지 않으므로, 그 교차 확인만으로 매도의 부재를 확증해서는 안 된다(SHALL NOT — 확증하면 살아 있는 매도를 은퇴시키고 다음 주기가 두 번째 매도를 낸다).

매도 mutation에 대해 사전 계정 기준선을 공급하려면 **매도용 부재 증거 모델이 먼저 정의되어야 한다**(SHALL). 그 모델 없이 기준선만 공급해서는 안 된다(SHALL NOT — 기준선의 부재가 park를 강제하는 현재 동작이 그 경우의 안전측이다).

기준선의 어떤 필드가 미측정인지 표현할 수 없다면 그 필드를 0으로 채워서는 안 된다(SHALL NOT — 0은 "변화 없음"과 구별되지 않아 해당 교차 확인을 침묵으로 통과시킨다).

#### Scenario: 브로커가 계약과 다른 status로 알려진 code를 낸다
- **WHEN** 주문 제출이 `opposite-pending-order-exists` code를 실은 응답으로 거절되고 그 HTTP status가 openapi가 선언한 422가 아니라 409이면
- **THEN** attempt는 code로 확정 거절 분류되어 종결하며, IN_DOUBT로 가지 않고 그 심볼의 다른 mutation을 차단하지 않는다

#### Scenario: 같은 문자열이 code가 아닌 필드에 있다
- **WHEN** 응답 본문의 `message`나 다른 필드에만 그 문자열이 있고 code 필드는 다른 값이면
- **THEN** 확정 거절로 분류되지 않는다

#### Scenario: code가 최상위에 있는 본문과 error 아래에 있는 본문
- **WHEN** 같은 code가 한 응답에서는 최상위 `code`로, 다른 응답에서는 `error.code`로 오면
- **THEN** 둘 다 같은 분류를 받는다

#### Scenario: 본문이 JSON이 아니거나 code가 없다
- **WHEN** 거절 응답의 본문을 JSON으로 읽을 수 없거나 어느 자리에도 code가 없으면
- **THEN** 이 분류는 적용되지 않고 attempt는 종전 경로로 처리된다

#### Scenario: 같은 status의 다른 code는 확정되지 않는다
- **WHEN** 재생 요청이 `409 request-in-progress`를 받으면
- **THEN** 종전대로 상한을 소비하지 않고 대기 후 재시도하거나 조회 폴백으로 전환하며, 확정 거절로 종결되지 않는다

#### Scenario: 매도의 부재는 잔고 delta만으로 확증되지 않는다
- **WHEN** 미체결 매도의 IN_DOUBT 해소에서 보유수량과 매수가능금액이 기준선과 같으면
- **THEN** 부재는 확증되지 않고 UNRESOLVED로 park한다
