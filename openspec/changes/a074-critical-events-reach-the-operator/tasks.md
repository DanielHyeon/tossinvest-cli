# a074 · Tasks

## 1. 근거 고정 (편집 전)

- [x] 1.1 운영 원장에서 격리 생성 시각과 첫 알림 시각의 간격을 읽기 전용으로 확인해 기록한다.
- [x] 1.2 `ExitCycle.Err`가 프로덕션에서 읽히지 않음을 현재 HEAD에서 확인해 기록한다.
- [x] 1.3 `alert_outbox`의 `attempts=0`이 "전송 실패"가 아니라 "시도 없음"임을 코드로 확인한다.
- [x] 1.4 편집 대상 기존 함수 6개를 확정하고 각각 Function Logic Map을 **편집 전에** 만든다.
- [x] 1.5 Pre-Edit 선언을 `review.md`에 기록한다.

## 2. 사이클 실패 보고

- [x] 2.1 실패한 사이클의 사유가 error 등급 구조화 로그로 남는다.
- [x] 2.2 성공한 사이클은 아무것도 남기지 않는다 (5초 주기 로그 스팸 금지).
- [x] 2.3 사이클 실패만으로는 알림도 운영 모드 강화도 일어나지 않는다.
- [x] 2.4 `Run`은 여전히 사이클 실패로 반환하지 않는다 (기존 계약 유지).

## 3. 격리 생성 이벤트

- [x] 3.1 `obs.EventExitSnapshotQuarantined`를 정의하고 `criticalEvents`에 등록한다.
- [x] 3.2 이벤트가 포지션·세대·version·사유·증거·격리 시각을 싣는다.
- [x] 3.3 `stored_snapshot_corrupt` 경로가 발행한다.
- [x] 3.4 `legacy_policy_identity_unknown` 경로가 발행한다.
- [x] 3.5 `ambiguous_recovery` 경로가 **같은 사이클에** 발행한다 — 되읽기로 얻는다.
- [x] 3.6 이미 `exit.judgement_refused`가 latch된 포지션에서도 발행된다.
- [x] 3.7 같은 quarantine version은 프로세스 수명 동안 한 번만 발행된다.
- [x] 3.8 새 version(해제 후 재격리)은 다시 발행된다.
- [x] 3.9 `RecordExitJudgementResult`와 `quarantineExitSnapshotTx`를 편집하지 않는다.

## 4. 알림 설정 블록

- [x] 4.1 `config.Engine.Notifications`가 추가되고 zero value가 off다.
- [x] 4.2 알림 블록이 없는 설정 파일이 오늘과 같이 파싱된다 (§0.2).
- [x] 4.3 거부된 블록은 부분 적용되지 않고 전체가 0이 되며 사유가 남는다.
      (채널 부재 판정은 환경을 봐야 하므로 5.5의 조립 단계가 갖는다.)
- [x] 4.4 `base_url`이 http/https가 아니면 거부된다.
- [x] 4.5 `internal/config`는 환경변수를 읽지 않는다. 토큰을 담을 필드가 아예 없다.
- [x] 4.6 automation_gate 블록이 없는 파일에서도 알림 블록이 병합된다.

## 5. Publisher 조립

- [x] 5.1 설정 off면 Publisher가 nil이고 오늘 동작과 동일하다.
- [x] 5.2 설정 on이면 `obs.Ntfy`가 base URL·topic·token으로 조립된다.
- [x] 5.3 `TOSSCTL_NTFY_TOKEN`이 토큰을 공급한다.
- [x] 5.4 `TOSSCTL_NTFY_TOPIC`이 파일의 topic보다 우선한다.
- [x] 5.5 거부된 설정은 기동을 막지 않고 publisher 없이 기동한다.
- [x] 5.6 `opts.Publisher`가 명시된 경우(테스트 주입)가 설정보다 우선한다.
- [x] 5.7 ntfy.sh 공개 서비스를 향하면 경고가 남는다.

## 6. Audit

- [x] 6.1 알림 설정 4항목이 audit에 기록된다.
- [x] 6.2 topic 값과 token 값이 audit에 **나타나지 않는다**.
- [x] 6.3 거부 사유가 audit에 기록된다.
- [x] 6.4 기존 audit 항목의 순서·이름이 바뀌지 않는다.

## 7. 전달 경로 통합

- [x] 7.1 publisher가 배선되면 critical 알림이 outbox 기록 후 실제로 전송되고 delivered로 표시된다.
- [x] 7.2 전송 실패는 기존 재시도·gate latch·ENTRY_BLOCKED 경로를 그대로 탄다.
- [x] 7.3 알림 본문·헤더에 토큰이 실리지 않는다.

## 8. GREEN · REFACTOR

- [x] 8.1 2~7장 구현.
- [x] 8.2 D1(로그 vs 알림)·D3(되읽기)·D5(값 대신 설정 여부)를 코드 주석으로 남긴다.

## 9. VERIFY

- [x] 9.1 변이: 사이클 실패 로그 제거 → 2.1 RED.
- [x] 9.2 변이: 격리 생성 이벤트를 normal 등급으로 → 3.1 RED.
- [x] 9.3 변이: in-process latch 제거 → 3.7 RED.
- [x] 9.4 변이: latch 키에서 version 제거 → 3.8 RED.
- [x] 9.5 변이: 거부된 알림 설정을 부분 적용 → 4.3 RED.
- [x] 9.6 변이: audit에 topic 값 기록 → 6.2 RED.
- [x] 9.7 변이: 설정 off에서도 publisher 조립 → 5.1 RED.
- [x] 9.8 upstream 상속 테스트 회귀 없음.
- [x] 9.9 `make test`, `make vet`, `make validate`, `make sdd-sync`, `make sdd-check`.
- [x] 9.10 `make gate CHANGE=a074-critical-events-reach-the-operator`.

## 10. 리뷰와 기록

- [x] 10.1 적대적 Eng 리뷰를 받고 `review.md`에 기록한다.
- [x] 10.2 발견 사항을 `issues.md`에 남긴다.
- [x] 10.3 PM story/tracker 동기화.

## 11. 사람 승인 후 운영 적용

- [ ] 11.1 배포 후 격리 생성 이벤트가 실제 사이클에서 발행되는지 실측한다.
- [x] 11.2 알림 전송 설정을 켜는 것은 사용자 판정임을 기록한다 (§0.7). 에이전트가 켜지 않는다.
