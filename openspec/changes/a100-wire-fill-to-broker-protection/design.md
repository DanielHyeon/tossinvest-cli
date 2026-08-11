## Context

체결이 나면 브로커에 손절이 남아야 한다. 지금은 남지 않는다. 엔진 프로세스가 죽으면 보호도
같이 사라진다.

이 change는 `a071-wire-kr-us-protection-readiness`의 task 3.5를 승계한다. a071은 그 일을
스스로 보류하면서 사유를 적어 뒀다 — "the repository has no production caller from a
journal-committed fill into an exact journal-derived stop/expiry and durable protection
Plan/Register lifecycle" (a071 `review.md`, C1/C2/M7 disposition). **그 없는 caller를 만드는 것이
이 change다.**

a071이 만든 계약은 그대로 쓴다. market-scoped `WIRED|UNWIRED` verdict, paired snapshot,
sealed supervisor binding, Gateway decision boundary는 이 문서에서 다시 설계하지 않는다.
`proposal.md`의 「Supersession」 절이 그 목록이다.

### 지금 상태를 정확히 말하면

배선이 "안 된" 것이 아니라 **provider가 아예 구성되지 않는다.**

`internal/protectionreadiness/production.go:293`:

```go
for _, assembly := range config.SupervisorAssemblies {
    if !validMarket(assembly.Market) || !assembly.Wired || !validDigest(assembly.ComponentDigest) || assemblies[assembly.Market].Market != "" {
        return   // ← initialize()를 여기서 포기한다. configured = false로 남는다.
    }
    assemblies[assembly.Market] = assembly
}
```

`productionProtectionAssemblies`가 두 시장 모두 `Wired: false`를 주므로(`protection_wiring.go:41-42`)
이 루프의 첫 항목에서 초기화가 끝난다. attestation 파일도, trust policy도, market config도
읽히지 않는다. 그래서 오늘의 프로덕션에는 **상태가 두 개** 있다.

| `TOSSOS_PROTECTION_MANIFEST_SHA256` | `globalRefusal` | `Current()` 반환 | 의미 |
| --- | --- | --- | --- |
| 미설정 (현재 운영) | `RefusalMissingEvidence` (`production.go:261`) | `DefaultSnapshot()` (`:139`) | 정상적인 미배선 |
| 설정됨 | `RefusalInvalid` (`:117` 초기값 유지) | `pairedRefusalSnapshot(RefusalInvalid)` (`:141`) | **진단 불가능한 거부** |

두 번째 줄이 문제다. 운영자가 manifest pin을 넣어도 "왜 안 되는지"를 알려주는 값이 나오지
않는다. 원인은 assembly의 `Wired: false`인데 refusal code는 `Invalid`다. a100은 이 경로를
지나가므로 여기서 반드시 부딪힌다.

### AST가 정한 것

`productionProtectionAssemblies`의 **분기는 0이다**(`analysis/function-logic/`). 조건 없이
리터럴을 반환하는 생성자이므로 "언제 WIRED가 되는가"라는 **판단 지점이 코드에 없다.**
a100은 조건을 추가하는 작업이 아니라 판단 주체를 만드는 작업이다.

그리고 배선 대상 13개 분기 중 9개가 true 결과를 한 번도 실행한 적이 없다. 호출자가 없었으니
당연하지만, 그대로 켜면 **숨어 있던 결함이 배선과 동시에 프로덕션으로 나간다.**

## Goals / Non-Goals

### Goals

- journal에 커밋된 체결로부터 정확한 보호 수량·trigger·만료를 유도하고, 브로커에 상주하는
  보호주문을 durable·idempotent하게 설치·교체·취소·복구한다.
- 그 설치가 실제로 확인될 때만 supervisor assembly가 `Wired: true`를 주장한다.
- 보호 수량이 보유 수량에 미달하는 동안 신규 진입을 닫는다.
- 배선 대상 함수의 미실행 거부 분기 9개를 배선 **이전에** RED 테스트로 덮는다.

### Non-Goals

- 레인 활성화(a105). 이 change는 `ProtectionWired`를 **생산 가능하게** 만들 뿐 어떤 레인도
  켜지 않는다. 6개 레인은 `Desired=OFF, Effective=OFF`로 남는다.
- 후보 threshold 승인(a101), 라이브 평가(a103), 사이징 역산(a104).
- 실계좌 주문. 검증은 `httptest` official fixture로 한다. 실계좌 1회는 a106이다.
- attestation 스키마·서명·키 수명·trusted-time floor. a071이 이미 만들었고 소비만 한다.
- `internal/protection.Controller`와 `Repository`의 부활 (아래 D1).

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
의존성이 된다.** 그것은 a071의 H5 disposition이 명시적으로 배제한 것이다 — "production has no
protection SQLite startup dependency. A colliding/failing `protection.db` path cannot stop journal,
Gateway, exit/fill/reconcile safety runtime assembly." 봉인 가드도 `protection.db` 문자열을 금지
심볼로 검사한다(`dormant_test.go:83`).

`domain.go`의 패키지 주석이 같은 말을 이미 하고 있다 — "Production controller assembly remains
dormant until an exact committed-fill lifecycle is implemented."

**결론:** `protection.Controller`/`Repository`는 a045 시대의 선례이며 a071이 `protectionlifecycle`로
대체했다. a100은 세 조각을 조립한다.

```text
protectionlifecycle   순수 상태 전이 (applyFill, prepareRegister, EntryOpen)
        ↕ (D4: journal이 durable store)
internal/protection   broker-neutral 도메인 타입 (Scope, ConditionalBody, BrokerTarget, BrokerProtection)
        ↕
protectionofficial    official API 전송 (Create/Replace/Cancel/Get/List/Sellable)
```

`protectionofficial.Gateway`는 이미 `protection.*` 타입으로 말한다(`gateway.go:37-232`). 따라서
어댑터가 필요한 곳은 `protectionlifecycle` ↔ `protection` 사이 한 군데뿐이고, 이는 순수 매핑이다.

**대가:** `protection.Controller` 824줄과 `Repository` 540줄이 프로덕션에서 영원히 죽은 코드로
남는다. 이 change는 그것을 지우지 않는다 — 삭제는 별도 change의 일이고, `execgw`가 같은 패키지의
도메인 타입을 쓰고 있어서 패키지 자체는 살아 있어야 한다. **다만 「죽었으나 테스트된」 상태를
문서에 남기는 것으로는 부족하므로, a100 완료 시 `Controller`/`Repository` 정리 change를 후속으로
등록한다.**

### D2. `Wired` 판단 주체는 새 supervisor이고, assembly는 그 결과의 전달자다

`productionProtectionAssemblies`를 순수 생성자로 남긴다. `Wired`를 리터럴에서 **파라미터**로
바꾸고, 판단은 새 컴포넌트가 진다.

```go
// 현재
func productionProtectionAssemblies(buildDigest string) []protectionreadiness.SupervisorAssembly

// a100
func productionProtectionAssemblies(buildDigest string, wired map[protectionreadiness.Market]bool) []protectionreadiness.SupervisorAssembly
```

판단 주체는 `internal/protectionsupervisor`(신규 패키지)다. **`internal/protection` 안에 두지
않는다** — 그러면 `protection.NewSupervisor` 금지 심볼과 충돌하고, 그 금지는 유지할 가치가 있다
(D6). supervisor가 `Wired: true`를 주장하는 조건은 셋 다 참일 때다.

1. 해당 시장의 official conditional-order 능력이 fixture로 검증된 계약과 정확히 일치한다.
2. journal에 커밋된 체결로부터 stop/expiry를 유도하는 경로가 조립되어 있다.
3. `protectionofficial.Gateway`가 그 시장에 대해 구성되어 있다.

셋 중 하나라도 아니면 해당 **시장만** `false`다. 시장 독립성은 a071의 계약이므로 유지한다.

**부수적으로 반드시 고칠 것:** Context의 표 두 번째 줄. supervisor가 `false`를 줄 때
`RefusalInvalid` 대신 진단 가능한 refusal code가 나와야 한다. `production.go:293`의 early return을
`globalRefusal`을 설정하고 나가도록 바꾸거나, assembly 검증을 별도 refusal로 분리한다. 이것은
a071 코드를 편집하므로 **Function Logic Map 대상이다.**

### D3. supervisor digest는 build-bound다 — manifest 재서명이 배포 절차에 들어간다

`production.go:187`:

```go
if err != nil || assembly.ComponentDigest != marketConfig.SupervisorDigest {
    continue   // 이 시장은 평가되지 않는다
}
```

`marketConfig.SupervisorDigest`는 **서명된 manifest의 `supervisor_digest` 필드**다
(`production.go:80`). `assembly.ComponentDigest`는

```go
digestProtectionIdentity("protection-supervisor/v1", buildDigest, "KR", "fill-lifecycle-unwired")
```

이고 `buildDigest`는 `version.Current()` + `debug.ReadBuildInfo()`의 `GoVersion`, `Main.Path/Version/Sum`,
그리고 `vcs`, `vcs.revision`, `vcs.time`, `vcs.modified`를 포함한다(`protection_wiring.go:24-37`).

**⇒ 커밋이 바뀌면 digest가 바뀌고, manifest는 무효가 된다. 재빌드마다 재서명이 필요하다.**

세 가지 선택지를 봤다.

- **(a) 그대로 받아들인다 — 채택.** 배포마다 사람이 새 manifest에 서명한다. 안전 불변식 §0-7
  ("운영 토글 flip과 live 검증은 사람이 직접 승인한다")과 정확히 일치한다. 보호 배선은
  **매 배포마다 사람 승인을 요구하는 것이 옳다.**
- (b) `buildDigest`를 component digest에서 뺀다 — 기각. a071 리뷰가 "current build/evidence binding"을
  명시적으로 승인했다. 빼면 롤백된 바이너리가 최신 manifest를 쓸 수 있다.
- (c) manifest에 digest 목록을 허용한다 — 기각. 서명 하나가 여러 빌드를 승인하게 되어 (b)와
  같은 구멍이 열린다.

(a)의 대가는 운영 마찰이다. 배포 절차 문서에 「manifest 재서명」 단계를 추가하고, 재서명이
누락된 배포는 **조용히 UNWIRED로 떨어지므로 그것을 눈에 보이게** 만든다 — 이것이 D2의 진단
가능한 refusal code가 필요한 두 번째 이유다.

또한 `"fill-lifecycle-unwired"` 문자열이 identity에 들어 있다. a100은 이를 배선된 상태를 뜻하는
값으로 바꾸므로(예: `"fill-lifecycle/v1"`) digest는 어차피 바뀐다.

### D4. 보호 상태의 durable store는 journal이며, 스키마 변경은 additive-nullable이다

`protectionlifecycle`은 순수하다 — 상태를 저장하지 않는다. D1이 별도 SQLite를 배제했으므로
남는 곳은 **기존 trading journal** 하나다.

체결이 journal에 커밋된 뒤에만 보호를 계획한다. 순서가 반대면 프로세스 손실 시 "체결은
있었는데 기록이 없는" 구간이 생긴다.

```text
fill 수신 → journal commit → (여기부터 a100) stop/expiry 유도 → Plan → Register → broker
                                                                      ↓
                                                          broker order id를 journal에 커밋
```

스키마는 **additive-nullable만** 허용한다. 기존 컬럼의 의미를 바꾸지 않고, 새 컬럼은 전부
nullable이며, 값이 없는 행은 「보호 미설치」로 읽힌다. 근거: 프로젝트 기억
`tossos-branch-behind-main-schema` — SchemaVersion이 낮은 바이너리가 뜨면 콘솔만 뜨고 엔진이
조용히 죽는다. 롤백된 바이너리가 새 컬럼을 모르더라도 기존 경로가 그대로 동작해야 한다.

역방향도 명시한다. **롤백은 이미 브로커에 상주하는 보호주문을 취소하지 않는다.** 이는 a071
Migration Plan 5번의 계약이며 a100도 지킨다.

### D5. 진입 차단의 주체는 `execgw` Gateway다. `interlock`은 보고만 한다

Open question 2의 답. 코드가 이미 자기 입으로 말하고 있다(`interlock.go:393-395`):

```go
status.Verified = true
// Starting is not permission. The gateway will refuse a raising mutation on
// its own; this is the same fact, reported before anybody tries.
status.EntryPermitted = status.Protection == ProtectionWired
```

`status.Protection`은 `execgw.ProfileProtection` 상수에서 온다(`interlock.go:166`). a071이
"`ProfileProtection=UNWIRED` remains reporting-only"라고 못 박은 그 값이다. 따라서
`interlock.EntryPermitted`는 **표시값이고 권한이 아니다.** 실제 거부는 노출을 늘리는 mutation마다
`execgw.Gateway.checkProtection`이 한다.

그런데 latch가 **두 종류**라는 점이 중요하다.

| latch | 범위 | 소유자 | 소비 지점 |
| --- | --- | --- | --- |
| readiness | 시장(KR/US) | `protectionreadiness` 스냅샷 | `checkProtection` → `ReadinessRequest{Market, OrderType, Quantity}` |
| coverage | 포지션 | `protectionlifecycle.EntryOpen` | **a100이 정해야 함** |

`ReadinessRequest`에는 포지션 식별자가 없다(`readiness_adapter.go:63-67`). 그러므로 coverage latch를
readiness 경로에 밀어 넣으면 a071이 봉인한 계약을 뜯어야 한다. **그렇게 하지 않는다.**

coverage latch는 **신규 진입 제안 지점**에서 소비한다 — `internal/app/engine/strategy_entry_supervisor.go`.
coverage는 임의 mutation이 아니라 "새 진입을 제안해도 되는가"에 대한 답이기 때문이다. 두 latch는
독립적으로 AND 된다.

### D6. 봉인 가드는 최소로 연다 — 금지 심볼 4개 중 1개만

봉인은 사고가 아니라 의도다. 그러므로 해제도 최소여야 하고, 무엇을 열었는지 정확히 적어야 한다.

`internal/protection/dormant_test.go:83`과 `internal/app/engine/a071_security_review_test.go:28`이
같은 4개를 금지한다.

| 금지 심볼 | a100의 처리 | 이유 |
| --- | --- | --- |
| `protectionofficial.New` | **해제** | 브로커 전송을 조립하려면 반드시 필요하다 |
| `protection.NewSupervisor` | **유지** | supervisor를 `internal/protectionsupervisor`에 두므로 이 금지는 그대로 유효하다 (D2) |
| `protection.db` | **유지** | D1이 별도 SQLite를 배제했으므로 계속 금지된다 |
| `GatewayFactory` | **유지** | 임의 factory 주입은 여전히 두 번째 mutation 경로다 |

해제하는 하나에는 **대체 단언**을 붙인다. `protectionofficial.New`가 gateway.go에 나타나는 것을
허용하되, 같은 테스트가 (1) 조립된 supervisor가 journal-backed이고 DB-free인지, (2) 조립이
`gateway.go` 단 한 곳인지를 계속 검사한다.

「app 코드 중 `internal/protection`을 import할 수 있는 파일은 `gateway.go` 하나」라는 규칙
(`dormant_test.go:61`)은 **바꾸지 않는다.** 이것이 "no second mutation path"의 실체이며, a100의
조립도 `gateway.go` 안에서 한다.

함께 바꿔야 하는 기존 단언 세 곳을 미리 적는다. 숨은 채로 깨지면 안 되기 때문이다.

- `a071_security_review_test.go:9` `TestUnprovenFillLifecycleKeepsBothProductionAssembliesUnwired` —
  `!assembly.Wired`를 무조건 단언한다. supervisor 입력에 따른 조건부로 바꾼다.
- `guardian_test.go:131-132`, `interlock_entry_test.go:70-71` — `EntryPermitted == true`를 실패로
  단언한다. `WIRED`가 생산 가능해지면 전제가 바뀐다.

### D7. `dispatch`는 건드리지 않는다 — FLM 면제 확정

`proposal.md`의 조건부 면제를 여기서 확정한다. `strategyDispatchCycle.dispatch`(분기 14 / return 16)는
**편집하지 않고, 그 내부 분기를 근거로 인용하지도 않는다.**

D4가 배선 지점을 「fill이 journal에 커밋된 이후」로 정했고, D5가 coverage latch를 entry supervisor에
두었으므로 `dispatch` 내부에 진입할 이유가 없다. 따라서 `Function Logic Map: not-applicable —
편집하지 않고 내부 분기를 근거로 쓰지 않음`이다.

**이 면제는 조건부로 남는다.** 구현 중 `dispatch`를 편집하거나 그 분기를 근거로 인용하게 되면
**그 시점에 AST를 만들고 tasks 착수 전에 완성한다.**

## Risks / Trade-offs

- **[배선과 동시에 미검증 결함이 나간다]** 13개 분기 중 9개가 미실행이다. 특히 `applyFill` B3
  (broker order id 불일치 — 잘못된 체결을 남의 포지션에 귀속시키는 것을 막는 유일한 방어선),
  B7(잔량 0 → `Terminal`), `prepareRegister` B3·B4(중복 등록 방지). → **거부 경로 RED 테스트 9건을
  배선보다 먼저** 수행한다. tasks의 순서가 이것을 강제한다.
- **[재빌드마다 manifest 재서명 누락]** → D3(a)를 채택한 대가. 누락이 조용한 UNWIRED가 되지
  않도록 D2의 진단 가능한 refusal code로 드러낸다.
- **[보호 설치 성공과 journal 커밋 사이의 프로세스 손실]** → a071이 이미 계약을 정했다. exact
  broker ID 조회로 복구하고, attested idempotency가 증명되지 않으면 재제출하지 않고
  `RECONCILE_REQUIRED`로 고정한다. a100은 그 계약을 구현할 뿐 새로 정하지 않는다.
- **[`protectionlifecycle` ↔ `protection` 도메인 타입 매핑의 손실]** 두 모델은 독립적으로 만들어졌다.
  매핑이 정보를 잃으면 lifecycle이 증명한 불변식이 전송 단계에서 깨진다. → 매핑은 순수 함수로
  두고 왕복(round-trip) property 테스트로 덮는다.
- **[죽은 코드 1,364줄이 남는다]** `Controller` + `Repository`. → 이 change에서 지우지 않고 후속
  정리 change로 등록한다. 지금 지우면 a100의 변경 집합이 리뷰 불가능하게 커진다.
- **[봉인 해제가 선례가 된다]** 한 번 열면 다음이 쉬워진다. → 4개 중 1개만 열고, 나머지 3개가
  왜 유지되는지를 D6의 표에 남긴다. 해제된 하나에는 대체 단언을 붙인다.
- **[손절 즉시성 약화]** 안전 불변식 §0-3에 정면으로 걸린다. → 이 change는 즉시성을 **강화**하는
  방향이며(엔진 사망 후에도 보호가 남는다), 반대 방향 변경은 이 change 범위에서 금지한다.
  기존 stop·긴급청산·대사·체결 경로는 readiness provider를 읽지 않는다는 a071의 정적 격리를
  그대로 유지한다.

## Migration Plan

1. 거부 경로 RED 테스트 9건 + `runInterlock` B3을 먼저 GREEN으로 만든다. 이 단계에서 프로덕션
   조립은 바뀌지 않는다.
2. `protectionlifecycle` ↔ `protection` 도메인 매핑과 journal additive-nullable 스키마를 추가한다.
   기존 설치는 새 컬럼이 NULL이므로 「보호 미설치」로 읽히고 동작이 바뀌지 않는다.
3. `internal/protectionsupervisor`와 `productionProtectionAssemblies`의 파라미터화를 추가한다.
   supervisor 입력이 없으면 두 시장 모두 `false`이므로 **기본 동작은 현재와 동일하다.**
4. 봉인 가드 3곳을 D6대로 바꾸고 `gateway.go`에서 조립한다. manifest pin이 없으면 여전히
   `DefaultSnapshot()`이다.
5. `httptest` official fixture로 KR·US 각각 등록·부분체결·수량 수렴·교체·취소·재기동 복구를
   증명한다. 실계좌는 건드리지 않는다.
6. 롤백은 신규 진입을 OFF로 유지하고 **이미 브로커에 상주하는 보호를 취소하지 않는다.**

토글 OFF는 upstream 동작과 동일해야 한다(안전 불변식). 1~3단계는 프로덕션 관찰 동작을 바꾸지
않고, 4단계 이후에도 manifest pin과 서명된 attestation이 설치되기 전까지는 현재와 같다.

## Open Questions

`proposal.md`가 남긴 넷 중 셋이 답했다.

| # | 질문 | 답 |
| --- | --- | --- |
| 1 | `Wired: true` 조건을 무엇으로 정의하는가 | D2 — 새 supervisor가 판단하고 assembly는 전달만 한다 |
| 2 | latch OFF의 주체가 `interlock`인가 `execgw`인가 | D5 — `execgw`. `interlock`은 보고. 단 coverage latch는 entry supervisor |
| 4 | journal 스키마 대응이 additive-nullable로 가능한가 | D4 — 가능하며 그것만 허용한다 |

남은 하나.

**3. 부분체결에서 「보호 수량 = 보유 수량」 판정의 시점과 재시도 계약.**

`protectionlifecycle.applyFill`이 매 전이마다 등식을 재평가한다는 것은 안다:

```go
position.EntryOpen = position.Phase == Active && position.EntryLatch == "" &&
    next.marketLatches[key.Market] == "" && position.Observed.Status == BrokerActive &&
    position.Observed.Quantity+position.OtherSellClaims == position.Holdings
```

모르는 것은 **연속 부분체결 중 브로커 왕복 횟수**다. 체결 3건이 연달아 들어오면 교체를 3번
할 것인가, 합쳐서 1번 할 것인가. 3번은 rate budget과 무보호 window를 늘리고, 1번은 마지막
체결까지 보호 부족 상태를 연장한다.

이 질문은 **실제 official API의 부분체결 이벤트 도착 간격을 모르면 답할 수 없다.** 추정으로
정하지 않는다. tasks 착수 전에 official fixture의 체결 이벤트 타이밍 계약을 먼저 확인하고,
그때까지는 **보수적 기본값(체결 1건당 즉시 1회 수렴)**을 전제로 설계한다. 보수적이라는 근거는
안전 불변식 §0-6("손절·익절·사이징 변경은 명확한 근거가 있는 보수 방향만 허용한다")이며,
무보호 window를 짧게 두는 쪽이 보수적이다.
