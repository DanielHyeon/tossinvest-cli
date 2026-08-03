# a063 · Tasks

## 1. 근거 고정 (편집 전)

- [x] 1.1 `ReleaseExitSnapshotQuarantine`의 프로덕션 호출자가 0임을 현재 HEAD에서 확인해 기록한다.
- [x] 1.2 격리를 만드는 3경로와 읽는 1경로를 열거해 기록한다 (proposal 표의 근거).
- [x] 1.3 운영 원장의 활성 격리를 읽기 전용으로 확인해 기록한다 (계좌 식별자는 남기지 않는다).
- [x] 1.4 편집 대상 기존 함수 목록을 확정하고 각각 Function Logic Map을 **편집 전에** 만든다.
- [x] 1.5 Pre-Edit 선언을 `review.md`에 기록한다.

## 2. 계약 — `internal/exitquarantine`

- [x] 2.1 `Row`/`Request`/`Preview`/`ApplyRequest`/`Result`와 오류 enum을 정의한다.
      capability-neutral: journal·broker·config·HTTP 의존이 없다.
- [x] 2.2 `Request.Validate`는 positionID·generation≥0·version>0·actor enum을 요구한다.
- [x] 2.3 잘못된 actor, version 0, 빈 positionID가 거부된다.

## 3. 원장 read — 활성 격리 목록

- [x] 3.1 `ActiveExitSnapshotQuarantines`가 현재 세대와 일치하는 미해제 격리만,
      market·symbol과 함께 반환한다.
- [x] 3.2 해제된 행과 세대가 어긋난 행은 제외된다.
- [x] 3.3 읽을 수 없는 snapshot은 목록을 죽이지 않고 이유를 남긴다.
- [x] 3.4 새 파일로 구현한다. 기존 journal 함수는 편집하지 않는다.

## 4. engine command service

- [x] 4.1 `Preview`가 활성 격리·저장된 보호선·capability·대기 시간을 반환한다.
- [x] 4.2 활성 격리가 없으면 `ErrNotQuarantined`로 거부한다.
- [x] 4.3 `Release`는 확인 없이 거부하고 capability를 소비하지 않는다.
- [x] 4.4 `Release`는 3초 지연 전에 `ErrCapabilityTooEarly`로 거부한다.
- [x] 4.5 capability는 1회용이고 재사용·만료·위조는 거부된다.
- [x] 4.6 preview 이후 quarantine version이 바뀌면 `ErrVersionMismatch`로 거부한다.
- [x] 4.7 성공 경로가 `ReleaseExitSnapshotQuarantine(HUMAN_REPAIR)`를 서버 조립 evidence로 호출한다.
- [x] 4.8 해제가 `exit_states`·`positions`·`position_policy_lifecycles`를 바꾸지 않는다
      (세 테이블 전체 덤프 비교).
- [x] 4.9 격리 capability가 없는 control plane은 `ErrUnwired`를 반환한다.

## 5. RPC transport

- [x] 5.1 `/v1/quarantines`·`/v1/quarantine/preview`·`/v1/quarantine/release`가 bearer를 요구한다.
- [x] 5.2 잘못된 method는 405다.
- [x] 5.3 client 왕복이 preview·release 결과를 보존한다.
- [x] 5.4 capability 없는 서비스에는 세 라우트가 등록되지 않는다 (§0.2).

## 6. 콘솔

- [x] 6.1 격리된 행에 「판정 격리 해제」 동작이 나오고, 격리가 없으면 나오지 않는다.
- [x] 6.2 preview 화면이 사유·증거·격리 시각·유지되는 보호선·재격리 경고를 보여준다.
- [x] 6.3 확인 체크박스와 3초 지연이 있고 **타이핑 확인 입력이 없다**.
- [x] 6.4 commander 미배선이면 조회 전용으로 뜨고 동작이 나오지 않는다.
- [x] 6.5 오류가 사용자 언어로 매핑된다 (version 충돌·만료·너무 이름·격리 없음·미배선).
- [x] 6.6 격리 조회 실패가 화면 전체를 죽이지 않는다.
- [x] 6.7 정책 select token을 격리 해제로 재생할 수 없다.
- [x] 6.8 두 mutation 라우트가 CSRF 게이트 아래에 등재된다.

## 7. 판정 재개와 재격리

- [x] 7.1 해제된 포지션이 다음 관측에서 판정되고 기준선이 그대로다.
- [x] 7.2 해제는 원장의 재격리 능력을 소모하지 않는다 — 같은 세대가 새 version으로
      다시 격리되고, 옛 해제 재생은 `ErrExitSnapshotReleaseStale`이다.

## 8. GREEN · REFACTOR

- [x] 8.1 2~7장 구현.
- [x] 8.2 자기 seam 결정(D1)·evidence 서버 조립(D3)·세대 규칙 단일 위치를 코드 주석으로 남긴다.

## 9. VERIFY

- [x] 9.1 변이: `findQuarantine`의 version 비교 제거 → 4.6 RED. (부산물: `ReleaseQuarantine`
      안의 두 번째 version 비교가 **도달 불가능한 죽은 가드**임을 변이로 발견해 제거했다.)
- [x] 9.2 변이: `Confirmed` 검사 제거 → 4.3 RED.
- [x] 9.3 변이: 원장 read의 세대 술어 제거 → 3.2 RED.
- [x] 9.4 변이: 콘솔 token purpose 검사 제거 → 6.7 RED.
- [x] 9.5 변이: 위험 지연 0으로 → 4.4 RED.
- [x] 9.6 변이: 해제가 exit state를 건드리도록 → 4.8 RED.
- [x] 9.7 `AUTHORITATIVE_RECONCILE` 프로덕션 호출자가 여전히 0임을 확인한다.
- [x] 9.8 upstream 상속 테스트 회귀 없음.
- [x] 9.9 `make test`, `make vet`, `make validate`, `make sdd-sync`, `make sdd-check`.
- [x] 9.10 `make gate CHANGE=a063-operator-can-lift-a-quarantine`.

## 10. 리뷰와 기록

- [x] 10.1 적대적 Eng 리뷰를 받고 `review.md`에 기록한다.
- [x] 10.2 발견 사항을 `issues.md`에 남긴다.
- [x] 10.3 PM story/tracker 동기화.

## 11. 사람 승인 후 운영 적용

- [ ] 11.1 배포 후 콘솔에서 활성 격리의 badge와 해제 화면이 실제로 뜨는지 실측한다.
- [x] 11.2 실제 해제 실행은 사용자 판정임을 기록한다 (§0.7). 에이전트가 실행하지 않는다.
