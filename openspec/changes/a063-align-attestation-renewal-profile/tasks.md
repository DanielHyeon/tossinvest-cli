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
      `make sdd-check`, and `make gate CHANGE=a063-align-attestation-renewal-profile`.
- [ ] 4.2 With explicit human approval, back up and install the service definition and reload user
      systemd without enabling or changing any trading control.
- [ ] 4.3 Start the console-profile survey in the approved window and retain at least three consecutive
      qualifying days before 2026-08-29.
- [ ] 4.4 Verify a fresh attestation is written to the engine profile, renewal failures are visible, and
      no live engine restart or order mutation is used for verification.
- [ ] 4.5 Complete independent diff/test review, PM synchronization and archive only after all evidence is
      recorded.
