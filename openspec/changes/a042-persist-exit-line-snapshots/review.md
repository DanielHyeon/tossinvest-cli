# Review: a042-persist-exit-line-snapshots

- Date: 2026-07-31
- Voices: Security, Test Architecture, Maintainability

## Findings and decisions

1. field별 `max`는 불가능한 snapshot을 만든다. 같은 policy digest의 검증된 coherent candidate 하나만 선택하고, 모호하면 포지션을 격리한다.
2. policy/snapshot/decision/observation identity를 exit event와 같은 transaction에 저장한다.
3. 현재 구 binary는 높은 schema version을 거부한다. rollback은 engine stop, DB/WAL/SHM 보존, pre-migration backup 복원과 broker reconcile이다.
4. migration crash, WAL reopen, commit/ack window와 emergency exit lock latency를 fault test로 고정한다.

## Verification evidence

- OpenSpec strict validation: pass.
- Schema plan: additive nullable, journal migration v10 reserved.

## Verdict

수정된 coherent recovery와 fail-closed downgrade 계약으로 구현을 승인한다. a041 실제 통합 commit을 base로 캡처해야 한다.
