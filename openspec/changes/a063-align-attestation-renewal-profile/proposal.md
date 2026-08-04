## Why

The unattended attestation timer reads the legacy data-profile soak record while the operator console
now writes the active survey record in its config profile. Renewal has therefore failed every six hours
since 2026-07-30, and the current production attestation expires on 2026-08-29; the first engine restart
after expiry would be refused without advance warning.

## What Changes

- Bind unattended soak attestation to the same explicit config profile used by the operator console, so
  the survey record, credentials, attestation output and engine interlock agree.
- Keep the read-only survey running long enough to collect the required three consecutive qualifying
  days, then verify a fresh attestation before 2026-08-29.
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
