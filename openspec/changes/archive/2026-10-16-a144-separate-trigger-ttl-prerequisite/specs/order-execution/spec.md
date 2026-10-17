## ADDED Requirements

### Requirement: Trigger-only verification admits exactly one non-prerequisite measurement

The system SHALL admit a user-supervised `conditional-trigger` redo when every procedure step other than `conditional-trigger` and `idempotency-ttl-edge` has a latest PASS verdict. `idempotency-ttl-edge` SHALL NOT be required to PASS for this admission, regardless of whether its latest verdict is PASS, skipped, failed, refused, missing, or nonterminal.

The system SHALL preserve every existing TTL entry as recorded and SHALL NOT rewrite it, synthesize an aggregate PASS, rerun TTL, or infer TTL capability from trigger admission. If ordinary no-opt-in Runner flow records a new skipped line for a missing or nonterminal TTL entry, that line SHALL not be treated as a PASS or as TTL execution. Every other procedure step, including read-only steps, SHALL still require latest PASS.

#### Scenario: TTL was skipped by default

- **WHEN** every non-TTL, non-trigger procedure step has latest PASS and TTL's latest verdict is skipped
- **THEN** trigger-only admission does not reject because of TTL and does not mutate or rewrite the TTL record

#### Scenario: TTL was failed, refused, missing, or nonterminal

- **WHEN** every non-TTL, non-trigger procedure step has latest PASS and TTL is failed, refused, absent, or nonterminal
- **THEN** trigger-only admission treats only TTL as the exception; it may perform normal preflight and ordinary skipped/excluded metadata recording, but does not execute the TTL step body, redo the TTL measurement, or send a TTL broker mutation

#### Scenario: another prerequisite is not PASS

- **WHEN** no `pending-create` checkpoint is unresolved and any procedure step other than TTL and conditional-trigger lacks latest PASS
- **THEN** trigger-only admission rejects before broker construction, confirmation, cleanup, planned mutation, or broker request

### Requirement: Trigger-only mode remains a single-step irreversible measurement

The system SHALL keep trigger-only mode restricted to confirm-each, resume, one redo equal to `conditional-trigger`, a usable owner-only receipt, no TTL-edge opt-in, no prior outstanding artifact, and no unresolved M0 checkpoint for new trigger execution; the existing `pending-create` read-only recovery exception remains unchanged. It SHALL retain the existing execution exclusion and raw-read-source requirements.

The Runner SHALL not execute the TTL step body, send a TTL broker mutation, or add a TTL planned mutation in this mode. Normal preflight and `Plan.Excluded` metadata are permitted. Existing terminal TTL entries SHALL remain settled; missing or nonterminal TTL entries may append the ordinary no-opt-in skipped line but SHALL not cause TTL execution. Any outstanding artifact, including one associated with an earlier TTL attempt, SHALL remain a separate pre-broker HOLD.

#### Scenario: TTL is passed together with trigger mode

- **WHEN** trigger mode includes the TTL-edge opt-in or any redo besides exactly conditional-trigger
- **THEN** the system rejects before broker construction and no mutation occurs

#### Scenario: outstanding artifact or non-pending checkpoint remains

- **WHEN** a prior verification artifact is outstanding or an M0 checkpoint other than `pending-create` is unresolved
- **THEN** the system holds before cleanup, broker construction, confirmation, or mutation regardless of TTL's verdict

#### Scenario: pending-create checkpoint remains

- **WHEN** the only unresolved M0 checkpoint is `pending-create`
- **THEN** the system retains existing read-only broker recovery and ends HOLD with zero confirmation, cleanup, or mutation regardless of TTL's verdict
