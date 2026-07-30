# Review: verify-observes-the-trigger

날짜: 2026-07-30 · 위험 등급: **High-risk** (이 도구가 처음으로 체결을 의도한 라이브 주문을 만든다)

## Pre-Edit Gate

```text
change id / task id:  verify-observes-the-trigger / 1.1-1.23
대상 심볼 (기존 함수 내부 수정):
  verifylive.outstandingLines             (record.go:494)  — 종결 판정 + 역행 가드 확장
  verifylive.(*Runner).sweepStep          (cleanup.go:277) — 의도적 보존 스킵 가산
  verifylive.(*Runner).checkConditionalCap(mutate.go:608)  — 단계 전용 상한
  verifylive.(*Runner).preflightStatic    (runner.go:530)  — 유예형/옵트인 상호작용
  verifylive.(*Runner).optedIn            (runner.go:584)  — 두 번째 플래그
  verifylive.(*Runner).preflight          (runner.go:495)  — mutatesNow
  verifylive.(*Runner).entryFor           (runner.go:606)  — Mutating을 실행 형태로
  verifylive.(*Runner).Plan               (plan.go:532)    — mutatesNow
  verifylive.(*Runner).stepConditionalTrigger (steps.go:718) — 전면 재작성
  verifylive.Steps() StepConditionalTrigger 항목 (verifylive.go:443)
  verifylive.WriteSteps                   (report.go:368)  — 시그니처 +1
가산 (신규):
  verifylive.NearStopTrigger, Artifact.Filled/FilledAt, MaxLiveConditionalsTrigger,
  FlagIncludeTrigger, Options.IncludeTrigger, Runner.deferredForm/mutatesNow/Abort,
  StepAbort, tossctl verify abort
기존 동작 파악 근거:
  analysis/function-logic-map.md (전 분기 표 + Branch Test Map)
  읽은 파일: record/cleanup/mutate/runner/plan/steps/retry/report/verifylive.go 전문,
             static_test.go, cmd/tossctl/verify.go, console/{pages,data}.go
  소비자 전수: grep Outstanding|PendingCleanup|Cancelled (9개 파일 32곳)
upstream 상속 테스트 영향: no — internal/verifylive는 TossOS 전용 패키지다
실패 테스트 선행 작성: yes (1.x 전부 RED 선행)
안전 불변식 §0 위반 여부 검토: 통과하되 조항 1·5·6은 아래 표에서 명시적으로 다룬다
```

## §0 대조

| 조항 | 이 change |
|---|---|
| 1 승인 없는 LIVE side effect 금지 | **가장 무거운 조항.** 이 change는 체결을 의도한 주문을 만든다. 세 겹으로 막는다 — ① 옵트인 플래그 없이는 경로 자체가 실행되지 않고 ② 실행해도 기존 배치 승인(만료되는 타이핑 문자열)을 거치며 ③ 실측(3.x)은 사람 입회를 tasks에 조건으로 박았다. 자동 테스트는 fake broker만 쓴다 |
| 2 mutating 자동 실행 금지 | `verify run`·`verify abort` 둘 다 `mutating: true` — 에이전트가 실행하지 않는다 |
| 3 토글 OFF = upstream 동작 | `--include-trigger` 없으면 발동 단계는 **오늘의 세 관측 + deferred 판정**을 그대로 남기고 `NearStopTrigger`는 호출되지 않는다. `Filled` 필드가 없는 기존 기록의 판정도 불변 (1.6) |
| 4 손절·비상 청산 즉시성 | 무관 — 이 패키지는 엔진 경로가 아니다. `internal/flatten` import는 static_test가 금지한다 |
| 5 High-risk 경로 | 해당. 주문·조건주문·원장 전부 건드린다 → full TDD + FLM + 적대적 리뷰 + gate |
| 6 손절·사이징은 보수 방향만 | **의도적 예외이고 그것이 측정 내용이다.** 이 도구의 다른 12개 단계가 지키는 "체결되지 않을 가격"을 발동 단계만 반전한다. 반전의 범위는 상수로 못박는다 — SINGLE + MARKET + SELL + 1주, 별도 함수, 옵트인 |
| 7 운영 토글 flip·live 검증은 사람 | 3.1(`exclude_symbols` 설정)과 3.3(실행) 모두 사람 몫으로 tasks에 명시 |
| 8 시크릿·개인정보 미저장 | 기록은 digest·상태 문자열·식별자만. 새 필드 `FilledAt`은 시각이다 |
| 9 주문은 공식 Open API만 | 변경 없음 |
| 10 실계좌 자동 테스트 금지 | `testenv.Guard` + fake broker 유지, static_test가 강제 |

## 계약과 어긋나 고친 것

구현 전 코드 확인에서 계약(design.md/tasks.md/spec)이 코드와 맞지 않는 것 둘을 찾았다.
FLM §0에 근거를 적었고 계약 문서를 코드에 맞춰 고쳤다.

1. **상한의 종류.** D6은 `MaxLiveOrdersTrigger`를 적었으나 주문 상한은 이 단계에서
   발동하지 않는다 — child 주문은 이 도구가 접수하는 것이 아니라 브로커가 만들고 우리가
   발견한다. 실제로 막는 것은 `MaxLiveConditionals`다. → `MaxLiveConditionalsTrigger`.
2. **`sweepStep`이 붙잡힌 child를 취소한다.** D4의 결말 ②는 `fail`이지 `abort`가 아니라
   sweep이 돌고, 오늘의 sweep은 이 단계가 남긴 모든 order artifact를 취소한다. 계약이
   "붙잡힌 채 보고된다"고 적은 것이 코드에서 성립하지 않았다. → `Deliberate` 스킵 가산.

## 적대적 Eng 리뷰

날짜 2026-07-30 · 시점 구현 후 · 방식: "이 단계가 실계좌에 무엇을 남길 수 있나"를 경로별로 추적.
아래 다섯은 전부 **구현을 고쳐서** 닫았고, 각각 테스트가 붙어 있다.

### A1. 중단된 실행이 발동 가능한 손절을 계좌에 남기고 도구는 깨끗하다고 보고한다 — **P0**

초안은 발동 단계의 조건주문에 `markHeld`를 걸었다. `markHeld`는 세 가지를 한꺼번에 말한다 —
의도적이다, 이 단계의 판정이 놓아준다, 이 사슬 소속이다 — 그리고 앞의 둘이 이 객체에는 거짓이다.

- `Deliberate`는 실행 끝 잔여물 검사(`undeliberate`)에서 **면제**시킨다. Ctrl-C로 창 한복판에서
  끊기면 시장 바로 위에 놓인 손절이 살아 있는데 실행은 깨끗하게 끝났다고 보고한다.
- `HeldUntil`은 다음 실행의 정리 prologue도 막는다. artifact 줄과 판정 줄이 **같은 entry**라
  `heldAfter`가 영영 거짓이 되어(`decided > at` 불성립) `verify abort` 외에는 아무도 못 건든다.

→ `sr.joinChain`을 새로 만들어 **사슬만** 기록한다. 결과: 중단되면 실행이 시끄럽게 실패하고,
다음 실행의 prologue가 취소를 승인 목록에 올린다. child 주문은 그대로 `markHeld` — 체결되게
두는 것이 측정이므로 면제가 맞다.
테스트 `TestATriggerWithNoObservableFillLeavesTheChildHeld`가 "깨끗하게 끝났다고 보고하지
않는다"를 단언한다.

### A2. 조건주문이 둘일 때 `conditional-cancel`이 엉뚱한 것을 취소한다 — **P0**

`liveConditional()`이 "outstanding 중 첫 조건주문"을 돌려준다. 발동 단계가 자기 것을 등록하면
전체 실행에 조건주문이 둘이 되고, 기록 순서상 **발동 단계의 것이 먼저**다. 결과:
`conditional-cancel`이 발동 측정의 대상을 취소하고 원래 지워야 할 것을 남긴다 — 한 번의 호출로
측정 두 개가 동시에 깨진다. 테스트가 실제로 이 상태를 재현했다.

→ 출처로 가른다. 생성 단계가 `conditional-trigger`인 조건주문은 제외한다. 생성을 볼 수 없는
조건주문(옛 기록, 이어하기)은 그대로 대상이라 기존 판정은 불변이다.

### A3. 임계에 도달했는데 발동하지 않은 경우가 `skipped`로 묻힌다 — **P1**

초안은 `triggeredAt`이 비었으면 무조건 "시장이 오지 않았다 → INCONCLUSIVE"였다. 그런데
**최종체결가가 발동가에 닿았는데도 발동하지 않았다**면 그것은 시장 조건이 아니라 브로커가
자기 보호 주문의 조건을 무시했다는 뜻이고, **2c가 조건주문에 기댈 수 없다는 근거**다. 이 change가
만들 수 있는 가장 중요한 부정 결과가 조용히 skip으로 기록되고 있었다.

→ 별도 결말. 임계 도달 후 `TriggerLinkWindow` 안에 발동이 없으면 `fail` +
`conditional.fires_when_its_condition_is_met=false`. 등록한 조건주문은 그대로 취소한다.
테스트 `TestAStopWhoseConditionWasMetAndDidNotFireIsAFailure`.

### A4. 발동은 봤지만 체결을 못 본 경로에서 조건주문을 "사라졌다"고 단정했다 — **P1**

`fail` 경로에서 조건주문을 `Filled`로 종결시키고 있었다. 두 방향의 오류가 비대칭이다 —
살아 있는 것을 사라졌다고 적으면 발동 가능한 주문을 아무도 추적하지 않게 되고, 사라진 것을
살아 있다고 적으면 필요 없던 취소가 404 한 번 나올 뿐이다.

→ 체결을 **실제로 확인한 경로에서만** 종결로 적는다. `fail` 경로는
`conditional_presumed_fired=true` 관측만 남기고 살아 있는 것으로 둔다(A1과 맞물려 시끄럽게 끝난다).

### A5. 관측 창의 요청량 — **P2**

발동을 본 뒤에도 매 tick마다 `/prices`와 `/orderbook`을 계속 읽고 있었다. 발동 후에는 임계 도달
여부가 이미 결정됐고 발동 시점 호가도 이미 기록됐으므로 순수 낭비다. 이 창은 429 하나가 측정
전체를 날리는 구간이다(J5).

→ `triggeredAt`이 잡히면 시세·호가 폴링을 멈춘다. 활성 구간 요청이 tick당 4 → 1~2로 준다.

### 닫지 않고 남긴 것

- **J1 `Order.SubmittedAt`** — 이 change는 그 필드를 시각 근거로 쓰지 않을 뿐 고치지 않는다.
- **취소 후 경합 판정의 잔여 불확실성** — 취소된 조건주문은 다시 읽히지 않으므로 최종 근거는
  보유 수량이다. 보유 수량마저 못 읽으면 `skip`이 아니라 `fail`로 끝내고 사람에게 직접 확인을
  요구한다(`raceUnknown`). 코드로 더 줄일 수 없다.
- **`TriggerPollIdle`·`TriggerLinkWindow`의 값** — 실측 전에는 근거가 추정이다. 상수 주석에
  근거를 적었고, 첫 실측 뒤 조정한다.

## 검증 결과

| 항목 | 결과 |
|---|---|
| `go test ./...` | **3806 passed**, 0 failed (57 packages) |
| `make vet` | 통과 |
| `make validate` | 30 passed, 0 failed |
| 실기록 재생 (2.3) | `capability-verify.jsonl` 61 entry / artifact 줄 16개, `capability-verify-us.jsonl` 34 entry / 18개 — **전 줄이 변경 전 규칙과 동일하게 분류**, outstanding 0, 정리 대상 0 |
| 상속 테스트 회귀 | 0 |
