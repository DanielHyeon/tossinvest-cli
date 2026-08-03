## Context

The console resolves its soak record from `--config-dir`, so the production console writes
`~/.config/tossctl/capability-soak.jsonl`. The existing user-systemd attestation service omits that
profile and therefore reads `~/.local/share/tossos/capability-soak.jsonl`, while its output still lands
under the config directory. The legacy record stopped on 2026-07-30, every six-hour renewal has been
refused, and the active attestation expires on 2026-08-29.

The survey and attestation command are local/read-only with respect to the account, but the resulting
file is a startup-interlock input. External service installation or activation remains a human-approved
operational action.

## Goals / Non-Goals

**Goals:**

- Make survey evidence, renewal input, renewal output and engine-interlock input resolve from one
  explicit production profile.
- Preserve renewal failure as observable state and warn before expiry blocks an engine restart.
- Collect and verify the required three consecutive survey days before 2026-08-29.
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
- **Deployment near expiry leaves too little evidence time** → complete profile alignment and begin the
  survey with more than three days of margin before 2026-08-29.

## Migration Plan

1. Capture the installed unit, timer state, current attestation expiry and both record paths read-only.
2. Add and test the repository-backed service definition without installing it.
3. With human approval, back up the installed unit, install the aligned definition and reload user
   systemd; do not alter engine or trading settings.
4. Start the console-profile survey in the approved window and keep it running for at least three
   qualifying days.
5. Verify timer failure/success visibility, issue a fresh attestation and confirm the engine-profile
   interlock reads that file without restarting the live engine.
6. Roll back by restoring the backed-up unit. Preserve both append-only records and the last valid
   attestation for diagnosis.

## Open Questions

- Select the existing operator surface that will carry the pre-expiry warning during implementation;
  prefer the console surface that already reports capability state over a new notification dependency.
