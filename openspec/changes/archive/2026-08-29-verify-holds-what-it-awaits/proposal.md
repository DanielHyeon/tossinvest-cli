# Proposal: verify-holds-what-it-awaits

## Why

발동 측정(`conditional-trigger`, task 2.5의 마지막 구멍)은 **정리 prologue가 현재
구조로는 반드시 파괴하는** 객체를 남긴다. 그 단계를 쓰기 전에 정리 규칙을 고쳐야 한다.

### 정리 규칙의 현재 모양

`cleanupFrom`([cleanup.go:121](../../../internal/verifylive/cleanup.go))이 대상을 정한다.

| artifact | 대상 조건 |
|---|---|
| `KindOrder` | **무조건 대상** |
| `KindConditional` | `Settled(conditional-cancel)` **그리고** `decidedAfter(conditional-cancel, a)` |

조건주문 쪽 두 번째 항은 `verify-reopens-conditional-chain`이 M37을 고치며 넣은 것이다.
그 change가 세운 원칙은 옳고 이 change는 그것을 **넓힌다**:

> 객체가 존재하기 전에 기록된 판정은 그 객체에 대한 판정이 아니다.

### 발동 측정이 깨는 두 전제

**P1 — 일반 주문에는 보호가 전혀 없다.** [cleanup.go:76-79](../../../internal/verifylive/cleanup.go)가
전제를 명시적으로 적어 뒀다.

> *"No step in the catalogue ever cancels an order from an earlier run — each step
> cancels what it placed itself — so if the prologue does not take it, nothing will."*

발동 측정은 이 전제를 정면으로 깬다. 조건주문이 발동하면 **child 시장가 매도**가 생기고,
그 주문은 **체결되어야 한다** — "취소하지 않고 남긴다"가 측정의 내용 자체다. 그런데 그 child는
이 도구가 만든 `KindOrder` artifact이므로 다음 실행의 prologue가 승인 목록에 올려 취소한다.
M37과 같은 형태다. 다른 점은 대상이 조건주문이 아니라 주문이라는 것뿐이고, 조건주문 쪽에만
가드가 있다.

**P2 — 기다린다는 사실을 말할 방법이 없다.** 현재 가드는 `StepConditionalCancel` 하나를
이름으로 박아 두고 묻는다. 발동을 기다리는 단계는 자기가 무엇을 기다리는지 기록에 남길 수
없고, 자기 측정 대상의 생존이 **다른 단계의 마지막 판정이 무엇이냐**에 달린다.

**그리고 기록은 이미 의도를 담고 있는데 정리가 읽지 않는다.** `Artifact.Deliberate`는
[record.go:190-192](../../../internal/verifylive/record.go)에 "살아 있는 것이 실수가 아니다"라고
적혀 있고 `markDeliberate`([runner.go:730](../../../internal/verifylive/runner.go))가 조건주문 세 곳에
찍는다. 화면·리포트·redo는 그것을 읽는다. **`cleanupFrom`만 안 읽는다.**

### 왜 지금인가

2.5의 남은 항목 중 2c-A 선행 필수는 발동 하나다(measurements.md "2.5에 남은 미측정"). 발동
측정 단계를 이 수정 없이 쓰면, 그 단계가 만든 child 주문을 다음 실행이 지우고 **체결 관측이
불가능해진다**. 순서를 바꿀 수 없다.

## What Changes

- **`Artifact.HeldUntil StepID`**(가산·`omitempty`) — 이 객체는 지정한 단계가 그 선언 **뒤에**
  기록된 terminal 판정을 받기 전에는 정리 대상이 아니다. 현재 조건주문 규칙을 필드 부재 시
  기본값으로 표현하므로 **기존 기록의 판정이 바이트 단위로 동일하다**.
- **`Artifact.ChainID`**(가산·`omitempty`) — 한 측정이 만든 객체들이 같은 식별자를 단다.
  정정이 새 id를 발급해 옛 id가 404가 되어도(M40) 둘이 한 사슬임을 기록이 말한다. 지금은
  산문 `note`에만 있다.
- **정리 대상 규칙을 한 문장으로 통합** — kind별 분기 대신 "gate 단계가 이 선언 뒤에 판정을
  내렸는가". 조건주문의 기존 gate(`conditional-cancel`)는 필드 없는 artifact의 기본값으로
  보존하고, 주문의 기존 gate(없음 = 무조건 대상)도 그대로 보존한다.

## Non-Goals

- **`conditional-trigger` 단계 구현** — 별도 change. 이 change는 그 단계가 존재할 수 있는
  조건만 만든다.
- **시각 기반 lease 만료** — 채택하지 않는다. 근거는 design.md D3.
- **운영자가 측정 사슬을 끝내는 콘솔 조작** — 별도 change(design.md D5의 잔여 위험).
- 노출 상한·승인 창·계획 인가·`ErrOutsidePlan` 레일 — 전부 무변경.

## Impact

- 위험도 **High-risk**: 라이브 취소 요청의 대상 목록을 정한다.
- 영향 파일: `internal/verifylive/{record.go,cleanup.go,runner.go,steps.go}`.
- 기록 스키마: 가산 nullable 2필드, `FormatVersion` 무변경(§0.6).
- **방향 — 이 change 단독으로는 취소 대상 목록이 바뀌지 않는다.** 기존 기록은 필드가 없어
  기본값으로 오늘과 같은 판정을 받고, 이 change가 새로 붙잡는 객체는 없다(주문에 `HeldUntil`을
  찍는 코드는 발동 change가 넣는다). 순수하게 **표현할 수단**을 만드는 change이고, 그것이
  라이브 취소 목록을 건드리는 change로서 가장 안전한 모양이다. 관측 가능한 차이가 없다는
  주장 자체가 task 1.8의 실기록 회귀 테스트가 증명할 대상이다.
