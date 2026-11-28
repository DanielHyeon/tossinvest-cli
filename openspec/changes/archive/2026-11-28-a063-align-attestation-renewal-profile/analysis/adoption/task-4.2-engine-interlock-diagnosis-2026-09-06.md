# A063 Task 42 — engine-run diagnosis

## Verdict

The one supplied explicit `tossctl engine run` did not remain running because
startup correctly failed closed at the **automation-gate capability-attestation
interlock**.  Its gate was enabled, but the attestation did not cover all
required order-lifecycle operations.  Consequently, the engine deliberately
started no loops and exited before normal steady-state operation.

Sanitized stable categories/identifiers observed:

- `engine.operating_mode`
- `automation-gate preconditions unmet`
- `attest: capability-attestation coverage incomplete`
- `startup interlock unmet; no loops started`

The supplied log records an `Error` outcome, but does **not** record a process
exit status.  The defensible status is therefore **numeric exit code unavailable
in the supplied evidence**; no numeric code is inferred from the error text.

## Classification

This is an expected safety/interlock refusal, not a recoverable local
configuration defect:

- The configuration was readable; the log does not indicate malformed or
  missing required configuration.
- The log does not indicate authentication, network, database-corruption, or
  journal-write contention failure.
- The engine lock and configuration-lock files are empty advisory/locking
  artifacts.  Their presence alone is not evidence of a blocking lock and must
  not be removed as a remedy.
- The engine's own help documents this exact fail-closed startup behavior: with
  the automation gate on, unmet interlock clauses refuse startup rather than
  starting any loop.

## Narrowest next action

The missing attestation coverage must be obtained for the required order and
cancel lifecycle operations, then the engine can be started again.

1. **Human authorization required:** inspect the planned live verification only:
   `tossctl verify run --list`
2. **Human authorization required:** a human may run `tossctl verify run` and
   provide its expiring terminal approval.  The local help explicitly states
   that this procedure places real, limit-only, single-share orders and cancels
   them; it is not safe for an agent to run automatically.
3. After the verification reports the required coverage, a human may start
   `tossctl engine run` again.  This can enable an automated trading engine, so
   it also requires the operating authorization applicable to live execution.

No config edit, lock-file removal, database manipulation, or engine retry is
the narrow remedy before step 2 succeeds.

## Evidence commands and exit status

| Command / evidence source | Purpose | Exit status |
| --- | --- | --- |
| supplied `tossctl engine run` log | classify the explicit run | numeric status absent from log |
| `tossctl engine --help` | local command metadata | 0 |
| `tossctl engine run --help` | local startup/interlock semantics | 0 |
| `tossctl verify --help` | determine the permitted remediation and authorization boundary | 0 |
| `stat` / filename-and-type inspection of allowed local metadata | check lock and journal artifact metadata only | 0 |

## Mutation confirmation

No mutation was performed.  I did not execute the engine, verification/survey,
orders, systemctl actions, configuration edits, source edits, repository writes,
or lock/database changes.  This report is the only requested diagnostic output.

## Independent review and Manager verification

- Adversarial review: **CLEAR**, retained as
  [`task-4.2-engine-adversarial-review.md`](task-4.2-engine-adversarial-review.md), SHA-256
  `c40549f3d9be2a2ae83f867d171daa3b7e46817dd7b7d7fe27607bafbd6a7d10`.
- Subsequent gstack review: **CLEAR**, retained as
  [`task-4.2-engine-gstack-review.md`](task-4.2-engine-gstack-review.md), SHA-256
  `7b0836a37f8d6db82ade2e452338891ad2273d886e50f25e554d31b2ce44a81a`.
- Manager independently confirmed both reviews distinguish the user-authorized engine launch from the
  unentered task-4.2 user-unit activation block, and preserve the separate human approval required by
  any live-order verification.

This diagnosis does not close task 4.2. A successful engine startup still needs complete capability
attestation; task 4.3, 4.4, 4.5 and archive remain pending.
