# a100 — 보유는 브로커에 손절을 남긴다

> **상태: 2026-08-15 current-main 재동결 중.** 2026-08-11의 범위 축소는 유지하지만 실행 base,
> runtime 격리, raw status와 child causal ownership 경계를 다시 동결한다. 다른 change로 옮기지
> 않는다. M-A 증거를 가능하게 하는 measurement-only M0를 먼저 닫고, M-A 완전 PASS와 0.11
> raw-status 표 전에는 제품 Terra 구현 로트를 열지 않는다.
> 리뷰 기록은 `review.md`다. 범위 축소의 근거는 `execgw.Gateway.checkProtection`의 reduce-only
> 단락이며, 이관 대상과 사유는 아래 「Non-goals」와 `review.md`에 있다.
> **디렉터리 이름은 `a100-wire-fill-to-broker-protection`으로 유지한다** — `base-commit.txt`,
> `analysis/` 경로, `STORY-TOS-a100.yaml`, PM registry가 이 ID에 걸려 있고, 이름 변경의 비용이
> 정확성에 주는 이득보다 크다. 제목만 새 범위를 말한다.

## Why

지금 TossOS의 보호는 **엔진 프로세스가 살아 있는 동안만 유효하다.** 엔진이 죽으면 손절도 사라진다.

이것은 추정이 아니라 코드에 적힌 사실이다.

- `internal/execgw/protection.go:55-62`는 `ProtectionWired`를 "a position survives the engine
  dying"으로 정의하면서 **"Nothing in this build produces this value"**라고 적는다.
  `ProfileProtection`은 `ProtectionUnwired` 상수로 고정돼 있다.
- `internal/app/engine/protection_wiring.go`의 `productionProtectionAssemblies`는 KR·US 두 시장
  모두 `Wired: false`와 `"fill-lifecycle-unwired"` 문자열이 섞인 component digest로 조립한다.

실제로 일어난 적도 있다. a056(2026-08-02)에서 엔진이 8분간 기동하지 못했고 그동안 OPEN 5건과
exit 정책 4건이 무감시로 남았다. 두 시장이 모두 닫혀 있어 손실이 없었을 뿐이다.

### 오늘 위험한 것은 기존 보유다

6개 레인이 모두 `Desired=OFF, Effective=OFF`이고 진입은 게이트웨이에서 전면 거부된다. 따라서
**신규 체결은 발생하지 않는다.** 촉발 조건을 「체결 delta > 0」으로 두면 이 change는 오늘 계좌에
**손절을 한 건도 남기지 않는다.**

보호가 필요한 것은 이미 계좌에 있는 보유분이다. 그중 **exit 관리로 편입된 것**은
`InitialStop`(= adoption의 `SyntheticStop`)을 journal에 갖고 있으므로
(`internal/journal/adoption.go:289`) 없는 것은 **그 stop을 브로커에 남기는 경로** 하나다.

**"모든 보유"가 아니다** (adversarial Eng 리뷰 정정). 편입은 토글이고 **기본값이 off**다
(`internal/config/engine.go:79-90` — "the zero value is the safe one. With `enabled` false the
engine behaves exactly as it did before this change"). 제외 목록에 있거나 편입되지 않은 보유는
`adoption.go`가 **unmanaged로 알림만 낸다.** 그리고 대사 루프 자체가 automation gate가
verified가 아니면 존재를 거부한다(`app/engine/reconcileloop.go:342`).

⇒ a100이 오늘 보호하는 것은 **exit 관리 상태가 열린 포지션**이며, 그 집합의 크기는 운영 설정에
달렸다. 2026-08-11 현재 이 배포의 `adoption.enabled`는 **true**다(tasks 0.9가 착수 시점에
다시 확인한다). 편입되지 않은 보유는 이 change의 대상이 아니고 **그 사실을 감추지 않는다.**

⇒ **촉발은 이벤트가 아니라 상태다.** 「보호가 설치되지 않은 포지션」이 있으면 수렴한다.
체결은 그 상태를 만드는 여러 원인 중 하나일 뿐이다.

### 그리고 이 일에는 `Wired`가 필요 없다

`internal/execgw/protection.go:89-91`:

```go
func (g *Gateway) checkProtection(ctx context.Context, plan mutationPlan, previous protectionCheckpoint) (protectionCheckpoint, *RejectedError) {
	if !plan.raisesExposure {
		return protectionCheckpoint{}, nil
	}
```

`raisesExposure`는 신규 주문에서 `strings.EqualFold(intent.Side, "buy")`(`gateway.go:377`),
취소에서 `false`(`:416`), 정정에서 수량·가격 증가 여부(`:447`)다. **보호주문은 매도이므로 이
함수는 readiness를 아예 조회하지 않는다.**

⇒ 브로커에 손절을 남기는 데 supervisor·`Wired` 생산·서명된 manifest·coverage latch가
**하나도 필요하지 않다.** 그것들은 오직 진입을 열기 위해 존재하고, 진입에는 그 밖에도 세 개의
잠금(레인 OFF·threshold 미승인·서명 manifest 부재)이 더 걸려 있다. 이 change는 진입을 열지 않는다.

### 이 일을 할 코드는 이미 존재하고 도달할 수 없다

| 패키지 | 크기 | 외부 non-test importer |
| --- | --- | --- |
| `internal/protectionofficial` | `gateway.go` 13.8K | **0** |
| `internal/protectionlifecycle` | `lifecycle.go` 18.6K + `state.go` 8.8K | **0** |

미배선은 사고가 아니라 **의도된 봉인**이다. `internal/protection/dormant_test.go`가
`protectionofficial.New`, `protection.NewSupervisor`, `protection.db`, `GatewayFactory`를
금지 심볼로 검사한다. 해제는 그 테스트를 같은 change에서 함께 바꾸는 방식으로만 가능하며,
그것이 이 change가 존재하는 이유다.

## What changes

**보호가 설치되지 않은 포지션을 브로커 상주 보호주문으로 수렴시키는 단일 경로**를 만든다.
기존 보유와 신규 체결을 구별하지 않는다 — 둘 다 「journal에 커밋된 상태에서 유도한 stop을
가진, 브로커 보호가 없는 포지션」이다.

1. journal에 커밋된 포지션 상태에서 보호 수량·trigger·만료를 정확히 유도한다.
2. 브로커에 상주 보호주문을 durable·idempotent하게 등록한다.
3. 브로커 order id와 상태를 journal에 커밋한다(additive-nullable 컬럼).
4. 보유 수량이 바뀌면 더 안전한 방향으로만 교체한다.
5. 수렴 실패는 typed reconcile reason과 관측 가능한 알림으로 끝난다 — 체결 감지·청산·대사를
   막지 않는다.
6. 한 포지션에 브로커측 매도 청구권이 둘이 되지 않도록 인프로세스 보호 매도(a087)와의
   권한 계약을 정한다.
7. raw conditional status와 triggered child id를 손실 없이 보존하고, child fill보다 먼저
   exact parent/client/scope/generation owner를 journal에 등록한다. 이미 관측된 child fill은 소급
   귀속하지 않고 `ATTRIBUTION_FAILED` reconcile/alert로 fail closed한다.

**`ProtectionWired`는 이 change 이후에도 계속 `false`다.** a071이 만든 "구조적으로 UNWIRED"
보장은 그대로 유지되고, 진입은 여전히 전면 차단이다.

## 선행 조건 — 착수 전 실측 2건

이 change의 전제는 **한 번도 관측된 적이 없다.**
`openspec/changes/verify-execution-capability/measurements.md:197`이 스스로 적어 뒀다.

> 2c의 기본 가설(SINGLE+MARKET 손절)이 실제로 **발동해 체결되는지는 양 시장 모두 미측정**이며,
> 이것이 2.5의 가장 큰 구멍이다.

등록·조회·존속·정정·취소는 실측됐고 **발동만 deferred**다. 발동하지 않는 상주 주문은 보호가
아니라 보호의 외관이다. 따라서 구현 착수 전에 둘을 측정한다. 자세한 절차는
`measurement-prereq.md`에 있고, **실계좌 주문을 포함하므로 실행 시점에 사람이 별도로 승인한다**
(안전 불변식 §0-1, §0-7). 이전의 「소액 라이브로 진행」을 이 주문에 대한 승인으로 쓰지 않는다.

| # | 측정 | 왜 착수 전인가 |
| --- | --- | --- |
| M0 | parent/child official GET을 한 프로세스 monotonic receipt로 기록하는 측정 도구 | 현 CLI의 wall-clock verify record와 외부 wrapper는 child causal order를 증명하지 못한다 |
| M-A | 조건주문 1건이 실제로 발동·체결되고 parent child-id local receipt가 child fill local receipt보다 먼저인지 (KR 우선, 최소 수량) | 발동하지 않으면 산출물은 보호가 아니다. causal owner를 fill 전에 만들 수 없으면 journal 계약을 구현할 수 없다 |
| M-B | 엔진 사망 → 재기동 → exit observer 재무장까지의 무보호 창 | **2026-08-11 측정 완료(정상 재기동 ≤6.7초).** 정상 기동 하한이며 a056 실패 창을 대체하지 않는다 |

M-A의 부분 통과와 trigger/수량 비교 실패는 구현 착수 실패다. `PAUSED`의 부재도 무장 증명이
아니다. M-A는 raw status와 두 local receipt의 causal order를 동결하며 완전 PASS 외에는 T1로 가지 않는다.

2026-08-15 read-only preflight는 **HOLD**였다. 토요일 장외, 서로 다른 auth/soak authority,
기존 `PAUSED` 2건·`WATCHING` 3건, fresh 비관리 1주 후보 부재보다 더 근본적인 차단은 현 도구가
HTTP response receipt의 process-local monotonic order와 fsync barrier를 남기지 않는다는 점이다.
따라서 승인 자체를 주문 실행 권위로 쓰지 않고, M0 GREEN·독립 리뷰 뒤 다음 KR 세션 직전 fresh
preflight와 exact human confirmation을 다시 받아야 한다. M0는 transport body-read 경계와 모든 HTTP
attempt를 직접 receipt화하고, create 직후 exact cleanup checkpoint를 기존 owner-only verify record에
fsync한다. create-response crash window는 broker call 전 fsync한 client-ID pending intent와 다음 resume의
official all-page unique reconciliation로 닫는다. M-A는 `--redo conditional-trigger` 단독 mode이며 prior
outstanding가 있으면 cleanup prologue도 실행하지 않으므로 idempotency replay/과거 cleanup과 섞이지 않는다.

## Supersession — a071 task 3.5 (분할)

a071 task 3.5의 원문은 "Wire the official protection gateway **and supervisor** into production
engine assembly"다. 리뷰 결과 이 change는 그중 **gateway 배선만** 승계한다.

| 3.5의 부분 | 승계자 | 근거 |
| --- | --- | --- |
| official protection **gateway**를 프로덕션에 배선 | **a100 (이 change)** | 보호 설치는 readiness를 읽지 않으므로 단독으로 성립한다 |
| **supervisor**와 `Wired` 생산 | **a105 (레인 활성화 권위)** | `Wired`의 유일한 소비자는 진입 인터록이고, 진입을 여는 change가 a105다 |

a071이 3.5를 보류한 사유는 "the repository has no production caller from a journal-committed fill
into an exact journal-derived stop/expiry and durable protection Plan/Register lifecycle"였다.
**그 없는 caller가 이 change의 주제이고, supervisor는 그 caller가 아니다.**

**a071에서 이미 확정됐고 이 change가 재설계하지 않는 것:**

- readiness는 market-scoped `WIRED|UNWIRED` + typed refusal이다. 결합 상태를 만들지 않는다.
- `WIRED`는 signed attestation과 sealed supervisor binding이 **둘 다** 검증될 때만 생긴다.
- reduce-only SELL/CANCEL/축소 AMEND, stop·긴급청산·대사·체결 경로는 readiness provider를
  읽지 않는다. 정적 격리 테스트가 이를 강제하며 이 change도 깨지 않는다.
- 공개 scalar override나 exported readiness 필드로 entry를 승인할 수 없다.
- `internal/protection`의 import 경계와 `gateway.go` 단일 진입 규칙(`dormant_test.go`).

**되돌리는 조건:** 이 change가 취소되면 gateway 배선분은 a071로 돌아간다. 그때 이 절과
a071 tasks.md §6을 함께 고친다.

## Non-goals

**a105로 이관한 것** (원안 a100의 범위였고 리뷰가 잘라냈다):

- `internal/protectionsupervisor` 신규 패키지와 시장별 `Wired` 판정
- `productionProtectionAssemblies`의 `wired` 파라미터화, identity 문자열 교체
- `ProductionProvider.initialize`의 refusal 분화, manifest digest 불일치 진단
- 배포 절차의 manifest 재서명, 서명 도구(현재 저장소에 서명자가 **없다** —
  `internal/attest/protection_signature.go`는 검증 전용이고 서명 함수는 `_test.go`에만 있다)
- 포지션 단위 coverage latch와 그 entry supervisor 소비
- `engine.runInterlock` B3(수락 audit 실패 시 기동 실패) — `WIRED`가 생산 가능해야 도달한다
- `a071_security_review_test.go`의 무조건 단언 반전, `guardian_test.go:131-132` ·
  `interlock_entry_test.go:70-71`의 `EntryPermitted` 전제 변경

**원래부터 범위 밖:**

- 레인 활성화(a105), threshold 승인(a101), 라이브 평가(a103), 사이징 역산(a104).
- 실계좌 **주문 전략**. 이 change의 기능 검증은 httptest 계약 테스트로 하고, 전략적 실계좌
  1회는 a106이다. 선행 실측 M-A는 기능 검증이 아니라 **전제 확인**이며 별도 승인 대상이다.
- `internal/protection.Controller`와 `Repository`의 부활(design D1).
- `internal/filldetect` 편집. 조립 지점만 바꾼다.

## Safety

- **High-risk 경로**다 — 보호주문·체결 경로·대사에 닿는다. Pre-Edit 선언과 Function Logic Map이
  면제되지 않는다.
- 안전 불변식 §0-3(손절·비상 청산의 즉시성을 약화하거나 지연하지 않는다)에 정면으로 걸린다.
  이 change는 즉시성을 **강화**하는 방향이며, 반대 방향 변경은 이 change 범위에서 금지한다.
- 토글 OFF는 upstream 동작과 동일해야 한다. 보호 배선이 조립되지 않은 상태의 동작은 현재와 같다.
- 자동 테스트에서 실계좌 주문이 발생해서는 안 된다. 격리된 config 디렉터리와 httptest로 강제한다.
- 선행 실측 M-A는 실계좌 주문 1건을 포함한다. **사람이 그 시점에 직접 승인한다.**

## Evidence produced

아래 표는 2026-08-11 historical freeze다. current-main `882a0b49`의 권위는
`analysis/current-main-evidence.md`, current AST/FLM/BTM과 `pre-edit-gate.md`다. 추가 경계는
`engineRuntime`, auxiliary stop seam, `Gateway.adapt`, `TrackedFillOrders`, `confirmedFillOwners`,
`resolveFillOrigin`, lifecycle public API guard다.

`tools/logic-map` AST 산출물 6건 + 측정된 branch test map.

| 대상 | 분기 | true 결과 실행됨 | 미실행 | a100의 처리 |
| --- | ---: | ---: | ---: | --- |
| `execgw.Gateway.checkProtection` | 5 | 5 | 0 | 인용만 (범위 축소의 근거) |
| `filldetect.Detector.pollLocked` | 10 | 5 | **5** | 인용만 (D8이 이 경로를 피한다) |
| `protectionlifecycle.applyFill` | 7 | 2 | **5** | **RED 대상** |
| `protectionlifecycle.prepareRegister` | 6 | 2 | **4** | **RED 대상** |
| `engine.runInterlock` | 3 | 2 | 1 | **a105로 이관** |
| `engine.productionProtectionAssemblies` | **0** | — | — | **a105로 이관** |

측정은 `go test -covermode=set -coverprofile`로 했고, 분기 *조건*이 아니라 **true 결과 본문의
실행 여부**를 봤다. 조건 statement가 covered인 것은 조건이 평가됐다는 뜻일 뿐이다.

### AST가 정정한 것 — 이 문서가 틀렸던 곳

**리뷰 이전 판의 이 문서는 "`analysis/`의 AST 산출물이 이 문서가 인용하는 **모든** 분기 주장의
근거다"라고 적었으나, 사실이 아니었다.** design.md의 Context 절 상태표는
`ProductionProvider.initialize`의 early return에서, D8-(5)는 체결 감지 루프의 에러 분기에서
유도했는데 둘 다 산출물이 없었다. 손으로 읽은 인용이었다.

그 결과가 무엇이었는지도 기록한다. 산출물을 실제로 만들자 **D8이 인용한 함수 이름이
틀렸다는 것이 드러났다.** `Detector.PollOnce`는 L277-281의 5줄 래퍼이고 **분기가 0**이다.
인용된 로직은 전부 `Detector.pollLocked`(L283-357, 분기 10)에 있다.

측정은 더 나아가 D8을 뒤집었다. `pollLocked` B6(`Ledger.Apply` 에러 → outage + **같은 사이클의
남은 스냅샷 폐기**)은 **한 번도 실행된 적이 없고**, D8의 데코레이터는 바로 그 분기를
프로덕션에서 처음 실행되게 만드는 배치였다. 게다가 B8의 신선도 표본은 `Apply` 왕복이 길어지면
같은 사이클 뒤 스냅샷에서 오염되어 `evaluateSLO` → `Gate.Block(ReasonFillDetectionSLO)`로 간다.
원안 tasks 4.6은 에러만 격리했고 지연은 다루지 않았다 — **보호를 설치하려다 체결 감지를 잃는
경로가 열려 있었다.**

이 문서가 그 규칙을 어긴 기록은 `review.md`에 남긴다. 지금 판의 인용은 전부 산출물이 있다.

### 배선 대상의 거부 경로가 미검증이다

a100이 프로덕션에 배선하려는 두 함수에서 **13개 분기 중 9개가 true 결과를 한 번도 실행하지
않는다.** 호출자가 없었으므로 당연한 결과이지만, 그대로 배선하면 **숨어 있던 결함이 배선과
동시에 프로덕션으로 나간다.**

특히 위험한 셋:

- `applyFill` B3 — broker order id 불일치 검사. 잘못된 체결을 남의 포지션에 귀속시키는 것을
  막는 유일한 방어선인데 미실행이다.
- `applyFill` B7 — 잔량 0 → `Terminal` 전이. **보호주문이 다 채워졌을 때 상태가 올바르게
  닫히는지 확인된 적이 없다.**
- `prepareRegister` B3·B4 — 이미 pending / 이미 active. 같은 포지션에 보호주문이 두 번 나가는 것을
  막는 방어선이며, 재시도·복구 경로에서 가장 먼저 부딪힐 분기다.

**따라서 tasks의 순서는 "거부 경로 RED 테스트 9건 → 배선"이다.** 배선이 먼저가 아니다.

## Function Logic Map 면제 기록

`internal/app/engine` — `strategyDispatchCycle.dispatch`: **not-applicable (조건부)**

design 단계에서 수렴 지점을 `dispatch` 바깥으로 확정했으므로(design D7) 이 함수를 편집하지
않고 그 내부 분기를 근거로 인용하지도 않는다.

> 리뷰 이전 판은 여기서 "분기 14 / return 16"이라고 **수치를 적었다.** 그 수치의 산출물이
> 이 change의 `analysis/function-logic/`에 없다. 면제 사유에 분기 수를 쓰는 것 자체가 규칙이
> 금지하는 「산출물 없는 분기 주장」이므로 **수치를 지운다.** 면제에 필요한 것은 "편집하지
> 않고 인용하지 않는다"뿐이고, 그것은 분기 수를 몰라도 성립한다.

`internal/filldetect` 전체: **편집 not-applicable.** a100은 이 패키지를 편집하지 않고 조립
지점만 바꾼다. 다만 **분기를 근거로 인용하므로** `pollLocked`의 산출물은 만들었다.
"편집하지 않음"은 인용의 면제 사유가 아니다.

**조건**: 구현 중 위 둘 중 하나를 편집하거나 그 내부 분기를 새로 근거로 삼게 되면 **그 시점에
Function Logic Map을 만들고 해당 task 착수 전에 완성한다.**
