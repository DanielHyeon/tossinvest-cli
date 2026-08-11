## Context

보유가 있으면 브로커에 손절이 남아 있어야 한다. 지금은 남지 않는다. 엔진 프로세스가 죽으면
보호도 같이 사라진다(`internal/execgw/protection.go:55-62`).

이 change는 `a071-wire-kr-us-protection-readiness` task 3.5 중 **official protection gateway를
프로덕션에 배선하는 부분**을 승계한다. 같은 task의 supervisor·`Wired` 생산 부분은 a105로 간다.
분할 근거는 `proposal.md`의 「Supersession」과 `review.md`에 있다.

a071이 만든 계약은 그대로 쓴다. market-scoped `WIRED|UNWIRED` verdict, paired snapshot,
sealed supervisor binding, Gateway decision boundary는 이 문서에서 다시 설계하지 않는다.

### 범위를 이렇게 자를 수 있는 이유

`internal/execgw/protection.go:89-91`은 `!plan.raisesExposure`면 즉시 반환한다. `raisesExposure`는
매수(`gateway.go:377`), 취소는 `false`(`:416`), 정정은 증가 여부(`:447`)다. AST 산출물이 이
5개 분기를 전부 열거하고 6개 경로가 모두 실행됨을 측정했다
(`analysis/function-logic/internal-execgw--gateway.checkprotection/`).

**⇒ 매도인 보호주문은 readiness를 조회하지 않는다.** 브로커에 손절을 남기는 데 `Wired`가
필요하지 않다. `Wired`의 유일한 소비자는 진입 인터록이고, 진입에는 그 밖에 세 개의 잠금이
더 걸려 있으므로 이 change가 그것을 열 수도 없다.

### 오늘의 대상 집합

레인 6개가 모두 OFF이므로 신규 체결은 없다. 보호가 필요한 것은 **이미 계좌에 있는 보유분**이고,
`ReconcileDriver`가 이미 그것을 exit 관리로 편입하면서 `SyntheticStop`을 journal의
`InitialStop`으로 얼려 둔다(`internal/journal/adoption.go:289`). 없는 것은 그 값을 브로커에
남기는 경로 하나다.

## Goals / Non-Goals

### Goals

- journal에 커밋된 포지션 상태에서 정확한 보호 수량·trigger·만료를 유도하고, 브로커에 상주하는
  보호주문을 durable·idempotent하게 설치·교체·취소·복구한다.
- **기존 보유와 신규 체결을 같은 경로로 처리한다.** 촉발은 이벤트가 아니라 상태다.
- 보호 미설치 시간이 관측 가능하고, 상한을 넘으면 알림이 난다.
- 배선 대상 함수의 미실행 거부 분기 9개를 배선 **이전에** RED 테스트로 덮는다.
- 한 포지션에 브로커측 매도 청구권이 둘이 되지 않는다.

### Non-Goals

- `Wired` 생산, supervisor, manifest 재서명, coverage latch → **a105**.
- 레인 활성화(a105), threshold 승인(a101), 라이브 평가(a103), 사이징 역산(a104).
- `internal/filldetect` 편집. 조립 지점만 바꾼다(D3).
- `internal/protection.Controller`와 `Repository`의 부활(D1).

## Decisions

### D1. 프로덕션 상태 core는 `protectionlifecycle`이다. `protection.Controller`는 쓰지 않는다

저장소에 보호 core가 두 개 있다. 어느 쪽인지는 취향이 아니라 **DB 의존성**이 정한다.

| | `internal/protection` | `internal/protectionlifecycle` |
| --- | --- | --- |
| 생성 | 2026-08-01 (a045) | 2026-08-04 (a071) |
| 크기 | `controller.go` 824 + `repository.go` 540 + `domain.go` 531 | `lifecycle.go` 18.6K + `state.go` 8.8K |
| 상태 저장 | `database/sql` — `protection_sagas`, `protection_mutation_attempts` (`repository.go:17,44`) | 없음 (순수 전이) |
| non-test importer | `execgw`, `app/engine/gateway.go` — **domain 타입만** | **0** |

`NewController`는 `*Repository`를 요구하고(`controller.go:102-105`), `NewRepository`는 `*sql.DB`를
요구한다(`repository.go:95`). 즉 Controller를 프로덕션에 넣으면 **두 번째 SQLite가 기동
의존성이 된다.** 그것은 a071의 H5 disposition이 명시적으로 배제한 것이며, 봉인 가드도
`protection.db` 문자열을 금지 심볼로 검사한다(`dormant_test.go:83`).

**결론:** a100은 세 조각을 조립한다.

```text
protectionlifecycle   순수 상태 전이 (applyFill, prepareRegister, EntryOpen)
        ↕ (D4: journal이 durable store)
internal/protection   broker-neutral 도메인 타입 (Scope, ConditionalBody, BrokerTarget, BrokerProtection)
        ↕
protectionofficial    official API 전송 (Create/Replace/Cancel/Get/List/Sellable)
```

**그런데 이 조립은 지금 형태로 컴파일되지 않는다** (adversarial Eng 리뷰 발견, 코드로 확인).

`internal/protectionlifecycle`의 lifecycle 함수는 **전부 패키지 레벨 unexported**다 —
`prepareRegister`, `applyFill`, `prepareReplace`, `recoverSubmit` … 하나도 exported가 아니다.
그리고 `external_api_test.go:11`의 `TestProductionAPIExportsNoAuthorityMintingFunction`이
**exported 패키지 레벨 함수가 하나라도 있으면 실패**시킨다. exported 메서드도 없다
(`State`의 메서드 셋은 `view`/`marketEntryOpen`/`reseal`로 전부 unexported).

⇒ **이 패키지는 외부에서 호출할 수 있는 표면이 0이다.** 그리고 `dependency_test.go:12`가
journal·protection·execgw·app 의존을 금지하므로 워커를 이 패키지 **안에** 둘 수도 없다.

즉 봉인은 넷이 아니라 **다섯**이었다. `internal/protection`의 금지 심볼 4개에 더해
`protectionlifecycle`의 API 가드가 있고, 원안 a100도 이 리뷰 이전 판도 그것을 보지 못했다.

**결정: 다섯 번째 봉인을 같은 change에서 명시적으로 연다.** 가드의 이름이 의도를 말한다 —
`NoAuthorityMintingFunction`. 지키려는 것은 「패키지가 영원히 호출 불가」가 아니라
「**권한을 발행하는** 함수를 노출하지 않는다」이다. 따라서:

- 노출하는 것은 **상태 전이 하나당 하나의 명명된 exported 함수**로 최소화하고, 각각이
  받는 것은 `State`와 값 타입뿐이다. 브로커 전송·승인·토글에 닿는 인자를 받지 않는다.
- 가드를 지우지 않고 **대체 단언으로 바꾼다**: exported 표면이 허용 목록에 정확히 일치하고,
  그중 어느 것도 transport·approval·toggle 타입을 시그니처에 갖지 않음을 정적으로 검사한다.
- `dependency_test.go`의 의존 금지는 **그대로 둔다.** 순수 core를 유지하는 것이 이 가드의 값이다.

**이 결정은 이 change의 봉인 표면을 넓힌다.** D6의 표에 다섯 번째 행으로 올린다.

**대가:** `protection.Controller` 824줄과 `Repository` 540줄이 프로덕션에서 죽은 코드로 남는다.
리뷰의 두 보이스가 모두 "a100 **이전에** 지우라"고 했으나 채택하지 않았다 — 삭제는 `execgw`가
쓰는 도메인 타입과 얽혀 있어 별도 리뷰 단위이고, 지금 합치면 이 change의 변경 집합이 다시
리뷰 불가능해진다. **대신 정리 change를 a100 완료가 아니라 지금 등록한다**(tasks 6.5).

### D2. 촉발은 이벤트가 아니라 상태다 — 수렴 루프

원안은 「체결 delta > 0」을 촉발로 삼았다. 그 설계는 **오늘 계좌에 손절을 한 건도 남기지
않는다.** 레인이 전부 OFF여서 신규 체결이 없기 때문이다. `proposal.md`의 Why가 기존 보유를
"더 위험한 쪽"이라고 부르면서 촉발이 그쪽에 닿지 않는 모순이었다.

**수렴 대상(desired state):** journal에 exit 관리 상태가 열려 있고 유효한 stop을 가진 모든
포지션은, 브로커에 그 stop을 실행하는 ACTIVE 상주 보호주문을 정확히 하나 가진다.

**수렴 조건(reconciliation):** 매 주기마다 desired와 observed를 비교한다.

| observed | 행동 |
| --- | --- |
| 보호 컬럼 NULL | 등록 시도 |
| pending (등록 응답 미확인) | 브로커에 exact operation 조회로 귀속. 증명 없으면 재제출하지 않음(a071 계약) |
| ACTIVE, 수량·trigger 일치 | 무동작 — **수렴 완료** |
| ACTIVE, 보유 수량 증가 | 더 안전한 방향으로 교체 |
| ACTIVE, 보유 수량 감소 | 축소 교체. trigger 후퇴는 거부 |
| terminal (다 채워짐/취소됨) | 포지션이 남아 있으면 재등록, 없으면 정리 |

체결은 이 상태를 만드는 여러 원인 중 하나다. **체결 경로에 특별한 배선이 필요하지 않다** —
체결이 journal에 커밋되면 다음 수렴 주기가 그것을 본다. 이것이 D3을 가능하게 한다.

**수렴 완료의 정의를 관측 가능하게 만든다.** 포지션별 `보호 미설치 시간`(desired가 생긴
시각부터 보호 확인까지)을 측정하고, 상한을 넘으면 알림을 낸다. 상한이 없으면 "수렴 중"과
"영원히 실패 중"이 구별되지 않는다.

**단, 「ACTIVE」는 지금 어댑터로 판정할 수 없다** (adversarial Eng 리뷰 발견, 코드로 확인).

```go
func lifecycle(status string, triggered bool) (bool, error) {
	switch status {
	case "WATCHING", "PAUSED", "ORDERING", "ORDERED":
		return triggered, nil
```

(`protectionofficial/gateway.go:308-310`.) 네 상태가 전부 같은 값으로 접힌다 — **무장된 주문과
일시정지된 주문이 구별되지 않는다.** `protection.BrokerProtection`은 `Terminal`·`Triggered`만
싣고 raw status를 싣지 않는다(`protection/domain.go:452`). `Quantity`와 `Trigger`는 싣는다.

따라서 수렴 판정은 다음으로 정의한다.

- **비교 가능한 것으로만 판정한다**: `!Terminal && !Triggered`, `Quantity == 보유 수량`,
  `Trigger == 유도값`. 이 셋은 어댑터가 실제로 노출한다.
- **`PAUSED`를 무장으로 오인할 수 있다는 것을 명시한다.** raw status를 도메인 타입에 실을지는
  `internal/protection`·`protectionofficial` 편집을 요구하므로 **tasks 0에서 결정**하고,
  실지 않기로 하면 M-A가 `PAUSED`가 실재하는 상태인지 관측한다. 실재하고 발동하지 않는다면
  **raw status 노출이 필수 task가 된다.**
- **"수렴하면 브로커를 다시 묻지 않는다"를 철회한다.** 상주 주문은 나중에 취소·만료·일시정지될
  수 있고, 그때 보호 미설치 시간이 다시 시작되어야 한다. 캐시가 아니라 **주기를 늘린
  재확인**으로 한다(rate limit은 주기로 다루고, 판정의 신선도를 버리지 않는다).

### D3. 수렴은 체결 감지 경로 **밖에서** 돈다 — 원안의 `Ledger` 데코레이터는 기각

원안 D8은 `filldetect.Ledger` 데코레이터로 보호를 계획했다. **AST 산출물이 그것을 뒤집었다.**

먼저 원안이 인용한 함수 이름이 틀렸다. `Detector.PollOnce`는 L277-281의 5줄 래퍼이고 **분기가
0**이다. 인용된 로직은 `Detector.pollLocked`(L283-357, 분기 10)에 있다
(`analysis/function-logic/internal-filldetect--detector.polllocked/`).

그 산출물을 측정하니 두 경로가 드러났다.

**(1) 에러 경로 — B6은 `continue`가 아니라 `return`이다.**

```go
for _, snap := range snaps {                       // B5, L316
	applied, err := d.Ledger.Apply(ctx, snap)
	if err != nil {                                // B6, L318 — 측정: 미실행
		d.outage.failure(clk.Now(), err)
		cycle.FinishedAt = clk.Now()
		return cycle, fmt.Errorf(...)              // 남은 스냅샷을 버린다
	}
```

한 주문의 `Apply` 실패가 같은 사이클의 **다른 주문들의 체결 반영까지** 중단시킨다. 그리고 이
분기는 **한 번도 실행된 적이 없다.** 데코레이터는 그것을 프로덕션에서 처음 실행되게 만드는
배치였다. 원안 tasks 4.6이 이 경로를 격리하려 했지만, 격리는 "에러를 삼킨다"는 규율이고
규율은 코드 리뷰가 놓치면 사라진다.

**(2) 지연 경로 — 원안이 다루지 않았다.**

```go
	latency := committed.Sub(snap.BrokerVisibleAt)  // L335
	...
	d.slo.observe(committed, latency)               // L342
}
...
d.evaluateSLO(now)                                  // L349 → Gate.Block(ReasonFillDetectionSLO)
```

`applied.CommittedAt`은 journal이 커밋에 찍은 시각이므로(`ledger.go:106-107`) **자기 자신의**
보호 왕복은 섞이지 않는다. 그러나 루프는 직렬이다(B5). 앞 스냅샷의 `Apply` 안에서 보호 왕복을
하면 **뒤 스냅샷의 커밋이 그만큼 밀리고**, 밀린 만큼 그 스냅샷의 신선도 표본이 커진다. 표본이
임계를 넘으면 `evaluateSLO`가 체결 감지 게이트를 잠근다. 사이클 자체도 길어지고, `Run`은
사이클이 **끝난 뒤에** `PollInterval`을 자므로(`detect.go:523`) 다음 관측도 늦는다.

**보호를 설치하려다 체결 감지를 잃는다.** 안전 불변식 §0-3 위반이며, 이 change의 자기 spec
delta(`specs/fill-detection/spec.md`)가 이미 금지하고 있던 것이다 — 설계가 자기 spec과
모순이었다.

**채택: 독립 수렴 워커.** 보호 수렴은 체결 감지와 대사 어느 루프에도 들어가지 않는다.

- 자기 주기로 돈다. `filldetect`도 `ReconcileDriver`도 편집하지 않는다.
- 입력은 journal의 커밋된 상태뿐이다. D4의 "커밋 이후에만 계획한다"가 **구조적으로** 성립한다 —
  워커는 커밋되지 않은 것을 볼 방법이 없다.
- 실패는 자기 안에서 끝난다. 다른 루프의 outage·SLO·게이트를 건드리지 않는다.
- 기존 함수 내부를 편집하지 않으므로 그 함수들의 실패 의미를 바꾸지 않는다.

**「커밋된 것만 읽는다」는 최신을 보장하지 않는다** (adversarial Eng 리뷰 발견). journal은
statement 단위로 직렬화될 뿐이고, 워커가 브로커 왕복을 도는 동안 `ReconcileDriver`가 같은
포지션의 수량을 조정하거나 포지션을 닫을 수 있다(`app/engine/reconcileloop.go:413`,
`internal/reconcile/converge.go:209`, `:244`).

> 구체적 경합: 워커가 10주를 읽는다 → 대사가 0주를 커밋한다 → 워커가 SELL 10을 등록한다.

따라서 **전송 직전 재확인과 응답 후 재확인을 계약에 넣는다.**

- 전송 **직전**에 포지션 수량·generation을 다시 읽고, 계획 시점과 다르면 **전송하지 않고**
  그 주기를 버린다. 다음 주기가 새 상태로 다시 계획한다.
- 응답을 받은 뒤에도 다시 읽어, 그 사이에 수량이 바뀌었으면 **수렴 완료로 기록하지 않고**
  다음 주기가 교체·취소로 수렴시킨다.
- tasks 5.3은 워커 주기 두 개의 겹침만 재현했다. **대사와의 겹침을 별도 fixture로 재현한다.**

**워커의 실행 주체와 기동 순서를 정한다** (adversarial Eng 리뷰 발견). 현재 감독되는 루프는
reconcile·exit·filldetect·strategy 넷이고(`cmd/tossctl/engine.go:377`) gateway 조립에는 워커
필드가 없다(`app/engine/gateway.go:92`). "`gateway.go`에서 조립"만으로는 **누가 돌리는지가 없다.**

- 조립은 `gateway.go`(import 경계 규칙), **기동·취소·감독은 `cmd/tossctl`의 런타임**이 한다.
  둘을 분리하지 않으면 `buildGateway` 안에서 도는 goroutine이 automation interlock 평가보다
  먼저 시작한다(`app/engine/engine.go:489`, `:533`).
- **automation gate가 verified가 아니면 워커는 돌지 않는다.** 대사·exit 루프가 같은 이유로
  거부한다(`reconcileloop.go:342` — "It refuses on an unverified automation gate … the loop
  writes to the ledger and, with adoption on, starts protecting positions with real sell orders").
  보호주문 등록은 실제 매도 주문이므로 같은 규율을 받는다.

`ReconcileDriver.RunOnce` 안의 한 단계로 넣는 대안은 기각했다. 그러면 기존 함수 내부를
편집하게 되고(FLM 대상이 늘고), 보호 실패가 대사 사이클의 실패 의미와 섞일 위험이 (1)과 같은
형태로 되살아난다. **같은 실수를 다른 루프에서 반복하지 않는다.**

### D4. 보호 상태의 durable store는 journal이며, 스키마 변경은 additive-nullable이다

`protectionlifecycle`은 순수하다 — 상태를 저장하지 않는다. D1이 별도 SQLite를 배제했으므로
남는 곳은 **기존 trading journal** 하나다.

```text
포지션 상태(체결이든 편입이든)가 journal에 커밋됨
        ↓  (여기부터 a100 — 워커가 읽는다)
stop/expiry 유도 → Plan → Register → broker
        ↓
broker order id와 상태를 journal에 커밋
```

스키마는 **additive-nullable만** 허용한다. 기존 컬럼의 의미를 바꾸지 않고, 새 컬럼은 전부
nullable이며, 값이 없는 행은 「보호 미설치」로 읽힌다.

#### FLM(tasks 0.4)이 이 규칙에 이빨을 붙였다 — 세 판정 리스트

exit 상태 행을 읽는 **단일 지점**은 `journal.scanExitStateResult`이고, 그 함수의 분기 22개 중
**20개가 부패 판정**이다. 판정은 세 층이며, 보호 컬럼은 **어디에도 들어가면 안 된다.**

| 판정 | 코드 | 넣었을 때 |
| --- | --- | --- |
| `v10Evidence` 17항목 → `anyV10` | `exit_snapshot.go:620-628` | v10 이전 행에 보호를 설치하는 순간 legacy 경로를 벗어나 `partial_snapshot_tuple`로 **부패 판정** |
| `full` 완결성 | `:674-676` | 보호 없는 정상 행이 즉시 `partial_evaluated_tuple` |
| 평탄화 컬럼 대 저장 JSON 비교 | `:689-695` | 보호 컬럼은 그 JSON에 없으므로 **모든 기존 행이 `flattened_snapshot_mismatch`** |

세 결과가 같다 — **부패로 판정된 exit 상태는 그 포지션의 exit 정책을 멈춘다. 보호를 설치하려다
손절을 끄는 것이므로 §0-4 위반이다.**

⇒ 「additive이고 nullable이며 기존 컬럼의 의미를 바꾸지 않는다」가 이 함수에서 갖는 구체적
의미는 **`row.Scan` 인자와 `ExitState` 필드에는 추가하되 세 판정 리스트에는 추가하지 않는다**이다.
tasks 2.3은 이것을 명시적 조건으로 갖는다.

**목록을 읽는 쪽도 하나다.** `Journal.OpenExitStates`의 주석이 "It is two things at once, and
**deliberately not two functions**"라고 못박았다(`apply_hook.go:614-620`). ⇒ 수렴 워커를 위한
별도 조회 함수를 만들지 않는다. 그리고 그 SQL의 lifecycle 필터 때문에 **재편입 중인 포지션은
목록에서 잠시 사라진다** — 「목록에 없음」은 「보호가 필요 없음」이 아니므로, 워커의 취소 판정은
이 목록의 부재만으로 내리지 않고 브로커 관측을 함께 본다.

**등록 전에 pending을 durable하게 커밋한다.** `recoverSubmit`은 영속된 pending command가 없으면
"no exact submit pending"으로 시장을 latch한다(`protectionlifecycle/lifecycle.go:63`, `:71`).
`prepareRegister`가 전송 **전에** operation key·generation·revision·body·pending 상태를 만든다
(`:23`). 따라서 저장 순서는 「plan → **pending 커밋** → 브로커 전송 → 응답 커밋」이며, pending
커밋을 건너뛰면 전송 직후 프로세스가 죽었을 때 조회할 operation identity가 존재하지 않는다.
tasks 2.3의 컬럼 목록은 이 pending 레코드를 포함해야 한다.

### 롤백은 「이전 바이너리로 같은 journal을 연다」가 아니다

리뷰 이전 판은 "롤백된 바이너리가 새 컬럼을 몰라도 기존 경로가 그대로 동작해야 한다"고 적었다.
**코드가 그것을 거부한다.**

```go
// ErrSchemaTooNew means the journal on disk was written by a newer TossOS. An
// older binary must not touch it: it would read columns it does not know about as
// absent and could take an unsafe decision from an incomplete picture.
```

(`internal/journal/journal.go:23-27`, 강제 지점은 `:240`.) 그리고 코드의 이유가 우리 것보다 낫다 —
보호 컬럼을 모르는 바이너리는 「보호 설치됨」을 「미설치」로 읽고 그 위에서 판단한다.
프로젝트 기억 `tossos-branch-behind-main-schema`가 기록한 2026-08-04 실발생이 같은 기전이다.

⇒ **additive-nullable은 "구버전 호환"을 주지 않는다.** 그것이 주는 것은 기존 컬럼의 의미가
바뀌지 않는다는 것뿐이고, 하향 호환은 애초에 이 journal의 계약이 아니다.

**따라서 롤백 계약을 다시 쓴다.**

- 스키마를 올린 뒤의 롤백은 **백업 복원**이며, 복원된 journal에는 백업 이후의 모든 변경이 없다
  (`internal/journal/backup.go`). **여기에 상주 보호주문의 기록도 포함된다.**
- 그 결과 구버전 엔진은 브로커에 있는 상주 주문을 모른 채 자기 청산을 낸다 — **매도 권한이
  둘이 되는 구체적 시퀀스다.** 브로커가 두 번째 매도를 거부하는지는 **미확인(UNVERIFIED)**이다.
- 그러므로 롤백 절차는 자동이 아니라 **사람 절차**다: (1) 브로커의 상주 보호주문 목록을 먼저
  기록하고, (2) 백업을 복원하고, (3) 기록한 목록을 근거로 각 주문을 유지할지 취소할지 사람이
  정한다. **a100은 이 절차를 문서로 만들고, 자동 취소도 자동 유지도 하지 않는다.**
- 「롤백이 상주 주문을 자동 취소하지 않는다」는 계약은 유지한다. 다만 그것이 **안전을 보장하지
  않는다**는 것을 함께 적는다 — 보장하는 것은 사람 절차뿐이다.

### D5. 이 change는 진입 latch를 만들지 않는다

원안 D5는 포지션 단위 coverage latch를 `strategy_entry_supervisor.go`에서 소비하게 했다.
**기각한다 — 이 change에서 그 latch는 구조적으로 무동작이다.**

진입은 이미 네 겹으로 닫혀 있다: (L1) 레인 6개 OFF, (L2) `checkProtection`이 `ProtectionWired`가
아니면 노출 증가 mutation 거부, (L3) threshold 미승인, (L4) 서명 manifest 부재. a100은 `Wired`를
생산하지 않으므로 L2가 그대로 닫혀 있다. **coverage latch가 열든 닫든 진입은 0건이다.**

무동작인 latch를 지금 만들면 (a) 검증할 수 없는 코드가 늘고, (b) `EntryPermitted` 단언을 바꾸는
기존 테스트 편집이 딸려 오며, (c) 그 latch를 실제로 필요로 하는 change(a105)가 아닌 곳에서
설계가 굳는다. **coverage latch는 a105의 것이다.**

이 change에서 보호 수렴 실패의 귀결은 **typed reconcile reason + 관측 가능한 알림**이다.
진입을 닫을 필요가 없다 — 이미 닫혀 있다.

### D6. 봉인 가드는 최소로 연다 — 금지 심볼 4개 중 1개만, 그리고 경계를 넓힌다

봉인은 사고가 아니라 의도다. 그러므로 해제도 최소여야 하고, 무엇을 열었는지 정확히 적어야 한다.

| 금지 심볼 | a100의 처리 | 이유 |
| --- | --- | --- |
| `protectionofficial.New` | **해제** | 브로커 전송을 조립하려면 반드시 필요하다 |
| `protection.NewSupervisor` | **유지** | a100은 supervisor를 만들지 않는다 |
| `protection.db` | **유지** | D1이 별도 SQLite를 배제했다 |
| `GatewayFactory` | **유지** | 임의 factory 주입은 여전히 두 번째 mutation 경로다 |
| `protectionlifecycle` API 가드 | **최소 해제 + 대체 단언** | D1 — 이 가드를 열지 않으면 조립 자체가 컴파일되지 않는다. 허용 목록 일치 + transport/approval/toggle 인자 부재를 정적으로 검사한다 |

**리뷰가 찾은 구멍을 함께 막는다.** `dormant_test.go:59`의 import walk는 정확히
`…/internal/protection`만 매칭하고, 허용 파일은 `internal/app/engine/gateway.go` 하나다.
`internal/protectionofficial`과 `internal/protectionlifecycle`은 **그 walk의 대상이 아니다.**
즉 원안대로면 「단일 mutation 경로」는 도메인 타입 패키지에 대해서만 강제되고, 실제 브로커
전송 패키지는 아무 app 파일에서나 import할 수 있었다.

a100은 그 walk를 **`protectionofficial`·`protectionlifecycle`까지 확장**한다. 봉인을 하나
열면서 경계는 넓힌다.

**FLM(tasks 0.4)이 허용 목록에 대한 계획을 바꿨다.** 허용 목록은 파일 단위이고
(`dormant_test.go:61` `allowed := map[string]bool{"internal/app/engine/gateway.go": true}`),
walk는 `cmd/`와 `internal/app/` **전수**를 돈다. 이전 판은 "허용 파일 목록을 `gateway.go`와
워커 조립 지점으로 명시한다"고 적었으나, 그러면 **두 번째 조립 경로가 생긴다 — 가드가 막으려던
바로 그것이다.** ⇒ **허용 목록은 `gateway.go` 하나 그대로 둔다.** 수렴 워커는 이 패키지들을
import하지 않고, `gateway.go`가 만든 값을 좁은 인터페이스로 받는다. 확장한 경계에도 같은
규칙을 적용한다.

해제하는 하나에는 **대체 단언**을 붙인다: 조립된 보호 경로가 journal-backed이고 별도 DB에
의존하지 않으며 조립 지점이 정확히 하나임을 정적으로 검사한다.

### D7. `dispatch`는 건드리지 않는다 — FLM 면제 확정

`strategyDispatchCycle.dispatch`는 **편집하지 않고, 그 내부 분기를 근거로 인용하지도 않는다.**
(리뷰 이전 판이 적었던 분기·return 수치는 이 change에 산출물이 없으므로 지웠다 — 면제 사유에
분기 수를 쓰는 것 자체가 산출물 없는 분기 주장이다.) D2가 수렴을 독립 워커에 두었고 D5가 진입 latch를 만들지 않으므로 `dispatch`
내부에 진입할 이유가 없다. 따라서 `Function Logic Map: not-applicable — 편집하지 않고 내부
분기를 근거로 쓰지 않음`이다.

**이 면제는 조건부로 남는다.** 구현 중 편집하거나 분기를 근거로 인용하게 되면 **그 시점에**
AST를 만들고 해당 task 착수 전에 완성한다.

### D8. 브로커측 매도 청구권을 하나로 **좁힌다** — 완전히 하나로 만들지는 못한다

측정이 이 문제를 직접 지목했다. `verify-execution-capability/measurements.md:55`(M13):

> **조건주문은 매도가능수량을 예약하지 않는다** — 등록 전 115, 등록 후에도 115. … 「한 심볼에
> 브로커측 매도 청구권 1개」 불변식을 **우리가 강제해야 하는 이유**.

a100 이후 한 포지션은 두 매도 권한을 갖는다: (i) 브로커 상주 조건주문 (ii) a087의 인프로세스
보호성 시장가 매도. 브로커가 수량을 예약해 주지 않으므로 **둘이 동시에 나가면 초과 매도가
가능하다.** 그리고 a091이 "매도 0인 손절은 critical"로 정의해 뒀으므로, 조건주문이 먼저 채워진
뒤 인프로세스 매도가 0주를 팔면 **정상 동작이 critical 알림을 만든다.**

계약을 정한다.

1. **인프로세스 보호 매도가 상위 권한이다.** 엔진이 살아 있으면 즉시성이 더 높다.
2. 인프로세스 보호 매도를 내기 **전에** 그 포지션의 상주 보호주문을 취소한다. 취소 확인 실패는
   매도를 막지 않는다 — §0-3이 청산 지연을 금지한다. 대신 typed reason을 남기고 대사가 뒤처리한다.

   > **FLM(tasks 0.4)이 이 규칙에 기전을 붙였다.** 취소가 실행되는 지점은
   > `ExitObserver.record`의 심볼 정리 블록이고, 그 블록은 `clearTheSymbol`의 결과를 두 가지
   > 방식으로 매도에 반영한다 — 에러면 `return err`로 **매도 없이 반환**하고(L1142, 커버리지
   > 측정상 **한 번도 실행된 적 없음**), `cleared == false`면 `orderable = false`로 **매도를
   > 보류**한다(L1145). ⇒ **상주 주문 취소를 `clearTheSymbol` 안에 넣거나 그 반환값에
   > 반영하면 취소 실패가 손절을 막는다.** 별도 시도로 두고 `err`·`cleared`·`orderable`
   > 어느 것도 그 결과로 바꾸지 않는 것이 이 규칙의 구현 조건이다.
   > 보류 사유(`ArmSuppressedReason`)에도 새 값을 추가하지 않는다 — 추가하는 순간 그것이
   > 「매도를 보류했다」는 뜻이 되기 때문이다.
3. 상주 주문이 먼저 채워져 포지션이 이미 flat이면 **보호 청산을 시도하지 않는다.** 그 포지션은
   보호에 실패한 것이 아니라 **보호된 것**이므로 `ProtectionCompletedByResting`으로 기록한다.

   > 이 형태를 고른 이유가 있다. a091이 `engine-safety`에 "확정 하한이 보호 청산의 제출 수량을
   > 0으로 만들면 critical 등급으로 보고해야 한다 (SHALL)"를 이미 넣어 뒀다. 상주 주문이 먼저
   > 채워진 경우를 **0주 청산 시도로 흘려보내면 정상 동작이 critical 알림을 만든다.** 반대로
   > 그 요구사항에 예외를 파면 a100의 delta가 MODIFIED가 되어 a071·a091과 archive 순서로 얽힌다.
   > **시도 자체를 하지 않으면 a091의 술어에 도달하지 않는다** — 요구사항을 건드리지 않고
   > 오탐도 만들지 않는 유일한 형태다.

3b. **익절도 완전 청산이고, 판정 기준은 이미 `isFullExit`다 — 넓힐 것이 없었다.**
   `isFullExit`는 `ActionLadderTakeProfit`을 포함하고 `isProtective`는 포함하지 않는다
   (`app/engine/exitloop.go:1207-1221`).

   > **FLM(tasks 0.4)이 이 항목의 이전 판을 정정했다.** 2차 개정은 "완전 청산을 내는 모든
   > 경로가 취소를 시도하도록 **넓힌다**"고 적었으나, AST가 보여 준 `record` L1117은
   > 이미 `orderable && (snapshot.CancelPendingFirst || isFullExit(proposal))`이다.
   > **진입 조건은 처음부터 익절을 포함하고 있었다.** 넓혀야 하는 것은 조건이 아니라
   > **취소 대상**이다 — 지금 취소하는 것은 브로커의 working order이고, 상주 조건주문은
   > `clearTheSymbol`의 시야 밖에 있다.

3c. **발동한 child 주문의 체결을 엔진이 귀속할 수 없다.** 어댑터가 `TriggeredOrderID`를
   `triggered := strings.TrimSpace(raw.TriggeredOrderID) != ""`로 **bool로 접어 버리고**
   (`protectionofficial/gateway.go:271`) child의 id를 버린다. 체결 감지는 `mutation_attempts`가
   소유한 broker order id만 추적하므로(`journal/fills.go:1538-1548`) 로컬 intent가 없는 child는
   추적 대상이 아니다. **결과: 브로커가 팔았는데 journal은 여전히 보유로 읽고, D2의 terminal
   행이 재등록으로 간다.** 대사가 결국 수량을 고치지만 그 전에 워커 주기가 두 번째 상주 주문을
   만들 수 있다.

   ⇒ **D2의 terminal 행은 「재등록」이 아니라 「보유 재확인 후에만 재등록」이어야 한다.**
   그리고 child id를 살리는 것(`BrokerProtection`에 `TriggeredOrderID`를 싣는 것)이 이 change의
   task가 된다 — 그것 없이는 "브로커가 우리 손절을 실행했다"를 원장이 알 방법이 없다.

4. 포지션이 비-보호 경로로 닫힌 경우(수동 매도, 대사 주도 청산)에도 상주 주문은 남는다.
   수렴 워커가 flat 포지션의 상주 주문을 취소한다(D2의 terminal 행). **다음 주기까지의 창이
   남으므로**, 그 창에서 상주 주문이 발동해도 브로커가 보유 없는 매도를 거부하는 것이
   최종 방어선이다. M13이 수량을 예약하지 않는다고 측정했으므로 **이 창에 새로 매수가 들어오면
   그 주식에 대해 발동할 수 있다** — 레인이 모두 OFF인 현재는 자동 매수가 없어 수동 매수만
   해당하고, 운영 문서에 적는다(tasks 6.4). a105가 진입을 열 때 이 창은 **닫아야 한다.**

**이 계약이 「청구권 하나」를 보장하지 않는다는 것을 적어 둔다.** 2번이 취소 확인 실패에도
매도를 진행시키므로, 그 순간 상주 주문과 인프로세스 매도가 **동시에 유효할 수 있다.** 그것은
버그가 아니라 선택이다 — §0-3이 청산 지연을 금지하므로 취소 확인을 기다리는 쪽이 더 위험하다.
브로커가 보유 없는 두 번째 매도를 거부하는지는 **미확인(UNVERIFIED)**이며 M-A에서 관측한다.

따라서 이 결정이 실제로 주는 것은 「청구권이 항상 하나」가 아니라 **「둘이 되는 창을 알려진
경로마다 최소화하고, 남는 창을 명시한다」**이다. spec의 요구사항도 그렇게 쓴다.

### D9. 상주 보호의 trigger는 `exit_states.baseline_price`다 — 이 판의 전제가 틀렸었다

**이 결정의 이전 판은 "ratchet 수준이 스칼라로 영속되는지 미확인"이라고 적고 상주 trigger를
얼린 `InitialStop`으로 두는 재난 하한(disaster floor)을 채택했다. tasks 0.8이 그 전제를
반증했다.** 영속된다. 그것도 정확히 필요한 성질을 갖고서.

측정한 것:

- `exit_states`는 `initial_stop`과 **별도로** `baseline_price`를 열로 갖고, exit 판정마다
  갱신된다(`internal/journal/exit_state.go:563`
  `UPDATE exit_states SET baseline_price=?,high_water=?,ratchet_level=?,active_rung=?…`).
- t0의 `baseline_price`는 `InitialStop`이다(`exitpolicy/ratchet.go:321-332`
  `OpenRatchetState` → `Baseline: formatPrice(s)`). 즉 하한은 그대로 보존된다.
- LADDER의 `RunnerTrailPct`는 **별도 경로가 아니라 후보로 합성된다** —
  `ladder.go:391-403`이 trail을 `StopCandidate{Name: CandidateHighWaterRunner}`로 만들어
  `ComputeProtectedStop`에 넣고, `:411`이 그 결과를 `out.Baseline`으로 쓴다.
  ⇒ **고점 추적분이 이미 `baseline_price` 안에 있다.**
- 단조성은 이 change가 만들 규칙이 아니라 **이미 있는 불변식**이다 —
  `candidate.go:86-88` "returns max(previous, valid candidates) … and **never a price below
  previous**", `apply_hook.go:559` "Baseline and HighWater are decimal strings, both monotone
  non-decreasing".

**채택: 상주 보호의 trigger는 `exit_states.baseline_price`에서 유도한다.**

- 근사·재계산은 여전히 금지다. 읽는 값이 `initial_stop`에서 `baseline_price`로 바뀔 뿐,
  "영속된 스칼라에서만 유도한다"는 규칙은 그대로다.
- "더 안전한 값으로만 교체한다"는 a100이 **부과할 규칙이 아니라 열의 불변식을 상속하는 것**이다.
  후퇴 거부는 방어적 단언으로 남기고(tasks 5.4), 그 단언이 깨지면 그것은 a100의 결함이 아니라
  exit 경로의 결함이므로 별도로 기록한다.
- **상주 trigger와 엔진 청산선의 차이는 영구 간극이 아니라 수렴 지연이다.** `baseline_price`가
  오르면 다음 수렴 주기가 상주 주문을 교체한다. 표시해야 할 것은 "다를 수 있다"가 아니라
  **마지막 교체 이후 경과와 현재 `baseline_price`와의 차**다(tasks 6.4).

**이 정정이 만든 새 비용:** trigger가 움직이므로 **교체가 자동 경로에 상시 존재한다.** 이전 판의
얼린 하한은 등록·취소만 필요했고 교체는 예외 경로였다. 이제
`POST /api/v1/conditional-orders/{id}/modify`(또는 cancel+create)가 정상 경로이며,
그 엔드포인트는 **verify 기록에 증거가 없다**(D11). D9의 정정이 D11의 비용을 키웠다.

### D10. 이 change가 주장하지 않는 것 — 발동은 아직 미측정이다

`measurements.md:197`이 조건주문의 **발동**을 양 시장 모두 미측정으로 남겨 뒀다. 따라서
이 change는 완료 후에도 다음을 **주장하지 않는다.**

- "엔진이 죽어도 포지션이 보호된다" — 선행 실측 M-A가 통과해야 참이 된다.
- `ProtectionWired`의 docstring이 말하는 상태. a100은 그 값을 생산하지 않는다.

주장할 수 있는 것은 "브로커에 손절 주문이 상주하며 그 존속·정정·취소가 실측됐다"까지다.
M-A가 실패하면(발동하지 않거나 체결되지 않으면) 이 change의 산출물은 보호가 아니므로
**설계가 아니라 전제가 무너지고, 그때는 이 change를 다시 연다.**

### D11. 새 브로커 엔드포인트는 attestation 카탈로그에 들어가야 한다 — 그리고 그 순서가 위험하다

`internal/app/engine/interlock.go`의 `RequiredEndpoints()`는 "the broker calls the engine makes
and therefore the ones a capability attestation has to cover"이고, 주석이 규칙을 명시한다 —
"목록을 확장하는 change는 가드를 함께 갱신한다".

현재 목록 8개에 **조건주문 create/get/modify/cancel도 sellable-quantity도 없다**(`:220-229`).
그런데 a100이 프로덕션에서 새로 호출하는 것이 정확히 그것들이다
(`internal/official/conditional_writes.go`, `protection_reads.go`).

⇒ **갱신하지 않으면** 엔진이 attestation이 덮지 않는 엔드포인트를 호출한다. 이 저장소의 규율상
그것은 결함이다.

⇒ **갱신하면** 그 엔드포인트를 덮지 않는 기존 attestation으로는 interlock이 거부한다. 같은
주석이 그 대가를 이미 적어 뒀다 — 목록에 잘못 넣으면 "missing US entry evidence refuse
reconcile, protection, exit and fill loops for **both markets**". **재발급 없이 배포하면 엔진이
뜨지 않는다.**

**결정: 목록 갱신과 attestation 재발급을 같은 배포 단위로 묶고, 순서를 tasks에 고정한다.**

#### tasks 0.10이 측정한 재발급의 실제 비용

**a071의 Ed25519 서명은 관계없다.** 그것은 protection capability attestation
(`internal/attest/protection_signature.go`)이고, 시동 interlock이 검증하는 것은 서명 없는
capability attestation(`internal/attest/attest.go`, `capability-attestation.json`)이다.
⇒ **C4(서명 도구 부재)는 a100을 막지 않는다.** 대신 다른 것이 막는다.

재발급 경로는 `tossctl soak attest`뿐이고, 그것은 두 출처만 합친다.

| 출처 | 규칙 | 코드 |
| --- | --- | --- |
| soak 기록의 성공한 read | **GET만.** 비-GET이 있으면 발급 전체를 거부 | `soak/attest.go:188-194` |
| verify 기록의 supervised proof | **`soak.LiveOnlyEndpoints()`에 있는 것만.** read는 거부 | `soak/attest.go:255-277`, `cmd/tossctl/soak.go:455-470` |

측정한 현재 상태(2026-08-11):

- **attestation**: 2026-07-30 발급, **2026-08-29 만료**. 싣고 있는 8개는 현재
  `RequiredEndpoints()`와 정확히 같다. supervised는 `POST /orders`, `POST /orders/{id}/cancel`
  2건뿐(2026-07-28).
- **soak 기록**: 최신 cycle이 **2026-08-05**이고 soak는 그 뒤로 돌지 않았다. `MaxRecordAge`는
  **48시간**이므로(`soak/attest.go:57`, `:95-100`) **오늘 `soak attest`는 거부한다.**
  ⇒ 재발급은 `tossctl soak run`을 다시 돌려 기록이 신선해질 때까지 기다리는 일이며,
  연속 3일 streak도 다시 필요하다.
- **verify 기록**(2026-07-31, KR): 조건주문 증거가 **부분적으로 이미 있다.**
  `POST /api/v1/conditional-orders` ok, `GET /api/v1/conditional-orders` ok,
  `GET /api/v1/conditional-orders/{id}` ok, `GET /api/v1/sellable-quantity` ok.
  **없는 것: `DELETE /api/v1/conditional-orders/{id}`와 `POST /api/v1/conditional-orders/{id}/modify`.**
  그 실행은 `conditional-persist` 단계가 `awaiting-restart`로 멈추면서 도달하지 못했다.
  이 증거들은 **2026-08-30에 30일 validity로 만료된다.**

⇒ **읽기와 쓰기의 조달 경로가 다르고, 읽기 쪽이 더 비싸다.**
수렴 워커가 매 주기 부르는 `GET /api/v1/conditional-orders`는 자동 경로의 상시 read이므로
`GET /api/v1/prices`가 목록에 들어간 것과 같은 이유로 목록에 들어가야 한다. 그런데 read는
supervised에서 조달할 수 없다(`acceptSupervised`가 명시적으로 거부한다 — "One supervised success
is not what days of unattended operation prove"). ⇒ **read-only soak 도구에 조건주문 조회 probe를
추가하고 3일 이상 새로 돌려야 한다.** 이것이 a100의 임계 경로에 있는 가장 긴 선행 작업이다.

⇒ **`acceptSupervised`의 주석이 조건주문 제외를 정당화하는 근거("the gate does not require
them")를 a100이 무효화한다.** 그 주석은 a100과 함께 고쳐야 하며, 고치지 않으면 코드가 자기
카탈로그에 대해 거짓을 말한다.

#### tasks 0.10 (a) 구현이 위 처방을 세 곳 정정했다 (2026-08-11)

- **read는 셋이다.** `protectionofficial.Gateway`의 자동 경로가 부르는 read는
  `GET /api/v1/conditional-orders`(`List`·`Cancel` 확인), `GET /api/v1/conditional-orders/{id}`
  (`Get`·`Create`/`Replace` 확인), `GET /api/v1/sellable-quantity`(`Sellable`)다
  (`gateway.go:103,122,173,193,221,246`). 셋 다 probe한다.
- **probe와 요구는 분리한다.** `BuildAttestation`이 싣는 것은
  `Window.SuccessfulEndpoints()`이고(`attest.go:183`) `RequiredEndpoints()`는 `Evaluate`의
  거부 기준일 뿐이다(`:130`). ⇒ **probe만으로 attestation에 실린다.**
  세 목록(`soak.RequiredEndpoints`·`LiveOnlyEndpoints`·`engine.RequiredEndpoints`)의 확장은
  a100 본체가 **(e)에서 한 번에** 한다 —
  `TestSoakAndLiveEndpointsCoverTheEngineInterlock`이 셋의 정합을 강제하므로 같이 움직인다.
  (a)에서 미리 넓히면 **거부가 증거보다 먼저 온다.**
- **(d)는 a100 바이너리로 돌린다.** `supervisedProofs`가 `LiveOnlyEndpoints()`로 미리 거르므로
  (`cmd/tossctl/soak.go:456-465`), 그 목록이 조건주문 mutation을 모르는 바이너리로 attest하면
  **M-A가 만든 DELETE·modify 증거가 조용히 버려진다.** (d)와 (e)는 같은 빌드다.
- **3일은 probe의 시계가 아니다.** `Window`는 streak 창이므로(`summary.go:107,162`) 창 안에서
  1회 성공하면 증명된다. 3일은 credential streak의 몫이고 그것은 tasks 0.12로 이미 돌고 있다.

그리고 배포 단위가 생각보다 무겁다. **soak는 `tossos:local` 컨테이너 안에서 돌고 같은
컨테이너가 엔진이다.** probe 배포 = 이미지 재빌드 + 컨테이너 재시작 = **손절 없는 창**이므로
KST 05:00~09:00에서만 하고, 그 재시작 구간이 곧 M-B의 측정 대상이다(tasks 0.10b).

**부수 발견 — a100과 무관하게 유효하다.** attestation은 2026-08-29에 만료되고 soak는 8월 5일
이후 돌지 않았다. 그때까지 새 soak를 돌리지 않으면 **automation gate를 켠 엔진은 뜨지 않는다.**
a100이 없어도 필요한 일이므로 지금 시작하는 것이 a100의 일정을 사실상 공짜로 만든다.

## Risks / Trade-offs

- **[배선과 동시에 미검증 결함이 나간다]** 13개 분기 중 9개가 미실행이다. → **거부 경로
  RED 테스트 9건을 배선보다 먼저.** tasks의 순서가 강제한다.
- **[수렴 워커가 브로커를 과도하게 호출한다]** 매 주기 전 포지션 조회는 rate limit에 걸릴 수
  있다. → 수렴이 끝난 포지션은 journal 상태로 판정하고 브로커를 다시 묻지 않는다. 주기와
  백오프를 config로 두되 기본값은 보수적으로.
- **[보호 설치 성공과 journal 커밋 사이의 프로세스 손실]** → a071이 계약을 정했다. exact broker ID
  조회로 복구하고, attested idempotency가 증명되지 않으면 재제출하지 않고 `RECONCILE_REQUIRED`로
  고정한다. a100은 구현할 뿐 새로 정하지 않는다.
- **[`protectionlifecycle` ↔ `protection` 도메인 타입 매핑의 손실]** 두 모델은 독립적으로 만들어졌다.
  → 매핑은 순수 함수로 두고 왕복 property 테스트로 덮는다.
- **[이중 매도 권한]** → D8이 계약을 정했다. 계약이 없으면 초과 매도와 오탐 critical이 동시에 난다.
- **[flat 포지션에 상주 주문이 남는 창]** 비-보호 경로로 포지션이 닫히면 다음 수렴 주기까지
  상주 주문이 남는다. M13이 수량 예약 없음을 측정했으므로 그 창에 **수동 매수가 들어오면 그
  주식에 대해 발동할 수 있다.** → 레인이 전부 OFF인 현재는 자동 매수가 없어 노출이 수동 매수로
  한정된다. 운영 문서에 적고(tasks 6.4), **a105가 진입을 열 때 이 창을 닫는 것을 이관 목록에 넣는다.**
- **[상주 trigger가 엔진의 현재 청산선보다 느슨하다]** ratchet가 올라가도 상주 trigger는
  `InitialStop`에 머문다. → D9가 이를 **재난 하한으로 명시**하고 콘솔·운영 문서에 차이를 표시한다.
  숨기면 "브로커에 손절이 있으니 안전하다"는 잘못된 안심이 생긴다.
- **[죽은 코드 1,364줄]** → 정리 change를 **지금** 등록한다(tasks 6.5).
- **[봉인 해제가 선례가 된다]** → 4개 중 1개만 열고, 같은 change에서 import 경계를 세 패키지로
  넓힌다(D6). 순 효과는 봉인 강화다.
- **[손절 즉시성 약화]** §0-3에 정면으로 걸린다. → D3이 수렴을 모든 안전 루프 밖에 두어
  기존 경로의 실패 의미를 바꾸지 않는다. D8-2가 인프로세스 청산을 취소 확인으로 막지 않는다.

## Migration Plan

0. **선행 실측 M-A·M-B**(`measurement-prereq.md`). 사람 승인 후 실행하고 결과를 기록한다.
   M-A가 실패하면 여기서 멈춘다.
1. 거부 경로 RED 테스트 9건을 GREEN으로 만든다. 프로덕션 조립은 바뀌지 않는다.
2. `protectionlifecycle` ↔ `protection` 도메인 매핑과 journal additive-nullable 스키마를 추가한다.
   기존 행은 새 컬럼이 NULL이므로 「보호 미설치」로 읽히고 동작이 바뀌지 않는다.
3. 수렴 워커를 만든다. 조립되지 않으면 아무 일도 일어나지 않는다 — 이 단계에서 프로덕션
   관찰 동작은 그대로다.
4. 봉인 가드를 D6대로 바꾸고(1개 해제 + 경계 3패키지 확장) `gateway.go`에서 조립한다.
5. `httptest` official fixture로 KR·US 각각 등록·수렴·교체·취소·재기동 복구·이중 권한 경합을
   증명한다. 실계좌는 건드리지 않는다.
6. 롤백은 이미 브로커에 상주하는 보호를 취소하지 않는다.

토글 OFF는 upstream 동작과 동일해야 한다(안전 불변식). 1~3단계는 프로덕션 관찰 동작을 바꾸지
않고, 4단계 이후 관찰 가능한 변화는 **브로커에 보호주문이 생기는 것 하나**다. 진입은 네 겹
잠금이 그대로이므로 0건이다.

## Open Questions

`proposal.md`가 남겼던 넷과, 리뷰가 새로 답한 넷.

| # | 질문 | 답 |
| --- | --- | --- |
| 1 | `Wired: true` 조건을 무엇으로 정의하는가 | **a105로 이관.** a100은 `Wired`를 생산하지 않는다 |
| 2 | latch OFF의 주체가 `interlock`인가 `execgw`인가 | `execgw`. 단 a100은 latch를 만들지 않는다(D5) |
| 3 | 부분체결의 수렴 시점과 재시도 계약 | D2 — 이벤트가 아니라 상태 수렴. 한 주기에 포지션당 1회 |
| 4 | journal 스키마 대응이 additive-nullable로 가능한가 | D4 — 가능하며 그것만 허용한다 |
| 5 | 촉발이 기존 보유에 닿는가 | D2 — 닿는다. 그것이 오늘의 주 대상이다 |
| 6 | 보호 왕복을 체결 감지 안에 둘 수 있는가 | D3 — **없다.** `pollLocked` B6·B8이 금지한다 |
| 7 | 인프로세스 매도와 상주 주문의 권한 관계 | D8 — 인프로세스가 상위. 이미 flat이면 시도하지 않는다 |
| 8 | 상주 trigger가 ratchet를 따라가야 하는가 | D9 — 아니다. 재난 하한이며 스칼라 유도만 허용 |
| 9 | 조건주문이 실제로 발동하는가 | **미측정.** 선행 M-A. D10이 주장 범위를 제한한다 |

### 남아 있는 미지수 — 답이 아니라 측정 대상

**수렴 왕복이 주기보다 긴 경우의 겹침.** 앞선 교체가 브로커 응답을 기다리는 중에 다음 주기가
오면 두 수렴이 겹친다. a071이 계약을 정해 뒀다 — 같은 operation key 재제출은 attestation이
broker-side idempotency를 증명할 때만 허용하고, 아니면 `RECONCILE_REQUIRED`로 고정한다. a100은
그 계약을 구현하고 **fixture로 겹침을 실제로 재현해서** 두 번째 수렴이 두 번째 보호주문을
만들지 않음을 증명한다(tasks 5절). 계약은 새로 만들지 않는다.
