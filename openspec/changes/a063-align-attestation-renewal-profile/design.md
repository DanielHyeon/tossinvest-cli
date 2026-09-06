## Context

The historical 2026-08-03 investigation reported the following path mismatch, which must be
revalidated against current code evidence and the installed deployment before implementation:
the console resolves its soak record from `--config-dir`, so the production console writes
`~/.config/tossctl/capability-soak.jsonl`. The existing user-systemd attestation service omits that
profile and therefore reads `~/.local/share/tossos/capability-soak.jsonl`, while its output still lands
under the config directory. The legacy record stopped on 2026-07-30, every six-hour renewal has been
refused, and the recorded attestation expired on 2026-08-29. Current operational state is unverified.

The survey and attestation command are local/read-only with respect to the account, but the resulting
file is a startup-interlock input. External service installation or activation remains a human-approved
operational action.

## Goals / Non-Goals

**Goals:**

- Make survey evidence, renewal input, renewal output and engine-interlock input resolve from one
  explicit production profile.
- Preserve renewal failure as observable state and warn before expiry blocks an engine restart.
- Collect and verify the required three consecutive survey days in an explicitly approved operating
  window. Record the actual dates and fresh attestation expiry; do not backdate evidence.
- Keep the service definition reproducible and drift-checkable from the repository.

**Non-Goals:**

- Changing soak criteria, attestation validity, interlock requirements or engine startup semantics.
- Placing, amending or cancelling an order.
- Flipping any trading, Guardian, lane, kill-switch, adoption or automation-gate setting.
- Migrating the 405-cycle legacy data-profile record into the console profile.

## Decisions

### D1. One explicit config profile is the path authority

The source-controlled service command will invoke
`tossctl --config-dir %h/.config/tossctl soak attest ...`. It will not combine an explicit record with an
implicit output path: both paths must continue to resolve through the command's normal profile logic.
The survey must run with that same profile before renewal is attempted.

### D2. The service definition becomes repository-backed

The executable command and its failure behavior will have a source-controlled template plus a drift
test or deterministic validation. Copying an untracked shell fragment directly into the user systemd
directory repeats the maintenance failure recorded in a060 I3.

### D3. A refused renewal remains failed and visible

Logging the refusal is not enough if the service then exits successfully. The service must retain the
command's non-zero result, and the normal operator surface must show renewal failure or impending expiry
without requiring an engine restart. The warning is advisory: it must not stop a running engine or relax
the startup interlock.

### D4. Installation and live verification stay human-gated

Implementation may prepare and test the service definition without activating it. A human must approve
installation/reload and the timing of the continuous survey. Verification reads service state, records
and attestation metadata; it does not restart the live engine or change operating toggles.

## Risks / Trade-offs

- **Survey and engine share broker rate budget** → start the survey in the approved low-contention
  window, preserve existing retry/backoff, and observe throttling without weakening criteria.
- **A wrong explicit profile could attest another environment** → derive all paths from one reviewed
  config root and verify account/profile identity before installation.
- **Timer failures become noisy while evidence is still accumulating** → report the unmet criteria and
  expiry horizon explicitly; do not mask the failure exit code.
- **Deployment near or after expiry leaves insufficient evidence time** → inspect current expiry and
  agree an operating window with at least three qualifying survey days. If already expired, preserve
  the startup interlock and report the condition; do not restart the live engine for verification.

## Migration Plan

1. Capture the installed unit, timer state, current attestation expiry and both record paths read-only.
2. Add and test the repository-backed service definition without installing it.
3. With human approval, back up the installed unit, install the aligned definition and reload user
   systemd; do not alter engine or trading settings. Before installing the unit, verify the selected
   binary supports `soak attest --record-renewal-status`; an older binary cannot execute this template.
   Any binary replacement or console deployment needed to expose the new warning must be included in
   the concrete operational approval scope, without restarting the live engine.
4. Start the console-profile survey in the approved window and keep it running for at least three
   qualifying days.
5. Verify timer failure/success visibility, issue a fresh attestation and confirm the engine-profile
   interlock reads that file without restarting the live engine.
6. Roll back by restoring the backed-up unit. Preserve both append-only records and the last valid
   attestation for diagnosis.

## Open Questions

The existing console capability-attestation section (`internal/console/templates.go`) is the warning
surface. No additional notification dependency or engine-control action is needed.

## 2026-09-05 implementation contract

Current read-only inspection by the implementation teammate finds the installed unit already uses the
explicit config profile but masks refusal with a shell success fallback. The observed expiry is
2026-10-05T13:13:01Z, so this continuation does not claim a current expiry incident. Sanitized inspection
and generated function evidence must be retained under `analysis/` before implementation freeze.

### D5. Renewal diagnostics are opt-in and advisory

The repository unit opts into recording the last renewal attempt with `--record-renewal-status`. With the
option off, existing command behavior is preserved. The status path derives from the same profile and
is shared by the console reader; the unit cannot override only the input or output path. The final path
binding appends `.renewal-status.json` to the full attestation path resolved by
`resolveSoakAttestationPath(root, "")`; the console derives it from that same resolved attestation path.
Using the full basename prevents collisions between different attestation files in one directory.
Recording mode rejects explicit `--record` and `--out` overrides before any issuance. Default and custom
configured attestation locations, including different basenames in the same directory, must be covered
by contract tests.

Use a versioned, bounded status object: UTC attempt timestamp, outcome (`issued`, `refused`, `failed`),
closed reason codes and known expiry timestamp. Do not persist account identifiers, credential values,
raw error strings, endpoint response contents or arbitrary paths. Map typed evaluation/stage results to
codes; do not parse free-form errors or duplicate qualification policy. Atomic replacement uses an
owner-only file. Status-write failure remains a nonzero command result, including when issuance itself
succeeded, and cannot modify qualification criteria or invalidate the last good attestation.

If existing evaluation exposes only textual reasons, use one typed internal issue evaluation to
produce both the existing public message result and closed diagnostic codes. Preserve every existing
predicate, message and return decision. Carry refusal codes with a typed error/result from that same
evaluation so an attestation attempt does not scan/evaluate the survey twice for diagnostics. Generate
pre-edit evidence and compatibility tests for any evaluator/issuer function touched by this refactoring.

The reader limits input to 4096 bytes and at most 16 unique closed reason codes, rejects unknown fields,
unsupported versions, symlinks/non-regular files, and files not owned by the current user or accessible
to group/others. Attempt timestamps later than the read clock are invalid. Expiry may be omitted for
failed/refused attempts where it is unavailable; the UI then uses independently loaded current
attestation metadata or displays an unknown horizon. An `issued` record requires a valid expiry after
its attempt timestamp. Current attestation metadata is authoritative for expiry; disagreement with an
issued diagnostic record means unknown diagnostics and never suppresses its expiry warning.

The user-systemd deployment is Linux-scoped. Platform-specific secure descriptor handling is isolated
from portable CLI code so supported release targets still build. Linux/macOS readers must reject
non-regular files without blocking on FIFO open. Where secure ownership/no-follow verification is not
implemented, diagnostics remain unknown rather than silently weakening validation; existing default-OFF
issuer and engine behavior remains available. Cross-compilation is build proof, not native runtime proof.

The reader validates version, size, timestamps and closed enum values. Missing/invalid/stale diagnostics
mean unknown diagnostic state, never renewal success. Display fixed, escaped operator messages and the
last-attempt age. Use a 72-hour pre-expiry advisory window (including its boundary), and treat a recorded
attempt older than 12 hours as stale for the existing six-hour cadence. These advisory thresholds do
not affect attestation validity, `Usable`, engine startup decisions or any running engine.

Show advisory warnings separately from the existing startup-denial reasons. Include failure/unknown
state even when the attestation file is absent. A successful later attempt replaces a prior failure;
an old success must not conceal expiry or stale diagnostics. Tests use isolated profiles and synthetic
fixtures only; those fixtures are engineering proof, never operational acceptance evidence.

Malformed attestation load errors on this surface must also use a fixed message rather than embedding
raw parser errors or paths. RED/GREEN proof must cover this existing error branch as well as the new
diagnostic reader. A status-write error must leave the attestation itself intact, with the command still
reporting nonzero failure.

## Explicit retrospective execution-baseline adoption

The completed a120 contract in canonical `sdd-workflow` governs the one-time
comparison exception for this change. P remains
`da80ce31b6a1ab5d443016768f970a82bab102db`; only fixed E
`e65e394bf84b3c6e4559a219e816af96d341d75d` may become the effective comparator
after all immutable-ledger, source-snapshot, independent-review and ordinary
function-map requirements pass. The detailed sequence is
`analysis/execution-adoption-plan.md`. This is retrospective exception evidence,
not a claim of original pre-edit compliance or forgiveness of historical debt.

The preparation uses a detached Linux worktree and the reviewed external SDD
interpreter, preserving all fifteen reviewed implementation digests and normal
Git file modes. It creates no local virtualenv and performs no deployment or
survey operation. The record binds actual dedicated adversarial and subsequent
gstack reviews only after they occur. All operational tasks remain separate.
