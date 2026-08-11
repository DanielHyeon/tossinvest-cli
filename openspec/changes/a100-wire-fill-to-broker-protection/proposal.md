# a100 — 체결은 브로커에 손절을 남긴다

> **상태: AST 근거 확보 (5개 대상).** `analysis/function-logic/`의 산출물이 이 문서가 인용하는
> 모든 분기 주장의 근거다. `python3 tools/logic-map/check_analysis.py --change a100-…` 통과.
> 아직 없는 것은 spec delta와 tasks이며, design 단계에서 `dispatch` 배선 여부를 정한다.

## Why

지금 TossOS의 보호는 **엔진 프로세스가 살아 있는 동안만 유효하다.** 엔진이 죽으면 손절도 사라진다.

이것은 추정이 아니라 코드에 적힌 사실이다.

- `internal/execgw/protection.go`는 `ProtectionWired`에 대해 주석으로
  "Nothing in this build produces this value"라고 명시하고, `ProfileProtection`을
  `ProtectionUnwired` 상수로 고정한다.
- `internal/app/engine/protection_wiring.go`의 `productionProtectionAssemblies`는 KR·US 두 시장
  모두 `Wired: false`와 `"fill-lifecycle-unwired"` 문자열이 섞인 component digest로 조립한다.
- `internal/app/engine/interlock.go`는 `EntryPermitted`를 `Protection == ProtectionWired`로 정한다.

그 결과 두 가지가 동시에 성립한다.

1. **신규 자동 진입이 한 건도 발생할 수 없다.** 노출을 늘리는 주문은 게이트웨이에서 거부된다.
2. **기존 보유는 프로세스 수명에만 의존해 보호된다.** 이것이 더 위험한 쪽이다.

한편 이 일을 할 코드는 **이미 존재하고 테스트까지 되어 있으나 도달할 수 없다.**

| 패키지 | 크기 | 외부 non-test importer |
| --- | --- | --- |
| `internal/protectionofficial` | `gateway.go` 13.8K | **0** |
| `internal/protectionlifecycle` | `lifecycle.go` 18.6K + `state.go` 8.8K | **0** |

미배선은 사고가 아니라 **의도된 봉인**이다. `internal/protection/dormant_test.go`와
`internal/app/engine/a071_security_review_test.go`가 `protectionofficial.New`,
`protection.NewSupervisor`를 **금지 심볼로 검사**한다. 따라서 해제는 그 두 테스트를 같은 change에서
함께 바꾸는 방식으로만 가능하며, 그것이 이 change가 존재하는 이유다.

## What changes (경계만 — 내부 분기는 AST 후 확정)

체결 delta를 받아 브로커에 상주하는 보호주문을 설치하고, 그 설치가 확인될 때만
`ProtectionWired`를 생산하는 **단일 경로**를 만든다.

설계 문서 `docs/TossOS_자동매매_전략_및_기술설계.md` §8.2가 요구하는 순서를 계약으로 삼는다.

1. 체결 delta 수신
2. 체결 수량만큼 stop/conditional 보호 주문 제출
3. 브로커 order id와 상태 수신
4. 보호 수량 합계가 보유 수량과 일치하는지 확인
5. 불일치 시 신규 진입 latch OFF
6. 보호 설치 실패 시 reduce-only 긴급 청산 또는 사람 승인된 fail-safe 실행

## Non-goals

- 레인 활성화(a105). 이 change는 `ProtectionWired`를 **생산 가능하게** 만들 뿐,
  어떤 레인도 켜지 않는다.
- 후보 threshold 승인(a101), 라이브 평가(a103), 사이징 역산(a104).
- 실계좌 주문. 이 change의 검증은 httptest 계약 테스트로 하고, 실계좌 1회는 a106이 담당한다.

## Safety

- **High-risk 경로**다 — 체결 감지, 보호주문, 진입 인터록에 모두 닿는다.
  Pre-Edit 선언과 Function Logic Map이 면제되지 않는다.
- 안전 불변식 §0-3(손절·비상 청산의 즉시성을 약화·지연하지 않는다)에 정면으로 걸린다.
  이 change는 즉시성을 **강화**하는 방향이며, 반대 방향 변경은 이 change 범위에서 금지한다.
- 토글 OFF는 upstream 동작과 동일해야 한다. 보호 배선이 꺼진 상태의 동작은 현재와 같아야 한다.
- 자동 테스트에서 실계좌 주문이 발생해서는 안 된다. 격리된 config 디렉터리와 httptest로 강제한다.

## Open questions (AST 이후 답한다)

1. `productionProtectionAssemblies`가 `Wired: true`를 낼 수 있는 조건을 무엇으로 정의하는가 —
   빌드 시점 digest인가, 런타임 브로커 능력 확인인가.
2. 보호 설치 실패 시 latch OFF의 주체가 `interlock`인가 `execgw` 게이트웨이인가.
3. 부분체결에서 "보호 수량 = 보유 수량" 판정의 시점과 재시도 계약.
4. `protectionlifecycle`의 상태와 journal 원장 스키마의 대응 — additive-nullable로 가능한가.

## Evidence produced

`tools/logic-map` AST 산출물 5건 + 측정된 branch test map.

| 대상 | 분기 | true 결과 실행됨 | 미실행 |
|---|---:|---:|---:|
| `execgw.Gateway.checkProtection` | 5 | 5 | 0 |
| `engine.runInterlock` | 3 | 2 | **1** |
| `engine.productionProtectionAssemblies` | **0** | — | — |
| `protectionlifecycle.applyFill` | 7 | 2 | **5** |
| `protectionlifecycle.prepareRegister` | 6 | 2 | **4** |
| **합계** | **21** | **11** | **10** |

측정은 `go test -covermode=set -coverprofile`로 했고, 분기 *조건*이 아니라 **true 결과 본문의
실행 여부**를 봤다. 조건 statement가 covered인 것은 조건이 평가됐다는 뜻일 뿐이기 때문이다.

### AST가 바꾼 것 — Open question 1의 답

`productionProtectionAssemblies`의 **분기가 0이다.** 조건 없이 `Wired: false`를 반환하는
직선 생성자이며, 따라서 **"언제 WIRED가 되는가"라는 판단 지점이 코드에 존재하지 않는다.**

a100은 이 함수에 조건을 *추가*하는 작업이 아니다. **판단 주체를 새로 만들고 그 결과를
주입**하는 작업이다. `Wired`를 리터럴에서 파라미터로 바꾸면 이 함수는 순수 생성자로 남고
판단은 새 provider가 진다 — 이 형태를 design의 출발점으로 삼는다.

### AST가 드러낸 최대 리스크 — 배선 대상의 거부 경로가 미검증이다

a100이 프로덕션에 배선하려는 두 함수에서 **13개 분기 중 9개가 true 결과를 한 번도 실행하지 않는다.**
호출자가 없었으므로 당연한 결과이지만, 그대로 배선하면 **숨어 있던 결함이 배선과 동시에
프로덕션으로 나간다.**

특히 위험한 셋:

- `applyFill` B3 — broker order id 불일치 검사. 잘못된 체결을 남의 포지션에 귀속시키는 것을
  막는 유일한 방어선인데 미실행이다.
- `applyFill` B7 — 잔량 0 → `Terminal` 전이. **보호주문이 다 채워졌을 때 상태가 올바르게
  닫히는지 확인된 적이 없다.**
- `prepareRegister` B3·B4 — 이미 pending / 이미 active. 같은 포지션에 보호주문이 두 번 나가는 것을
  막는 방어선이며, 재시도·복구 경로에서 가장 먼저 부딪힐 분기다.

**따라서 tasks의 순서는 "거부 경로 RED 테스트 9건 → 배선"이다.** 배선이 먼저가 아니다.

`runInterlock` B3(수락 audit 기록 실패 시 기동 실패)도 미실행이다. 안전 불변식 §0-5의 실패
방향이며, `WIRED`가 생산 가능해지면 "진입이 허용된 사실이 기록되지 못했는데 기동이 계속되는"
상태가 실재하게 되므로 함께 메운다.

### 기존 테스트 전제 변경

`internal/app/engine/guardian_test.go:131-132`와 `interlock_entry_test.go:70-71`은 현재
`EntryPermitted == true`를 **실패로 단언**한다. a100이 `WIRED`를 생산 가능하게 만들면 이 두
단언의 전제가 바뀌므로 같은 change에서 함께 갱신해야 한다.

## Function Logic Map 면제 기록

`internal/app/engine` — `strategyDispatchCycle.dispatch`: **유예(not-applicable, 조건부)**

AST를 생성한 결과 **분기 14 / return 16**으로 이 change의 다른 대상을 모두 합친 것보다 크다.
a100이 이 함수를 편집할지, 그 분기를 근거로 인용할지가 **아직 정해지지 않았다** — 배선 지점을
`dispatch` 바깥에 두는 설계가 가능하기 때문이다. 분기 14개를 먼저 분석하면 설계 결정에 따라
버려질 수 있으므로 scaffold를 제거하고 design 단계로 넘긴다.

**조건**: design이 `dispatch`를 편집하거나 그 내부 분기를 근거로 인용하기로 정하면,
**그 시점에 Function Logic Map을 만들고 tasks 착수 전에 완성한다.** 이 문서와 design은
그때까지 `dispatch`의 내부 분기를 근거로 사용하지 않는다.
