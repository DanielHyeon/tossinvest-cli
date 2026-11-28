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
