# Autoplan restore point

Captured 2026-09-05T13:38:10.750756+00:00
HEAD e65e394bf84b3c6e4559a219e816af96d341d75d

## proposal.md

## Why

The 2026-08-03 investigation recorded a mismatch between the unattended timer's legacy data-profile
soak record and the console's config-profile survey record. It reported six-hour renewal failures since
2026-07-30 and an attestation expiry of 2026-08-29. These are historical observations, not a verified
description of the current deployment. The 2026-09-05 continuation must re-establish installed service,
profile and expiry evidence before making an operational decision.

## What Changes

- Bind unattended soak attestation to the same explicit config profile used by the operator console, so
  the survey record, credentials, attestation output and engine interlock agree.
- In a human-approved operating window, collect the required three consecutive qualifying survey
  days and verify a fresh attestation. The original 2026-08-29 deadline has passed; it cannot be claimed
  as met by this continuation.
- Add a pre-expiry operational check that makes timer failure or insufficient survey evidence visible
  before an engine restart is blocked.
- Preserve every trading, Guardian, lane, kill-switch, adoption and automation-gate setting. This change
  does not place orders or relax the engine interlock.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `engine-safety`: unattended attestation renewal must consume and produce evidence within the same
  explicit operating profile, and renewal failure must be surfaced before the active attestation expires.

## Impact

- External operator service: `~/.config/systemd/user/tossos-attest.service` and its timer/logging path.
- Read-only survey lifecycle initiated from the operator console.
- Attestation evidence and expiry monitoring used by the engine startup interlock.
- Operations documentation and verification evidence; runtime trading behavior and operating toggles are
  out of scope.


## design.md

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
   systemd; do not alter engine or trading settings.
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

The reader limits input to 4096 bytes and at most 16 unique closed reason codes, rejects unknown fields,
unsupported versions, symlinks/non-regular files, and files not owned by the current user or accessible
to group/others. Attempt timestamps later than the read clock are invalid. Expiry may be omitted for
failed/refused attempts where it is unavailable; the UI then uses independently loaded current
attestation metadata or displays an unknown horizon. An `issued` record requires a valid expiry after
its attempt timestamp. Current attestation metadata is authoritative for expiry; disagreement with an
issued diagnostic record means unknown diagnostics and never suppresses its expiry warning.

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


## specs/engine-safety/spec.md

## ADDED Requirements

### Requirement: unattended attestation renewal uses one operating profile

The unattended renewal command SHALL use the same explicit config profile as the operator console's
read-only survey. The soak record, credentials, attestation output and engine-interlock input SHALL be
resolved through that profile's normal path rules and SHALL NOT mix an implicit data-profile record with
a config-profile output. Existing evidence SHALL NOT be copied between profiles to satisfy this
requirement.

#### Scenario: renewal consumes the console survey record

- **WHEN** the console-profile survey has produced qualifying evidence and unattended renewal runs
- **THEN** renewal reads that record and writes the attestation where the same profile's engine interlock reads it

#### Scenario: legacy data-profile evidence remains separate

- **WHEN** a stale legacy record exists under the data profile
- **THEN** renewal for the console profile does not read, move or relabel that record

### Requirement: renewal failure is visible before expiry blocks startup

A refused or failed unattended renewal SHALL preserve a failed status and its reasons. The normal
operator surface SHALL report renewal failure or impending attestation expiry before the active
attestation expires, without requiring an engine restart to reveal the problem. The warning SHALL NOT
stop a running engine and SHALL NOT weaken or bypass the startup interlock.

#### Scenario: incomplete evidence is reported before expiry

- **WHEN** renewal cannot issue a fresh attestation while the active attestation approaches expiry
- **THEN** the operator sees the failed status, unmet criteria and expiry horizon before startup is affected

#### Scenario: a running engine is not stopped by the warning

- **WHEN** the active attestation enters the warning window while an engine is already running
- **THEN** the system reports the condition without stopping that engine or changing the next-start interlock decision

#### Scenario: the console shows bounded advisory renewal diagnostics

- **WHEN** the operator opens the existing capability-attestation section
- **THEN** it shows the last renewal outcome, fixed reason messages, attempt age and expiry horizon without raw error text, credentials or account identifiers in renewal diagnostics
- **AND** missing, invalid or stale diagnostics are reported as unknown rather than successful
- **AND** warnings at or within 72 hours of expiry remain separate from startup-denial reasons and do not change attestation usability

### Requirement: attestation renewal preserves operational safety state

Preparing, installing and verifying unattended renewal SHALL NOT change trading, Guardian, lane,
kill-switch, position-adoption or automation-gate settings. Installing or activating an external service
definition SHALL require explicit human approval. Survey execution SHALL remain read-only and no live
order command SHALL be invoked as part of renewal verification.

#### Scenario: renewal is deployed without changing trading state

- **WHEN** the aligned renewal service is installed and verified with human approval
- **THEN** all operating toggles and risk settings remain byte-for-byte unchanged and no order-side effect occurs

### Requirement: deployment acceptance retains actual approved operational evidence

Acceptance and archival of this change SHALL require recorded explicit human approval for the reviewed
service template and target profile before external installation, reload or activation, and approval of
the production survey window before survey activation. The approval record SHALL identify the template
digest and authorized operations without exposing credentials or account identifiers.

Operational acceptance SHALL retain the actual dates and a non-sensitive profile reference for at least
three consecutive survey days that qualify under the existing qualification criteria, and verification
of a fresh attestation in that same profile. This requirement SHALL NOT alter issuer qualification
criteria, attestation validity or startup interlocks. Synthetic tests, copied or relabeled legacy
records, and backdated evidence SHALL NOT substitute for this operational proof.

#### Scenario: engineering is ready but operational evidence is incomplete

- **WHEN** isolated tests pass but explicit operational approval or three qualifying survey days are absent
- **THEN** engineering results are retained and the change remains unaccepted and unarchived

#### Scenario: approved operational verification is complete

- **WHEN** the approved same-profile survey has three consecutive qualifying days under existing criteria
- **THEN** acceptance records their actual dates and fresh same-profile attestation verification without a live engine restart or order mutation


## tasks.md

## 1. Evidence and proposal freeze

- [x] 1.0 Reserve `STORY-TOS-a063`, create the paired OpenSpec change, and capture
      `da80ce31b6a1ab5d443016768f970a82bab102db` as the pre-implementation base.
- [ ] 1.1 Capture the installed attestation unit/timer, both resolved record paths, current timer result
      and active attestation expiry using read-only commands; redact account and credential data.
- [ ] 1.2 Record the source-of-truth path flow and select the existing operator surface for the pre-expiry
      warning.
- [ ] 1.3 Complete independent security/operations review of this proposal and record the decision in
      `review.md` before implementation.

## 2. RED

- [ ] 2.1 Add a contract test that fails when the repository-backed renewal command omits the explicit
      console config profile or overrides only one of record/output.
- [ ] 2.2 Add a contract test that fails when the service masks a refused `soak attest` exit status.
- [ ] 2.3 Add a test for visible renewal failure/impending expiry that preserves running-engine and
      startup-interlock behavior.

## 3. GREEN

- [ ] 3.1 Add the minimal repository-backed service definition using the explicit console profile.
- [ ] 3.2 Preserve renewal exit status and expose bounded failure/expiry details on the selected existing
      operator surface.
- [ ] 3.3 Keep the survey and attestation path read-only and preserve every operating setting.

## 4. VERIFY and approved operations

- [ ] 4.1 Run focused tests, `make test`, `make vet`, `make validate`, `make sdd-sync`,
      and `make sdd-check`; retain actual results. The final change gate follows operational proof.
- [ ] 4.2 With explicit human approval, back up and install the service definition and reload user
      systemd without enabling or changing any trading control. Record approval scope, target profile,
      reviewed template digest and approved survey window before the corresponding operations.
- [ ] 4.3 Start the console-profile survey in the approved window and retain at least three consecutive
      qualifying days, recording actual dates. The original 2026-08-29 deadline was not verified and
      has passed; current evidence must not be backdated or replaced by synthetic test records.
- [ ] 4.4 Verify a fresh attestation is written to the engine profile, renewal failures are visible, and
      no live engine restart or order mutation is used for verification.
- [ ] 4.5 Complete independent diff/test review, PM synchronization and archive only after all evidence is
      recorded. Run `make gate CHANGE=a063-align-attestation-renewal-profile` after all prerequisite
      tasks are complete; archive is contingent on its success and Manager acceptance, never on task
      checkboxes alone. If that final gate fails, reopen this task and retain the failure evidence.


## base-commit.txt

da80ce31b6a1ab5d443016768f970a82bab102db
