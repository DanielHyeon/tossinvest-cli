# Review: verify-plans-the-object-it-mutates

## Pre-Edit Gate (High-risk)

```text
Pre-Edit Gate:
- change id / task id: verify-plans-the-object-it-mutates / 1.1~1.8
- 대상 심볼(패키지.함수):
    internal/verifylive.Step                    (가산 필드 ActsOnConditional)
    internal/verifylive.Steps                   (기존 함수 — 카탈로그 2항목 선언)
    internal/verifylive.Runner.mutationSymbol   (기존 함수 내부 편집 — 이 change의 중심)
    internal/verifylive.Runner.Plan             (기존 함수 내부 편집 — 제외 분기 1개 추가)
    internal/verifylive.Plan.Authorises         (기존 함수 내부 편집 — 종목 비교 축소)
    internal/verifylive.Runner.preflightStatic  (무변경 — diff 문맥만, base revision 고정)
- 기존 동작 파악 근거:
    · 실측: ~/.local/share/tossos/capability-verify.jsonl 13:28:16Z·13:28:54Z 두 run.
      approval pass(requests_listed=2, steps_listed="conditional-modify, conditional-cancel",
      plan_digest sha256:c2a06fd2…) 직후 conditional-modify fail(ErrOutsidePlan).
      두 run의 호출은 GET /api/v1/conditional-orders/{id} 각 1건 — 전송 0건.
    · mutationSymbol 호출부 2곳(CodeGraph + grep): plan.go:564 Plan, runner.go:550 preflightStatic
    · Authorises 호출부: mutate.go:176 authorise + 테스트 2건(transient_test.go:104,
      cleanup_test.go:57). mutate.go가 인가를 거치는 유일한 경로임은
      static_test.go TestTheApprovedPlanIsTheOnlyThingMutateGoActsOn이 고정
    · liveConditional 호출부(CodeGraph): stepSellableReserved·stepConditionalPersist(비-mutating),
      stepConditionalModify·stepConditionalCancel(mutating), 그리고 이제 mutationSymbol
    · NeedsHolding 선언 위치: verifylive.go:323 sell-boundary, verifylive.go:369
      conditional-register — 두 곳뿐
    · 콘솔이 넘기는 값: cmd/tossctl/console.go:745 Symbol=consoleProbeSymbol(KR이면 상수 005930),
      :746 HoldingSymbol=firstUsableHoldingIn
    · 기존 테스트: plan_test.go 24건(특히 TestThePlanDigestIsPinnedAcrossBuilds,
      TestASymbolSubstitutionIsNotAuthorised, TestEveryPlannedLineSaysWhatItIsAndHowItEnds),
      cleanup_test.go 7건, us_market_test.go
- upstream 상속 테스트 영향: no — internal/verifylive는 TossOS 신규 패키지다.
  전체 go test 3742 → 3749 (신규 7건), 회귀 0.
- 실패 테스트 선행 작성: yes — RED 3회 관측·기록(컴파일 실패 1, 해석 미수정 1, 종목 없는 줄 1).
  원문은 analysis/function-logic/internal-verifylive--runnermutationsymbol/branch-test-map.md.
- 안전 불변식 §0 위반 여부 검토: 통과 (아래 §0 대조표)
```

### §0 대조

| 조항 | 판단 |
|---|---|
| §0.1 승인 없는 LIVE side effect 금지 | 통과 — 구현·테스트 중 브로커 호출 0건(fake broker만). 실계좌 확인(task 3.1)은 사람이 콘솔에서 승인해 실행한다. |
| §0.2 토글 OFF = upstream | 해당 없음 |
| §0.3 손절·비상 청산 즉시성 | 해당 없음 — 엔진 런타임 무변경. 대상 조건주문은 검증이 만든 것이다. |
| §0.4 rate limit 계상 | 통과 — 새 호출 종류 0. `mutationSymbol`이 부르는 `liveConditional`은 기록만 읽는다(네트워크 0). |
| §0.5 운영 설정 audit | 통과 — 승인 기록(`approval.*`) 항목 무변경. |
| §0.6 스키마 변경 | 통과 — `Step.ActsOnConditional`은 가산·`omitempty`. 기록 스키마(`Entry`)와 `RecordFormatVersion` 무변경. |
| §0.7 운영 토글 flip은 사람이 | 통과 — flip 없음. 자동 승인 경로 신설 0. |
| §0.8 scope 밖 위험 코드 변경 금지 | 통과 — 엔진·인터록·Guardian·원장 무변경. diff는 `internal/verifylive` 3파일. |
| §0.9 보수 방향만 | 아래 A1에서 정면으로 다룬다. |

## 리뷰 (2026-07-29, 적대적 Eng 관점 포함)

### A1. "라이브 요청을 하나 더 가능하게 만드는 change다" — **맞다. 그것을 감추지 않는다**

정직한 서술은 이렇다: 이 change 이후 `conditional-modify`·`conditional-cancel`은
**보낼 수 있게 된다**. 지금은 인가에서 막혀 아무것도 못 보낸다. 그러니 "순수한 축소"라고
말하면 거짓이다.

무엇이 가능해지는지가 논점이다.

- 두 요청은 **이미 승인 화면에 표시되고 있었다**. 실기록의 `approval.requests_listed = 2`,
  `approval.steps_listed = "conditional-modify, conditional-cancel"`. 사람은 이 두 줄을 읽고
  승인했고, 도구가 그 승인을 실행하지 못했을 뿐이다. 이 change는 승인 범위를 넓히는 것이
  아니라 **이미 받은 승인과 실제 행동을 일치시킨다**.
- 새 주문은 생기지 않는다. modify는 발동가를 시장에서 **한 호가 더 멀리** 옮기고
  (`PriceOneTickFurther`), cancel은 이 도구가 등록한 조건주문을 **제거**한다. 노출이
  늘어나는 방향의 요청은 이 change로 가능해지는 집합에 없다.
- 사람의 승인 행위는 여전히 모든 요청 앞에 하나씩 있다. 승인 창(5분)·nonce·노출 상한·
  1주 규칙 무변경.

그리고 반대 방향의 비용이 실재한다. 지금 상태로 두면 이 도구가 등록한 조건주문
`p7hQz7HAXc…`를 **이 도구가 제거할 수 없다**. 정리 prologue도 건드리지 않는다
(`decidedAfter`가 07-26 판정을 올바르게 거절한다). 도구가 계좌에 남긴 것을 도구가 치우지
못하는 상태를 유지하는 것이 더 보수적이라는 주장은 성립하지 않는다.

§0.9 판정: **통과**. 가능해지는 요청 집합은 정정 1건·취소 1건이고 둘 다 노출을 줄이는
방향이며, 승인은 이미 받은 것이다.

### A2. "`NeedsHolding: true` 두 글자면 끝나는 것 아닌가" — **아니다, 두 가지 이유로**

① `NeedsHolding`은 `preflightStatic`(runner.go:545)에서 "계좌에 쓸 수 있는 보유가 없으면
건너뛴다"까지 뜻한다. 보유가 사라지고 조건주문만 남은 계좌 — 잔여물을 반드시 치워야 하는
바로 그 상태 — 에서 `conditional-cancel`이 skip된다.

② 잔여 조건주문이 현재 `holdingSymbol`과 다른 종목일 수 있다. `firstUsableHoldingIn`은
"이 시장의 첫 쓸 만한 보유"를 고를 뿐이고, 조건주문이 걸린 종목과 같다는 보장이 없다.
그 경우 `NeedsHolding`은 여전히 틀린 종목을 싣는다. 대상을 이름하려면 **대상을 봐야 한다**.

### A3. "계획 digest가 움직인다 — 진행 중인 검증이 깨지지 않나" — **깨지지 않는다, 확인함**

테스트 fixture는 전부 `Symbol == HoldingSymbol`이라 `TestThePlanDigestIsPinnedAcrossBuilds`가
그대로 통과한다(실행 확인). 그러나 실계좌에서는 probe 005930·조건주문 333430이므로 그 두
줄의 digest는 **바뀐다**. 그것이 문제인지 확인했다:

`plan.Digest()`의 유일한 비-테스트 사용처는 `runner.go:466`이고, 기록에 **쓰기만** 한다.
읽어서 비교하는 코드는 없다(`grep -rn "plan_digest|Digest()" --include=*.go`, 비-테스트
결과 3건 중 verifylive는 이 한 곳). `--resume`은 새 계획을 만들어 **다시 승인을 묻는다**
(`approveBatch`). 즉 digest는 run별 증거이지 run 간 대조 키가 아니다.

남는 사실 하나는 정직하게 적어 둔다: 이 change 이후 KR 조건주문 줄의 digest는 이전
기록들과 다르다. 그것은 **다른 목록이기 때문이고**, 다른 목록이면 digest가 달라야 맞다.

### A4. "종목 없는 줄 제외(D3)와 와일드카드 제거(D4)는 scope 밖 아닌가" — **경계 논점, 수용하되 포함**

이 change가 세우는 요구는 "계획의 각 줄은 그 요청이 무엇에 대한 것인지 말한다"이다.
"아무것도 말하지 않는 줄"은 그 요구의 부정이지 별개 주제가 아니다. 그리고 이것은 가정이
아니라 관측이다 — RED 3에서 US·보유 0 계좌의 계획에 **종목을 말하지 않는 라이브 주문 줄이
4개** 실렸다. `Authorises`의 옛 분기는 그런 줄에 어떤 종목의 요청이든 통과시킨다.

두 수정은 각각 독립적으로 되돌릴 수 있게 task를 분리했다(1.6, 1.7). 리뷰가 거절하면
1.2~1.5만으로도 오늘의 결함은 고쳐진다.

### A5. "정적 AST 테스트가 취약하지 않나" — fail-closed로 썼다

`dispatch`의 switch를 못 읽으면 `t.Fatal("dispatch bound no step to a body")`,
`StepID` 상수를 못 읽으면 `t.Fatal`, 알려진 두 본문에서 `liveConditional`을 못 찾으면
`t.Fatal`. 조용히 0건을 검사하고 통과하는 경로가 없다. 카탈로그 식별자를 손으로 베껴
두지 않고 `verifylive.go`의 const 블록에서 읽는다 — 베낀 목록이 낡는 것이 이 change가
고치는 결함의 형태이기 때문이다.

### A6. "남는 좁은 구간이 있다" — 있다, 그리고 거절 방향으로 틀린다

종목 A의 잔여 조건주문이 정리 대상이면서 같은 실행이 보유 B로 `conditional-register`를
다시 도는 경우, 계획은 A를 싣고 실행 시점에는 B가 살아 있다. 결과는 `ErrOutsidePlan`:
아무것도 전송되지 않고 실행이 멈춘다. 다음 실행에서는 A가 outstanding에서 빠져 정상
진행한다. **스스로 복구되고, 틀린 것을 보내는 방향으로 틀리지 않는다.** 계획 시점의
`willRun`을 `mutationSymbol`까지 끌고 들어가면 preflight의 run-time 호출부와 의미가 갈라지고,
그 갈라짐이 바로 이 change가 고치는 결함의 형태다. 만들지 않는다(design.md D1).

### A7. "실계좌에서 되는가" — **아직 모른다. 그렇게 보고한다**

이 change의 GREEN은 fake broker 위에서다. 실계좌 확인은 task 3.1이고 사람이 콘솔에서
승인해 실행한다. 게다가 지금은 KR 장 마감이라 정정·취소가 시간외에 접수되는지는
**측정된 적이 없다**. 22:43 KST에 `conditional-register`가 통과한 것은 관측된 사실이지만
정정·취소가 같다는 근거가 아니다. 실패하면 `order-hours-closed` 계열 사유가 기록되고
그것도 실측이다 — 그때 실패하는 자리는 **인가가 아니라 브로커**여야 하며, 그 구별이
이 change가 만드는 차이다.

### A8. "정리 경로가 깨지지 않나" — 테스트로 고정

`Authorises`의 빈 종목 동작은 "빈 계획 줄 ↔ 빈 요청"만 남는다.
`TestAPlanLineWithoutASymbolAuthorisesNothingWithOne`의 두 번째 단언이 그것이고,
`cleanup_test.go` 7건과 `transient_test.go`가 그대로 통과한다.

### 결정

수용하고 진행한다. 미해결 코드 이슈 없음. A7(실계좌 확인)은 task 3.1로 남으며,
그 전에는 이 change를 "실측으로 검증됨"이라고 보고하지 않는다.

## Function Logic Map

적용함. 5개 target — 수정한 기존 함수 4건(`Runner.mutationSymbol`, `Runner.Plan`,
`Plan.Authorises`, `Steps`)과 diff 문맥에만 걸린 무변경 함수 1건
(`Runner.preflightStatic`, base revision으로 고정).

`python3 tools/logic-map/check_analysis.py --change verify-plans-the-object-it-mutates`
→ `evidence complete or diff-proven exempt`

## 실행한 명령과 결과

```
go test ./internal/verifylive/ -count=1     210 passed
go test ./... -count=1                      3749 passed in 57 packages  (07d4ba0 대비 +7, 회귀 0)
go vet ./...                                No issues found
openspec validate … --strict                valid
```
