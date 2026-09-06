# Pre-Edit Gate — journal snapshot-only lineage

- Change/tasks: `a064-add-multi-market-strategy-evidence` 2.5, 3.3, 5.2, 5.3
- Existing symbols: `journal.insertExactStrategyDecision`, `journal.(*ReadOnly).checkSchema`
- Current evidence: HEAD `75cb371a`; CodeGraph definitions, callers, context and impact inspected before editing.
- Direct callers: strategy test planner and production `RecordStrategyDecisionAndReserve`; read schema checker is called only by `OpenReadOnly`.
- Existing tests: exact/idempotent/collision/rollback strategy lineage, schema version/direction, damaged v20 campaign schema and compile-time read-only method set.
- Upstream impact: yes. Preserve blank legacy lineage, immutable exact replay, transaction rollback and released schema diagnostic ordering.
- RED first: required.
- Safety invariant review: passed. The change adds no payload/revision/credential storage and no Guardian, dispatch, broker, apply-hook or operating-toggle authority.
