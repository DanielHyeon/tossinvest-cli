# a084 · Tasks

## 1. 근거 고정 (편집 전)

- [x] 1.1 `ExitObserver.workingSet`의 Function Logic Map과 Branch Test Map을
      **편집 전에** 작성한다. High-risk 경로이므로 면제 없다.
- [x] 1.2 `Journal.RecordExitJudgementResult`(격리 분기)의 Function Logic Map과
      Branch Test Map을 작성한다.
- [x] 1.3 Pre-Edit 선언을 `review.md`에 기록한다.
- [x] 1.4 CodeGraph로 `activeExitSnapshotQuarantineTx`, `quarantineExitSnapshotTx`,
      `ReleaseExitSnapshotQuarantine`의 호출부와 impact를 확인한다.
- [x] 1.5 활성 격리 3건의 저장 snapshot을 현재 비교기로 재생해 전 가격축 0건 격리를
      기록한다 (읽기 전용, proposal의 표).

## 2. 스키마

- [x] 2.1 `SchemaVersion` 29 → 30, `exit_snapshot_quarantines.selector_revision`
      additive-nullable 추가. backfill 없음.
- [x] 2.2 마이그레이션 테스트: 29에서 올라온 DB의 기존 행이 NULL을 유지한다.
- [x] 2.3 `QuarantineReleaseSelectorRevised` kind를 추가하고 기존 kind 검증을 넓힌다.

## 3. RED — exitpolicy

- [x] 3.1 `RecoverySelectorRevision`이 존재하고 양수다.
- [x] 3.2 개정 상수의 의미를 고정하는 문서 테스트 — 선택 의미가 바뀌면 올린다는 계약.

## 4. RED — journal

- [x] 4.1 각인이 다른 활성 격리에서 재판정이 성공하면 `SELECTOR_REVISED`로 닫히고
      판정이 기록된다. **현재 코드에서 실패한다.**
- [x] 4.2 재판정이 거부하면 옛 행이 닫히고 현재 개정으로 각인한 새 행이 열린다.
- [x] 4.3 판정 기록이 실패하면 해제도 함께 취소되고 격리가 활성으로 남는다 (원자성).
- [x] 4.4 각인이 같은 활성 격리는 지금과 같이 `ErrExitSnapshotQuarantined`로 끊는다.
- [x] 4.5 신규 격리는 현재 개정으로 각인된다.

## 5. RED — engine (workingSet)

- [x] 5.1 각인이 다른 활성 격리를 가진 포지션이 `refused`가 아니라 판정으로 간다.
      **현재 코드에서 실패한다.**
- [x] 5.2 각인이 NULL인 활성 격리도 판정으로 간다.
- [x] 5.3 각인이 같은 활성 격리는 지금과 같이 `refused`에 남고 판정되지 않는다.
- [x] 5.4 재판정 후 새 격리가 열리면 다음 주기에는 `refused`에 남는다 (순환 없음).
- [x] 5.5 회귀: `stored_snapshot_corrupt`·`legacy_policy_identity_unknown` 격리
      생성 경로는 지금과 동일하고 현재 개정으로 각인된다.

## 6. GREEN

- [x] 6.1 `exitpolicy.RecoverySelectorRevision` 추가.
- [x] 6.2 스키마·kind·읽기/쓰기 경로에 `selector_revision` 배선.
- [x] 6.3 `workingSet`의 활성 격리 분기에 개정 비교를 넣는다.
- [x] 6.4 `RecordExitJudgementResult` 성공 분기에 트랜잭션 내 해제를 넣고, 거부
      분기에 닫고-새로-여는 동작을 넣는다.
- [x] 6.5 3·4·5장 전부 GREEN.

## 7. REFACTOR

- [x] 7.1 `recovery.go`에 개정 번호를 언제 올리는지, 빠뜨렸을 때 왜 fail-safe인지 적는다.
- [x] 7.2 `exit_snapshot.go`에 NULL 각인의 의미를 적는다.
- [x] 7.3 a078 `issues.md` I1의 설계 공백이 닫혔음을 이 change의 `issues.md`에 기록한다.

## 8. VERIFY

- [x] 8.1 변이 검증: 개정 비교를 `==`로 뒤집으면 5.1이 RED가 되는지 확인하고 되돌린다.
- [x] 8.2 변이 검증: 거부 분기의 재각인을 빼면 5.4가 RED가 되는지 확인하고 되돌린다.
- [x] 8.3 PLTR의 실제 저장값(entry 146.1 / stop 141.717 / high 148.55, 각인 NULL)으로
      재판정이 `recomputed`를 선택하는지 확인한다 (읽기 전용 fixture).
- [x] 8.4 배포 전 `main`과 `SchemaVersion` 대조.
- [x] 8.5 upstream 상속 테스트 회귀 없음 (650 green).
- [x] 8.6 `make test`, `make vet`, `make validate`, `make sdd-sync`, `make sdd-check`.
- [x] 8.7 `make gate CHANGE=a084-a-quarantine-outlives-its-cause`.

## 9. 리뷰와 기록

- [x] 9.1 적대적 Eng 리뷰를 받고 `review.md`에 기록한다.
- [x] 9.2 `issues.md`에 남긴다 — `ambiguous_recovery`의 evidence가 접미사 없는
      sentinel이라 어느 분기에서 나왔는지 사후에 알 수 없다는 점.

## 10. 개정 2 — 독립 리뷰가 연 blocking (2026-08-05)

- [x] 10.1 D7 RED: 롤백한 빌드가 더 새로운 선택기의 거부를 뒤집는 것과,
      선택기가 결정하지 않은 사유의 격리가 재판정 대상이 되는 것을 고정한다.
      `a084b_rejudgement_eligibility_test.go`.
- [x] 10.2 D7 GREEN: `NeedsReJudgement`를 사유(`ambiguous_recovery`)와 방향(`<`)으로
      게이트하고, `ReleaseExitSnapshotQuarantine`의 `SELECTOR_REVISED` 확장을 되돌린다.
- [x] 10.3 D8 RED: 가격이 움직이지 않은 재판정이 격리 행을 각인하지 않아 영구히
      반복된다는 것을 고정한다. `a084b_rejudge_bound_test.go`.
- [x] 10.4 D8 GREEN: `StampExitSnapshotQuarantineSelector`를 `workingSet`의 통과
      결정 지점에서 부르고, `!snapshot.Changed` 조기 반환이 재판정을 건너뛰지 않게 한다.
- [x] 10.5 D9 RED: 재판정 통과가 판정보다 먼저 실주문을 취소한다는 것을 고정한다.
- [x] 10.6 D9 GREEN: 재판정 통과는 clear를 시도하지 않고 기존 arm-suppression 계약
      (`noteDelay` + `ArmSuppressedReJudge`)을 재사용한다.
- [x] 10.7 중복 제거: `QuarantineExitSnapshot`이 `quarantineExitSnapshotTx`를 부르게 한다.
- [x] 10.8 I10: CI가 `make test`를 돌리고 `cover` 타깃도 timeout을 받는다.
- [x] 10.9 `issues.md`에 B1~B4와 I9~I11을 기록한다.
- [ ] 10.10 gstack 독립 리뷰 재실행 + `make gate`.
