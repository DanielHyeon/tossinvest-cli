# Tasks 4.2–4.4 operational evidence template

**Status: prepared only. This template does not authorize or perform an operation.**

## Required explicit human approval before task 4.2

Record all values below before a mutable command is run:

| Field | Required value |
| --- | --- |
| Approval identity and timestamp | Human-provided, current; not inferred from this change. |
| Approved scope | Backup existing user-unit files, install only the reviewed binary and two user-unit templates, run `systemctl --user daemon-reload`, enable/start only `tossos-attest.timer`, and start the approved console-profile survey using the bounded procedure below. |
| Explicit exclusions | No engine restart, no automation-toggle change, no live order mutation, no unrelated unit or drop-in change, no survey reset. |
| Console and engine profile binding | Console target is `$HOME/.config/tossctl`. Before installation, record the normalized config-dir selected by the running engine in a redacted profile reference and prove it equals the normalized console target. If the engine has no explicit config-dir or the values differ, stop: this approval cannot install or survey until a separate same-profile plan is reviewed. Resolve and retain only redacted record/attestation sibling-path references after this equality check. |
| Candidate digest | `d9ef7f5ecac62a9e93159ebfc5cf5cf50d1ac1ee93b68c7ff9d48704777f6a6b` for `/tmp/a063-reviewed-20260906/tossctl`, re-verified immediately before install. |
| Service digest | `0bf50972660fa14ece6953b1d43aa51fe2df407a51783c7ea0c6decb3100de96` for `deploy/systemd/tossos-attest.service`. |
| Timer digest | `0db4e0cbfada2a59de4756747f17a3d7596d80dc7a0adafb3b3460849611edde` for `deploy/systemd/tossos-attest.timer`. |
| Approved survey window | A future start date/time and at least three consecutive qualifying calendar days. It cannot use the expired 2026-08-29 deadline or prior survey records. |

Run only the bounded preflight, installation, and rollback procedures already reviewed in
[`../deployment-plan.md`](../deployment-plan.md). Copy only redacted output into the completion record.

### Bounded task 4.3 survey-start procedure

The approval must additionally name the existing console process or console launch procedure that uses
`$HOME/.config/tossctl`; it must not name a different profile. Before starting, use the console's existing
survey control to determine whether a survey for that same resolved record is already running. If it is
running, retain that fact and do not interrupt or duplicate it. If it is absent, the approved human may
use the existing console's survey-start control exactly once; that path delegates to the profile-carrying,
detached `soak run` procedure. Do not invoke `engine run`, change console/engine launch arguments,
reset a record, pass `--record`, or start a second survey through a shell command.

The actual completion record must include the human-operated start/leave-running decision, the actual
local start date/time, the redacted same-profile record reference, and the approved three-day window.
This is a read-only survey of official account endpoints; it does not authorize any order command.

## Task 4.2 completion record

- Approval details and survey window: **pending**
- Preflight result: **pending**
- Backup directory and validation: **pending**
- Template/binary re-verification immediately before install: **pending**
- `systemd-analyze --user verify` and `systemctl --user daemon-reload`: **pending**
- Timer enabled/active result: **pending**

## Task 4.3 completion record

For each of three consecutive qualifying days record the actual local date, command/result category,
and redacted status only. Do not retain account, credentials, raw record, raw attestation, or full
service output. A missed or nonqualifying day restarts the three-day observation window.

| Actual date | Qualifying result | Redacted evidence reference |
| --- | --- | --- |
| pending | pending | pending |
| pending | pending | pending |
| pending | pending | pending |

## Task 4.4 completion record

After the three-day window, use read-only checks to establish all of the following:

1. the approved console profile and running engine profile are the same normalized config-dir;
2. the engine-profile attestation modification time is newer than the approved survey-window start;
3. the engine-profile renewal-status file exists at the same profile's resolved sibling path and contains a bounded
   outcome consistent with the last renewal;
4. a deliberately non-mutating renewal failure fixture or approved safe failure observation is visible
   on the existing operator surface without persisting raw error/account data; and
5. evidence records no engine restart and no order mutation.

Actual values: **pending**.

## Task 4.5 handoff

Once all pending fields are backed by actual dates and redacted evidence, request a fresh independent
adversarial review, then a subsequent gstack review, PM synchronization, a final `make gate
CHANGE=a063-align-attestation-renewal-profile`, and Manager acceptance. Archive remains prohibited
until that final gate succeeds.

## Readiness reviews

- Adversarial re-review: **CLEAR for prepared readiness only**, retained as
  [`task-4.2-4.4-adversarial-rereview.md`](task-4.2-4.4-adversarial-rereview.md), SHA-256
  `ac3d26295e8c5bc18e3bf0947068d259eedf5c525bc3a245873700baf984f7a2`.
- Subsequent gstack review: **CLEAR for prepared readiness only**, retained as
  [`task-4.2-4.4-gstack-review.md`](task-4.2-4.4-gstack-review.md), SHA-256
  `fe45fb952c237da20822e2e06e3b2a87d320b4d00d47d7bdd8ddb0b1d5c836de`.

These reviews approve the packet's fail-closed preparation, not a mutable operation. Tasks 4.2–4.4
remain pending until a current human approval and real, time-bound evidence fill the records above.
