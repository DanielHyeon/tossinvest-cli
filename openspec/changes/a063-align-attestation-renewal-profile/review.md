# Review — a063-align-attestation-renewal-profile

- Date: 2026-08-03
- Stage: proposal freeze; implementation not started
- Voices: maintainability/release review + adversarial operations/security perspective
- Evidence: a060 `issues.md` I7, installed user-systemd unit/timer read-only inspection, current path
  resolution in `cmd/tossctl/soak.go`

## Findings and disposition

- **Accepted**: bind the timer to one explicit config profile instead of copying or relabeling legacy
  evidence. One profile authority removes the input/output asymmetry without weakening attestation.
- **Accepted**: make the service definition repository-backed and test its argv/failure behavior. The
  retired untracked soak autostart artifact demonstrated the maintenance cost of external-only logic.
- **Accepted with gate**: preserve non-zero renewal status and surface impending expiry on an existing
  operator surface. The warning is advisory and cannot stop a running engine or relax startup checks.
- **Required**: human approval before installing/reloading the external service or starting the
  multi-day production survey. No operating toggle or live-order command is authorized by this plan.
- **Required deadline**: collect at least three qualifying survey days and verify a fresh attestation
  before the current one expires on 2026-08-29.

## Scope decision

The plan is ready for implementation after the independent Manager verifies the artifacts. No runtime
code, external service, survey process or operating setting was changed while creating this change.

Function Logic Map: not-applicable — this commit creates planning artifacts and amends comments/contracts
only; implementation tasks will create any required analysis before editing existing functions.
