# multi-horizon-risk-buckets — a126 delta

> 사용량 정의(D5 — "Position 에 귀속된 filled 노출")에 수명주기를 세운다. 떠나는 사건·
> 소급 여부·부분 종결 귀속(Q1·Q2·Q4)은 freeze 가 스키마 영수증으로 확정하며, 이 델타의
> 문장은 그 확정에 맞춰 정밀화될 수 있다.

## ADDED Requirements

### Requirement: 체결 노출은 귀속 포지션과 함께 bucket 을 떠난다

포지션에 귀속된 체결 노출이 원장 사실로 소멸하면 그 몫의 `filled_minor` 는 그 포지션이 점유했던 모든 bucket 의 사용량에서 빠져야 한다(SHALL — 떠나는 사건은 freeze 가 Q1 로 확정).
사용량 감소는 그 원장 사실에 결속되지 않고는 일어나서는 안 된다(SHALL NOT —
느슨한 감소는 진입 cap 을 여는 fail-open 이다).

이 수명주기는 자동 회계이며 운영자 완화(entry-lock·overage latch 해제)와 별개의 축이다.
감소가 overage latch 재계산과 만나는 순서는 정의되어야 한다(SHALL — Q3).

#### Scenario: 포지션 종결이 사용량을 줄인다

- **WHEN** 어떤 포지션의 귀속 체결 노출을 소멸시키는 원장 사건이 커밋된다
- **THEN** 그 포지션이 점유했던 각 bucket 의 `filled_minor` 에서 그 몫이 빠지고, 같은
  한도에서 진입 가능량이 그만큼 회복된다

#### Scenario: 원장 사실 없는 감소는 없다

- **WHEN** 대응하는 원장 소멸 사건 없이 사용량 감소가 시도된다
- **THEN** 감소는 일어나지 않는다
