# Design: verify-holds-what-it-awaits

## D1. 규칙은 하나다 — "gate 단계가 이 선언 뒤에 판정을 내렸는가"

현재 `cleanupFrom`은 kind로 갈라 두 규칙을 쓴다. 이 change는 규칙을 하나로 만들고 kind별
동작을 **기본값**으로 표현한다.

```text
gate(a) = a.HeldUntil                    필드가 있으면 그것
        = StepConditionalCancel          없고 kind == conditional-order  (기존 규칙)
        = ""                             없고 kind == order              (기존 규칙)

대상이다 ⟺ gate == ""  또는  (Settled(gate) 그리고 heldAfter(entries, gate, l.at))
             l.at = Outstanding이 이 객체로 고른 줄의 index (D2)
```

`heldUntil`이 비어 있으면 판정이 지금과 **완전히 같다**. 이것이 §0.2("토글 OFF = upstream
동작")를 기록 스키마에서 지키는 방식이다: 기존 KR·US 기록의 모든 artifact는 필드가 없으므로
오늘과 동일하게 판정된다.

**기본값은 `Kind`로만 갈린다 — `Deliberate`는 읽지 않는다.** 옛 코드가 정확히 그랬기
때문이다(`cleanupFrom`의 `switch a.Kind`). `Deliberate`가 아닌 조건주문도 같은 gate를 받고,
그래서 보존이 "대부분의 경우"가 아니라 **전부**다. 이 사실은
`TestALegacyRecordIsJudgedExactlyAsBefore`의 마지막 항목이 고정한다.

## D2. 비교 기준은 "gate를 지목한 그 줄"이다

규칙을 한 문장으로 적으면 이렇다.

> **어떤 줄이 지목한 gate는 그 줄보다 뒤에 판정해야 그 객체를 놓아준다.**

`Outstanding`은 각 식별자의 **마지막 비취소 언급**을 고른다. 우리가 `HeldUntil`을 읽는 줄이
바로 그 줄이므로, 비교 기준도 그 줄이어야 한다. gate를 지목한 줄과 판정을 재는 기준선이
같은 줄이라는 것이 이 정의의 전부다.

현재 `decidedAfter`는 기준선이 **최초** 언급이다. 지목이 나중에 바뀔 수 있는 순간(같은
객체를 다른 gate로 다시 붙잡는 경우) 두 정의가 갈린다.

| 기록 | `decidedAfter`(현재) | `heldAfter`(이 change) |
|---|---|---|
| register(CO) → cancel fail → persist가 다시 붙잡음 | 기준=0, 판정=1 → **놓아준다** | 기준=2, 판정=1 → **붙잡는다** |
| register(CO) → persist가 다시 붙잡음 → cancel fail | 기준=0, 판정=2 → 놓아준다 | 기준=1, 판정=2 → 놓아준다 |

**첫 줄이 이 change가 고치는 것이고, 둘째 줄이 M22를 되살리지 않는다는 증거다.** 실패한
취소가 지목보다 **뒤**에 있으면 여전히 놓아주므로 잔여물 교착은 자라지 않는다. 앞에 있을
때만 붙잡고, 그때는 redo가 새 판정을 찍어 풀 수 있다.

**시계가 아니라 index다.** `cleanup.go:136-149`가 이미 그 이유를 적어 뒀고 여기서도 같다 —
기록은 append-only JSONL이라 index만 단조이고, 취소 줄은 zero time을 싣는다.

**설계 중 폐기한 대안**: 처음에는 기준선을 "가장 최근의 붙잡음 선언"으로 잡으려 했다.
`Outstanding`이 고른 줄과 다른 줄을 기준선으로 삼는 정의여서, 취소 실패 뒤에 다른 단계가
객체를 다시 언급하기만 해도 붙잡히는 경우가 생긴다 — M22를 조건주문 쪽에 되살리는 형태다.
읽는 줄과 재는 줄을 같게 두는 것이 그 위험을 정의 수준에서 없앤다.

## D3. 시각 기반 lease 만료를 채택하지 않는 이유

lease에 만료 시각을 주면 **시간이 지났다는 이유로 살아 있는 브로커 객체를 취소 대상으로
되돌리는 규칙**이 생긴다. 그것이 정확히 M37의 형태다 — 운영자가 볼 수 없는 파생이 측정
대상을 지운다. 게다가 만료는 **발동 대기 창 한복판에서 발화한다**: 대기가 길어지는 것은
측정이 실패했다는 뜻이 아니라 아직 가격이 임계를 안 넘었다는 뜻이다.

그래서 해제는 **위치 기반**으로만 한다: gate 단계가 판정을 내리면 풀린다. 프로세스가 죽어
영영 판정이 없으면 객체는 계속 붙잡힌 채 `Outstanding`에 보이고 화면·요약이 매번 보고한다 —
**보이는 상태로 멈추는 것**이 조용히 취소되는 것보다 안전하다(§0.3의 방향).

대가는 D5에 잔여 위험으로 적는다.

## D4. `Deliberate`는 유지한다

`Deliberate`는 화면 문구·`undeliberate`(runner.go:756)·`redo.go:114`가 읽는다. `HeldUntil`이
그것을 대체하지 않고 **덧붙는다**: `Deliberate`는 "이것이 실수가 아니다"(사람에게 하는 말),
`HeldUntil`은 "누가 판정할 때까지 손대지 마라"(정리 규칙에게 하는 말)다. 같은 순간에 함께
찍히지만 읽는 쪽이 다르므로 하나로 합치면 한쪽 의미가 다른 쪽을 따라 흔들린다.

`markDeliberate`가 두 필드를 함께 쓰도록 확장하고, gate 단계를 인자로 받는다.

**리뷰 A2의 우려는 구현에서 해소됐다.** 처음에는 붙잡음 탐색이 `Deliberate`도 읽어야 할 것
같았고, 그러면 두 필드가 결합되어 D4의 분리 주장이 무너진다. 실제로는 그럴 필요가 없었다 —
정리 판정 경로(`cleanupFrom`·`holdGate`·`heldAfter`) **어디도 `Deliberate`를 읽지 않는다.**
기본값이 `Kind`로 갈리기 때문이고, 그것이 옛 코드가 하던 일과 정확히 같다. 두 필드는
구현에서도 독립이다.

## D5. 이 change가 닫지 않는 것 — 잔여 위험

**붙잡힌 사슬을 끝내는 운영자 조작이 콘솔에 없다.** gate 단계가 영영 판정을 못 받으면
`MaxLiveConditionals`(1)이 차서 다음 조건주문 측정이 막힌다. CLI `--redo <gate-step>`은
탈출구지만 사용자는 콘솔로 운전하고, 콘솔이 제시하는 `RedoSet`은 마지막 판정이 fail·skipped인
단계만 담는다 — gate가 `pass`면 목록에 없다. **이것은 M37을 못 벗어나게 만들었던 바로 그
구조**이며(`verify-reopens-conditional-chain` review.md), 이 change는 그것을 재현하지 않지만
해소하지도 않는다.

발동 change가 이 조작을 함께 만들어야 한다. issues.md에 이연으로 기록한다.

## D6. 범위 밖

- `conditional-trigger` 단계 자체, child 주문 관측, 발동 지연 시각 4종
- `ProtectiveCapability` 산출
- 노출 상한 값·승인 창·계획 인가 — 무변경
- 브로커 truth 재조회로 붙잡음을 검증하는 것 — 기록이 권위인 현재 설계를 바꾸는 일이라 별건
