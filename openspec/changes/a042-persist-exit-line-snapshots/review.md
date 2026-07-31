# Review: a042-persist-exit-line-snapshots

- Date: 2026-08-01
- Voices: Security, Test Architecture, Maintainability

## Findings and decisions

1. field별 `max`는 불가능한 snapshot을 만든다. 같은 policy digest의 검증된 coherent candidate 하나만 선택하고, 모호하면 포지션을 격리한다.
2. policy/snapshot/decision/observation identity를 exit event와 같은 transaction에 저장한다.
3. 현재 구 binary는 높은 schema version을 거부한다. rollback은 engine stop, DB/WAL/SHM 보존, pre-migration backup 복원과 broker reconcile이다.
4. migration crash, WAL reopen, commit/ack window와 emergency exit lock latency를 fault test로 고정한다.
5. `NextTarget`/`NextProtection`은 output digest를 권위로 삼지 않는다. evaluation event에 exact
   ratchet config + real-breakeven + taken ratio 또는 exact ladder table을 저장하고, policy
   ID/version/digest가 그 정의와 일치할 때만 선택 policy로 두 값을 다시 파생해 exact compare한다.
   whole-output digest는 이 semantic derivation 뒤에 적용하는 accidental-corruption seal이다.
6. v9 legacy row는 기존 `policy_id`가 non-NULL이라는 이유만으로 partial-v10으로 분류하지 않는다.
   a041의 pinned legacy RATCHET/common-policy meaning은 read-time typed identity로만 복원하고 DB의
   NULL version/digest는 유지한다. adoption context가 필요한 RUNNER는 position record로 resolve하며,
   unknown legacy ID만 현재 generation을 durable quarantine한다.
7. LADDER의 첫 rung 이전 상태는 `ActiveRung=-1`이 정상이다. exact ladder table로 첫 target/stop을
   다시 파생하며 `-2` 이하나 rung 수 이상의 값은 거부한다.
8. legacy 판정은 `policy_id`를 제외한 모든 v10 컬럼이 NULL일 때만 허용한다. SEED에 output 컬럼
   하나라도 남으면 `partial_seed_tuple`, status가 NULL인데 하나라도 남으면 `partial_snapshot_tuple`이다.
9. orderable snapshot의 proposal은 action/level이 exact evaluator output과 같아야 한다. 단, working
   order를 제거하지 못한 경우에는 snapshot을 비-orderable로 변조하지 않고 arm만 보류하며
   `working_order_not_cleared`를 event/read model에 기록한다. 알 수 없는 사유는 fail-closed다.
10. ratio 범위·level·whole-share projection·state-only/orderable/suppression·policy-kind/action·rung
    bounds를 semantic recovery validation에서 독립 검증한다.
11. 실제 Linux subprocess를 transaction 내부와 commit 직후에 SIGKILL하여 rollback/생존 경계를
    검증한다. hook error rollback은 이 테스트의 대체물이 아니다.

## Verification evidence

- OpenSpec strict validation: pass (`openspec validate a042-persist-exit-line-snapshots --strict --no-interactive`).
- Schema v10: additive nullable snapshot/event columns, generation-scoped quarantine ledger.
- Function Logic Map: `python3 tools/logic-map/check_analysis.py --change a042-persist-exit-line-snapshots` pass.
- SDD: `make sdd-check` pass. CodeGraph hard evidence matches; CodeGraphContext/GBrain freshness is advisory WARN only.
- Full verification: `make test`, `make lint`, and focused race suite for recovery/journal/engine pass.
- Atomicity/reopen: `after_state`, `after_arm`, `after_event` injected failures reopen to the prior complete
  state; actual Linux SIGKILL before commit leaves the prior state/event/proposal and SIGKILL after commit
  preserves state/event/arm together; successful saved-monotone recovery reopens to the exact selected tuple.
- Migration/rollback: failed v10 DDL rolls back columns/table/metadata/user_version, names the pre-migration
  backup and broker reconcile instruction, and the backup-only restore migrates to v10 with all v9 rows.
  Existing detached-WAL test and the new snapshot SIGKILL artifact checks pin DB/WAL/SHM preservation.
- Adversarial recovery: crossed axes, immutable entry/stop/policy drift, forged derived line with a newly
  requested digest, trailing JSON, partial tuple, duplicate decision, stale/future observation, and release
  CAS/evidence cases pass.
- Proposal coherence: missing/mismatched action or level and unknown arm-suppression reason fail before a
  write; the known uncleared-working-order branch persists the exact orderable snapshot, no arm, and typed
  event/read-model evidence.
- Emergency isolation: a synchronously blocked corruption alert occurs only after another position's
  emergency proposal reaches the submit seam.

## Verdict

구현·자체 검증은 완료했다. task 3.2의 적대적 Eng 독립 리뷰와 최종 gate 판정은 메인 에이전트가
수행하므로 이 작업 브랜치에서는 승인/체크하지 않는다.
