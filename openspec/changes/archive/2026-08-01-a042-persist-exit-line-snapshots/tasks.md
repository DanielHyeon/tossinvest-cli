## 1. 스키마 분석과 RED

- [x] 1.1 a041 완료를 확인하고 journal migration/recovery 함수의 Function Logic·Branch Test Map과 rollback 계획을 작성한다.
- [x] 1.2 additive migration, atomic/coherent snapshot, legacy NULL, impossible mixed tuple, commit 전후 crash와 monotone reopen RED 테스트를 추가한다.

## 2. 영속화와 복구

- [x] 2.1 nullable snapshot schema와 typed journal read/write 및 saved/recomputed/effective provenance read model을 구현하고, a041의 `PolicyIdentity`/`ExitDecisionProvenance` seam을 저장·복원한다. legacy NULL을 현재 registry digest로 backfill하지 않는다.
- [x] 2.2 evaluation event와 snapshot을 원자적으로 commit하고 recovery에 max-safe 규칙을 연결한다. event에는 saved-before, recomputed, effective, effective source와 projected quantity/ratio/state-only/suppressed reason을 함께 보존한다.
- [x] 2.3 corrupt snapshot은 현재 position generation만 격리하고 명시적인 repair/reconcile로만 해제하며, 알림 실패와 무관하게 다른 emergency exit가 진행되는 테스트를 통과한다.

## 3. 검증

- [x] 3.1 DB/WAL/SHM 보존, pre-migration backup 원자 복원, broker reconcile, reopen, fault-injection, full test·vet·validate를 통과한다.
- [x] 3.2 적대적 Eng 독립 리뷰와 `make gate CHANGE=a042-persist-exit-line-snapshots`을 통과한다.
