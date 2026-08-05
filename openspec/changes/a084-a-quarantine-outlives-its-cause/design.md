# a084 · Design

## D1 — 왜 "선택기가 바뀌었을 때"인가

격리는 판정이다. 판정의 출력을 바꿀 수 있는 것은 둘뿐이다.

```
입력이 바뀐다     saved snapshot, recomputed snapshot, policy identity
판정자가 바뀐다   SelectRecoverySnapshot의 의미
```

격리된 세대는 **새 snapshot을 쓰지 않는다**. `RecordExitJudgement`가
`ErrExitSnapshotQuarantined`로 끊고, `workingSet`은 아예 판정에 보내지 않는다. saved는
격리 직전의 값에 고정되고, recomputed는 매 관측마다 새 가격으로 다시 계산되지만 —
그 가격 축은 이미 전 구간이 같은 답을 준다는 것을 재생으로 확인했다(proposal의 표).
policy identity는 정의상 고정이다.

그러므로 **판정자가 바뀐 시점이 재시도가 새로운 정보를 만드는 유일한 시점**이다.
다른 어떤 시점의 재시도도 같은 산수를 다시 하는 것이고, 그것은 재시도가 아니라 소음이다.

## D2 — 기각한 대안

**시간 기반 재시도 (N시간마다).** 기각. 원인이 결정적이므로 같은 입력에 같은 답이
나온다. 원장에 격리·해제·재격리가 무한히 쌓이고, 그 소음 아래에서 *진짜로* 모호한
포지션과 *과거 결함의 잔재*를 구별할 수 없게 된다. 격리의 의미가 사라진다.

**매 관측 재시도.** 기각. 위와 같고 더 심하다. 격리는 "판정하지 않는다"는 상태인데
매번 판정하면 격리가 아니다.

**입력 변화 감지.** 기각. 격리된 세대는 입력을 갱신하지 않는다 — 그것이 격리의 정의다.
감지할 변화가 구조적으로 없다.

**배포 시각 비교 (빌드가 격리보다 새로우면 재시도).** 기각. 무관한 배포가 전부 재시도
사유가 된다. 콘솔 문구 하나 고친 배포가 손절 판정 경로의 격리를 푸는 근거가 되어서는
안 된다. 각인은 **그 판정을 내리는 코드**를 가리켜야 한다.

## D3 — 각인의 형태

```go
// exitpolicy
// RecoverySelectorRevision is the revision of the recovery-selection semantics.
// Bump it when SelectRecoverySnapshot, compareRecoveryStage or
// validateRecoverySnapshot changes what it accepts or refuses.
const RecoverySelectorRevision = 2
```

`1`은 a078이 고친 뒤의 현재 의미가 아니라 **a084 이전의 알 수 없는 과거** 전체를
가리키지 않는다 — 그 역할은 NULL이 한다. `2`는 a084가 배포하는 현재 의미다.
`1`은 예약해 두고 쓰지 않는다: 각인 없는 세계와 각인 있는 세계 사이에 번호를 하나
비워 두면, 나중에 원장을 읽는 사람이 "1은 어디 갔나"를 묻고 그 답이 이 문단이 된다.

**개정 번호를 자동 산출하지 않는다.** 선택 로직의 digest를 계산해 각인하는 방법도
있지만, 무관한 리팩터가 digest를 바꾸고 그때마다 전 포지션이 재판정된다. 손으로 올리는
번호는 "이 변경이 판정을 바꾼다"는 **판단**을 기록하는 것이고, 그 판단은 사람이 해야 한다.

**빠뜨렸을 때의 방향.** 선택기를 고치고 번호를 안 올리면 자동 재시도가 일어나지 않는다.
그것은 오늘의 동작 — 사람이 콘솔에서 푼다 — 이므로 fail-safe다. 반대 방향(번호를
잘못 올려 불필요한 재시도)도 안전하다: 재시도는 같은 선택기를 돌리고 같은 답이면 다시
격리한다. 어느 쪽으로 틀려도 새 위험이 생기지 않는다.

## D4 — NULL의 의미

`selector_revision`을 additive-nullable로 넣고 기존 행을 backfill하지 않는다(§0.6).

```
NULL   a084 이전에 기록됐다. 어느 개정에서 나왔는지 모른다. → 한 번 재시도한다.
n      개정 n이 내린 판정이다. → 현재 개정 ≠ n 일 때만 재시도한다.
```

NULL을 재시도 대상으로 두는 것이 이 change가 PLTR을 낫게 하는 방법이다. backfill로
현재 개정을 채워 넣으면 세 행 모두 "현재 판정자가 내린 판정"이 되어 영원히 재시도
대상에서 빠진다 — 사실이 아니고, 고치려는 문제를 그대로 둔다.

## D5 — 재시도가 판정을 약화하는가

**아니다. 재시도는 판정을 재개하는 것이고, 판정 자체는 조금도 느슨해지지 않는다.**

```
workingSet     각인이 다르면 refused로 빼지 않고 judge로 보낸다
judge          지금과 똑같이 SelectRecoverySnapshot을 돌린다
  성공         격리를 SELECTOR_REVISED로 닫고 판정을 기록한다
  거부         옛 행을 닫고 현재 개정으로 각인한 새 행을 연다. 아무것도 arm하지 않는다
```

거부 경로가 아무것도 arm하지 않는 것은 기존 코드의 성질이다. `RecordExitJudgement`는
`SelectRecoverySnapshot` 실패 시 격리를 쓰고 커밋한 뒤 `ErrExitSnapshotQuarantined`를
반환하며, 제안·arm은 그보다 뒤에 있다.

**성공 경로가 주문을 낼 수 있다.** 낼 수 있고, 내야 한다. 격리 동안 얼어 있던 것은
high water이지 시장이 아니다. PLTR을 실제 저장값으로 재생하면 현재가 159.62에서
rung 3, protection 151.6518, action `LADDER_HOLD_STOP_PROMOTED` — 주문 없이 보호선만
승격한다. 반대로 현재가가 저장된 보호선 아래라면 손절이 나가는데, 그것이 §0.3이
지연을 금지하는 바로 그 평가다. 지금은 그 평가가 16시간째 멈춰 있다.

**§0.9 방향 판정.** 이 change는 손절폭·익절폭·사이징을 바꾸지 않는다. 저장된 진입가·
최초 손절·보호선은 그대로 쓰이고, a079가 "release는 아무것도 고치지 않는다"고 적은
그 성질을 그대로 유지한다. 바뀌는 것은 **평가가 재개되는가**뿐이고, 재개는 보호가 있는
쪽이다.

## D6 — 순환하지 않음의 증명

재시도 조건은 `stamp != RecoverySelectorRevision`이다. 재시도가 거부로 끝나면 새 행이
`RecoverySelectorRevision`으로 각인된다. 그러므로 같은 개정에서 같은 세대는 **정확히
한 번** 재시도된다. 개정이 올라가면 다시 한 번. 상한은 개정 수이고 개정은 사람이 올린다.

`quarantine_version`은 기존대로 세대 안에서 단조 증가하므로 원장은 이렇게 읽힌다.

```
v1  ambiguous_recovery   selector NULL   quarantined 2026-08-04T13:31:05Z
                                         released    SELECTOR_REVISED  (재시도)
v2  ambiguous_recovery   selector 2      quarantined ...                (여전히 거부일 때만)
```

v2가 없으면 재시도가 성공한 것이다.

## D7 — 해제 kind

`ReleaseExitSnapshotQuarantine`은 kind를 `HUMAN_REPAIR` 또는
`AUTHORITATIVE_RECONCILE`로 제한한다. 세 번째를 더한다.

```go
QuarantineReleaseSelectorRevised = "SELECTOR_REVISED"
```

기존 두 kind에 얹지 않는다. `HUMAN_REPAIR`는 감사 추적에서 사람을 가리키므로 기계가
쓰면 거짓이 되고, `AUTHORITATIVE_RECONCILE`은 계좌 권위를 가리키므로 여기서는 무관하다.
"어느 근거로 풀렸나"가 원장에서 구별되어야 한다는 것이 kind를 두는 이유다.

## D8 — 어디서 해제하는가

재판정 성공 시의 해제는 **판정을 기록하는 같은 트랜잭션 안**이어야 한다. 밖에서 풀고
판정이 실패하면 격리 없이 판정도 없는 상태가 되고, 다음 관측이 그 포지션을 정상으로
착각한다.

`RecordExitJudgementResult`의 트랜잭션은 이미 `SelectRecoverySnapshot`을 돌리고
실패 시 `quarantineExitSnapshotTx`를 부른다. 성공 분기에 대칭으로
`releaseExitSnapshotQuarantineTx`를 두는 것이 최소 변경이고, 두 분기가 한 트랜잭션
안에서 대칭이 된다.

## D9 — RED이 되어야 하는 테스트

```
1. workingSet:  각인이 다른 활성 격리를 가진 포지션이 refused가 아니라 judge로 간다
2. workingSet:  각인이 같은 활성 격리는 지금과 똑같이 refused에 남는다
3. journal:     재판정 성공 시 격리가 SELECTOR_REVISED로 닫히고 판정이 기록된다
4. journal:     재판정 거부 시 옛 행이 닫히고 현재 개정으로 각인한 새 행이 열린다
5. journal:     새 행이 열린 뒤 같은 개정에서는 다시 재시도되지 않는다 (D6)
6. 원장 재생:   PLTR의 실제 저장값 + NULL 각인 → 재시도되고 recomputed가 선택된다
```

1과 3이 현재 코드에서 **실패**한다. 나머지는 회귀와 종결 조건이다.
