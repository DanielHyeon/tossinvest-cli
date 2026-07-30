# CodeGraph baseline

Date: 2026-07-31
Base commit: `1dbef864038c81f0a2982a03e7d9549369d21669`

## Hard evidence

- `tools/pm/generate_master_tracker.py::validate` is the PM authority used by
  `make sdd-check`.
- The validator checks Initiative → Epic → Feature → Story forward/reverse links,
  then builds a `by_change` map from each Story's flat `change_id`.
- Active OpenSpec changes missing from `by_change` pass when listed in
  `_registry.yaml.bootstrap_change_allowlist`.
- Story lifecycle is read from manually stored `status`; it is not derived from
  proposal, task, or archive evidence.
- `render` also reads manual status and the flat `change_id`, so generated trackers
  can reproduce stale portfolio claims.

## Blast radius

- `tools/pm/generate_master_tracker.py`
- `tools/pm/test_generate_master_tracker.py`
- `docs/pm/portfolio/**`
- `docs/pm/generated/**`
- `docs/WORKFLOW.md`
- `openspec/specs/sdd-workflow` through this change's delta and later archive

No runtime trading package, broker client, Guardian, journal, or deployment service is
in the call/dependency blast radius.
