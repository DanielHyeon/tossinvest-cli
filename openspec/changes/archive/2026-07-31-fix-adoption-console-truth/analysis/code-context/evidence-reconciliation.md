# Evidence reconciliation

## Agreements

- Current HEAD and CodeGraph agree that console path selection differs from
  the engine only at `runConsole`'s unconditional `journal.DefaultPath`.
- Current HEAD and browser evidence agree that the inline `oninput` is required
  to update the range label while the deployed CSP prohibits that handler.
- Current HEAD and SQLite evidence agree that v8 has the tables checked by
  `OpenReadOnly` but lacks the v9 `policy_id` column selected later.

## Advisory differences

- CodeGraphContext returned no extra candidates. No conflict exists to resolve.

## Reconciled edit boundary

- Add a console path resolver that mirrors the engine's explicit-config/default
  rule; retain `OpenReadOnly`.
- Extend read-only compatibility checking only for the column the read API
  requires; do not migrate, fallback, or weaken the query.
- Convert a human percentage to the existing fraction at the HTTP boundary;
  do not change config or exit-policy math.
- No engine, soak, broker, toggle, or order path is in scope.

