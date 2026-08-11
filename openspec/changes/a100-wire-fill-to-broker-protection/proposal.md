# a100 — 체결은 브로커에 손절을 남긴다

> **상태: design 작성 완료.** `analysis/function-logic/`의 AST 산출물이 이 문서가 인용하는 모든
> 분기 주장의 근거다. `python3 tools/logic-map/check_analysis.py --change a100-…` 통과.
> `design.md`가 open question 1·2·4를 답했고 `dispatch` 면제를 확정했다(D7). 아직 없는 것은
> spec delta와 tasks이며, open question 3은 official fixture 확인 후 답한다.

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

## Supersession — a071 task 3.5

이 change는 신규 발명이 아니다. **`a071-wire-kr-us-protection-readiness`의 task 3.5를 승계한다**
(2026-08-11 결정). a071 tasks.md §6과 status.md에 대응 기록이 있다.

a071은 21/25 완료 상태에서 3.5를 스스로 보류했고, review.md가 사유를 적어 뒀다 — "the repository
has no production caller from a journal-committed fill into an exact journal-derived stop/expiry and
durable protection Plan/Register lifecycle". **그 선행 조건이 이 change의 주제다.**

**a071에서 이미 확정됐고 이 change가 재설계하지 않는 것:**

- readiness는 market-scoped `WIRED|UNWIRED` + typed refusal이다. 결합 상태를 만들지 않는다.
- `WIRED`는 signed attestation과 sealed supervisor binding이 **둘 다** 검증될 때만 생긴다.
- reduce-only SELL/CANCEL/축소 AMEND, stop·긴급청산·대사·체결 경로는 readiness provider를
  읽지 않는다. 정적 격리 테스트가 이를 강제하며 이 change도 깨지 않는다.
- 공개 scalar override나 exported readiness 필드로 entry를 승인할 수 없다.
- `internal/protection`의 import 경계와 `gateway.go` 단일 진입 규칙(`dormant_test.go`).

**이 change가 새로 지는 것:** journal-committed fill → 정확한 stop/expiry 유도 → durable
Plan/Register lifecycle → supervisor assembly의 `Wired` 판단, 그리고 그 경로의 거부 분기 검증.

**되돌리는 조건:** 이 change가 취소되거나 프로덕션 배선이 범위에서 빠지면 3.5는 a071로 돌아간다.
그때 이 절과 a071 tasks.md §6·review.md addendum을 함께 지운다.

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

## Open questions

| # | 질문 | 상태 |
| --- | --- | --- |
| 1 | `productionProtectionAssemblies`가 `Wired: true`를 낼 조건 | **답함** — `design.md` D2 |
| 2 | latch OFF의 주체가 `interlock`인가 `execgw`인가 | **답함** — `design.md` D5 |
| 3 | 부분체결에서 "보호 수량 = 보유 수량" 판정 시점과 재시도 계약 | **미해결** — official fixture의 체결 이벤트 타이밍 확인이 선행. 보수적 기본값(체결 1건당 즉시 1회 수렴)을 전제로 설계 |
| 4 | `protectionlifecycle` 상태 ↔ journal 스키마 대응 | **답함** — `design.md` D4 |

`design.md`가 추가로 정한 것: 프로덕션 상태 core는 `protectionlifecycle`이며
`protection.Controller`/`Repository`는 쓰지 않는다(D1 — 별도 SQLite가 기동 의존성이 되어 a071 H5
disposition과 `protection.db` 금지에 걸린다). 봉인 가드는 금지 심볼 4개 중 `protectionofficial.New`
하나만 연다(D6).

## Evidence produced

`tools/logic-map` AST 산출물 5건 + 측정된 branch test map.

| 대상 | 분기 | true 결과 실행됨 | 미실행 |
| --- | ---: | ---: | ---: |
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

`internal/app/engine` — `strategyDispatchCycle.dispatch`: **not-applicable (조건부)**

AST를 생성한 결과 **분기 14 / return 16**으로 이 change의 다른 대상을 모두 합친 것보다 크다.
design 단계에서 배선 지점을 `dispatch` 바깥으로 확정했으므로(`design.md` D7) 이 함수를 편집하지
않고 그 내부 분기를 근거로 인용하지도 않는다. 근거: D4가 배선 지점을 「fill이 journal에 커밋된
이후」로, D5가 coverage latch를 entry supervisor로 정했다.

**조건**: 구현 중 `dispatch`를 편집하거나 그 내부 분기를 근거로 인용하게 되면 **그 시점에
Function Logic Map을 만들고 tasks 착수 전에 완성한다.**
