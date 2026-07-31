# Review: a042-persist-exit-line-snapshots

- Date: 2026-08-01
- Voices: Security, Test Architecture, Maintainability

## Findings and decisions

1. field별 `max`는 불가능한 snapshot을 만든다. 같은 policy digest의 검증된 coherent candidate 하나만 선택하고, 모호하면 포지션을 격리한다.
2. policy/snapshot/decision/observation identity를 exit event와 같은 transaction에 저장한다.
3. 현재 구 binary는 높은 schema version을 거부한다. rollback은 engine stop, DB/WAL/SHM 보존, pre-migration backup 복원과 broker reconcile이다.
4. migration crash, WAL reopen, commit/ack window와 emergency exit lock latency를 fault test로 고정한다.
5. output digest는 권위가 아니다. evaluation event에 원래 `SnapshotContext`와 ratchet/ladder
   evaluator input, 직전 high-water/protection/stage를 저장하고 같은 pure evaluator와
   `ChangedFromState`를 다시 실행한다. 모든 `ExitLineSnapshot` 필드가 exact-match일 때만 복구한다.
   whole-output digest는 이 semantic replay 뒤에 적용하는 accidental-corruption seal이다.
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
10. 수량·level·보호선·whole-share projection·state-only/orderable/suppression·policy-kind/action·rung,
    `CancelPendingFirst`, `Changed`까지 exact evaluator replay로 검증한다. 같은 정수 projection이
    나오는 수량 변경도 input/identity 불일치로 거부한다.
11. 실제 Linux subprocess를 transaction 내부와 commit 직후에 SIGKILL하여 rollback/생존 경계를
    검증한다. hook error rollback은 이 테스트의 대체물이 아니다.
12. engine은 요청에 담긴 proposal을 제출 권한으로 보지 않는다. journal transaction commit 뒤
    반환된 `ExitArmArmed` + `ArmedProposal`만 제출하며, saved-monotone 선택은 typed no-arm 결과다.
13. EVALUATED의 `projected_quantity`와 `state_only`가 NULL이면 partial tuple이다. legacy judgement는
    arm-suppression reason을 가질 수 없고, event read는 알려진 enum과 완전한 orderable evidence를
    함께 검증한다.
14. event read는 `effective_source`에 따라 effective JSON이 recomputed 또는 saved candidate와
    exact-match하고 `SelectRecoverySnapshot` 결과도 같아야 한다. orderable recomputed의 reason 삭제,
    다른 valid JSON 치환, partial tuple, saved source 위조, non-orderable suppression은 모두 corruption이다.
15. recovery evidence의 직전 상태는 별도 중복 필드로 저장하지 않고 exact evaluator input의
    high-water/baseline/level 또는 activated rung을 직접 사용한다.
16. event legacy 판정은 20개 v10 컬럼이 모두 NULL일 때만 허용한다. evaluation event는 generation,
    policy/decision/snapshot/observation identity, observation metadata, projection/state-only와 JSON/source가
    완전해야 하며 optional next line/ratio/suppression도 NULL-vs-value까지 recomputed candidate와 일치한다.

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
- Durable execution authority: saved-monotone recovery over an orderable stale recomputation returns no
  armed proposal, leaves pending false, preserves saved effective state, and retains both candidates in audit.
- Exact semantic replay: forged ratchet level, ladder protection, same-floor remaining quantity,
  cancel-first and changed bits all fail closed; evaluated NULL state-only/projected fields and forged event
  arm-suppression evidence are typed corruption.
- Event state-machine read: deleted/empty suppression reason, swapped valid effective JSON, forged source,
  missing effective candidate, known reason on non-orderable evidence, and forged armed action all fail closed
  in strict and per-row read-only projections.
- Flattened event evidence: every v10 column as sole lifecycle evidence and every required NULL/identity-policy-
  observation-projection mismatch on evaluation events fail closed in both readers; `decision_id` loss can no
  longer be hidden while weakening dedup evidence.
- Emergency isolation: a synchronously blocked corruption alert occurs only after another position's
  emergency proposal reaches the submit seam.

## Verdict

구현·자체 검증은 완료했다. task 3.2의 적대적 Eng 독립 리뷰와 최종 gate 판정은 메인 에이전트가
수행하므로 이 작업 브랜치에서는 승인/체크하지 않는다.
