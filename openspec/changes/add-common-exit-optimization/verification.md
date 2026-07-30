## Verification evidence

Date: 2026-07-31

### Behavior and regression checks

- Focused Go tests passed for `internal/exitpolicy`, `internal/config`,
  `internal/journal`, `internal/app/engine`, `internal/console`, and
  `cmd/tossctl`.
- Race-enabled tests passed for the affected packages.
- Full `make test`, `make vet`, and `make validate` passed.
- `openspec validate add-common-exit-optimization --strict` passed.
- Function Logic Map AST/risk evidence was refreshed after implementation and
  every recorded RED/GREEN branch result is complete.
- Independent adversarial review found and resolved the fail-open runner parse,
  stale policy comments, and migration-v8 test-boundary issues.
- `make sdd-sync` and `make sdd-check` passed immediately before the completion
  bookkeeping. The first gate invocation reached the task-completeness check and
  refused only because tasks 5.3 and 5.4 had not yet been checked; the final gate
  is run after this evidence and episodic memory are retained.

### Safety evidence

- No LIVE gate was changed.
- The engine was not started.
- No order, cancellation, amendment, adoption, or other account mutation was
  submitted.
- Optimization POST wiring can only persist the selected registered policy ID
  and its audit record; it cannot change trading or automation capabilities.
