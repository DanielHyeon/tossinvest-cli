# a166 time namespace 로트 — 계약 개정 문안(동결 계약의 개정, Manager 승인 전 본문 미반영)

a166 은 4판 proposal-freeze READY 상태다. 아래는 **문안만**이고 design.md · spec delta · proposal.md ·
`analysis/poll-clock-contract.md` 본문은 Manager 승인 뒤 2단계에서 고친다. tasks.md 는 건드리지 않는다
(새 task 줄이 필요하면 Manager 가 정한다 — 결정 4). 근거: `investigation.md` §2–§5.

## 1. design.md 77행(재시작 · hint 복원 문단)

현재:
> 같은 boot/time namespace에서는 주입된 공유 boottime 표본과 duration으로 남은 시간을 구한다.
> … clock/namespace/레코드·복원 anchor의 무결성을 확인할 수 없으면 해당 engine poll을 fail-closed로 유지하고 오류를 드러낸다.

개정:
> 같은 boot에서는 주입된 공유 boottime 표본과 duration으로 남은 시간을 구한다. 각 표본은 채취하는 프로세스의
> 현재·자식 time namespace가 같고 monotonic·boottime offset이 모두 0임을 확인한 것이다. offset 0인 time namespace의
> CLOCK_BOOTTIME은 초기 namespace와 같은 시계이므로, 참여자마다 time namespace 신원(device/inode)이 달라도 같은 boot의
> 표본끼리 비교한다. namespace 신원은 한 표본 안의 앞뒤 일관성 검사와 진단에만 쓰고 프로세스 간 비교 조건으로 쓰지
> 않는다. … clock(offset 비영·namespace 표본 중 변경 포함)/레코드·복원 anchor의 무결성을 확인할 수 없으면 해당 engine
> poll을 fail-closed로 유지하고 오류를 드러낸다.

나머지 문장(다른 boot 보수 보류, anchor 재사용, wall-time 상한 없음)은 그대로.

## 2. design.md 101행(provisioning 점검)

현재: 「directory/파일 소유권·mode·symlink/hardlink·regular-file·fs 종류·device/inode·boot/time namespace를 검증한다.」

개정: 「… device/inode·boot·time namespace offset 0(현재·자식 namespace 동일)을 검증한다. 참여자 사이 time namespace
신원 일치는 요구하지 않는다.」

## 3. proposal.md 24행(제약)

현재: 「제약: Linux의 지원 filesystem/time namespace, …」
개정: 「제약: Linux의 지원 filesystem, monotonic·boottime offset 0인 time namespace(참여자마다 namespace가 달라도 된다), …」

## 4. analysis/poll-clock-contract.md

- 「첫 지원 범위」: 「실행 중 setns/unshare/time namespace 교체 없음」 뒤에 덧붙임 — 「단 같은 boot의 서로 다른
  offset 0 namespace에서 채취한 표본끼리의 비교는 지원한다(2026-11-25 B1, 사용자 결정 1번).」
- 「복원 산술 계약」 첫 두 문장 교체:
  > 동일 boot이면(namespace 신원과 무관) now.uptime >= saved.uptime일 때만 차이를 빼며 음의 남은 시간은 0으로
  > 포화한다. 동일 boot인데 uptime이 역행하면 오류다. 다른 boot에서는 현재 uptime을 빼지 않는다.
- 「수용 증거」 끝에 덧붙임: 「다른 namespace 양성 사례와 실제 커널 time namespace 재현
  (`analysis/time-namespace-lot/run-*.log`)을 수용 증거로 남긴다.」

## 5. spec delta — `specs/order-execution/spec.md`

Requirement 본문(8행)은 바꾸지 않는다. 기존 시나리오 하나를 보강하고 하나를 추가한다 → **Requirement 수준 수정**
(WORKFLOW 리뷰 등급 「requirement 변경」).

기존(85–87행) 보강:
```
#### Scenario: 유효한 폴링 hint와 프로세스 재시작
- **WHEN** 유효한 30초 이하 폴링 hint가 남은 상태에서 엔진 프로세스가 재시작하면(컨테이너 재생성으로 time namespace가 바뀐 경우 포함)
- **THEN** 해당 종류의 남은 제한을 복원하며, 무관한 주문이나 모든 계좌 요청에 새 30초/5분 기동 패널티를 추가하지 않는다
```

추가:
```
#### Scenario: 다른 time namespace 참여자가 남긴 폴링 hint
- **WHEN** 같은 boot에서 monotonic·boottime offset이 0인 서로 다른 time namespace의 참여자가 남긴 유효 hint를 다른 참여자가 복원하면
- **THEN** namespace 신원 차이만으로 거절하지 않고 같은 boottime 시계로 남은 시간을 구하며, 만료된 hint는 지워진다. offset이 0이 아니거나 현재·자식 namespace가 다르거나 표본 중 namespace가 바뀌면 표본을 만들지 않고 해당 poll을 fail-closed로 유지한다
```

## 6. 리뷰 등급

- WORKFLOW 「requirement 변경」: 수정분에 대한 gstack 리뷰 재실행. a166 은 주문 경로 인접(엔진 폴링 · 보호 가격 조회
  보류)이므로 적대적 Eng 보이스 포함. 기록은 `time-namespace-lot/review.md`(이 로트), 2단계에서 change `review.md`
  에 요약을 옮긴다.
