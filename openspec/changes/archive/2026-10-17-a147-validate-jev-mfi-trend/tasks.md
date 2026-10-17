## 1. Implementation readiness

- [x] 1.1 Confirm implementation authorization, Story/contract READY, immutable baseline and applicable evidence; current design approval alone is not implementation authorization.
- [x] 1.2 Record evidence and boundary for any changed existing functions before editing; preserve r1, historical lots, trading and operational settings.

## 2. Evidence and analysis implementation

- [x] 2.1 Implement minimal append-only prospective manifest and offset-bearing closed-bar sidecar with immutable source/identity linkage; do not alter r1 input or prior evidence.
- [x] 2.2 Implement offline MFI14/three-bar-delta, as-of checks, identity joins, missingness accounting and market/session/selection reports as specified.
- [x] 2.3 Test MFI calculations, 18-bar boundary, equal and zero flows, boundary MFI values, delta epsilon, missing and exceeded age/delay limits, timestamps/offsets, future and duplicate bars, late-retrieved bars without as-of proof, gaps, invalid prices and volumes, and incomplete or invalid labels.
- [x] 2.4 Test J1/J2 deduplication, selected/nonselected/unknown denominators and same-cycle averaging, manifest immutability, interrupted/partial sessions and no silent replacement; verify no r1 or historical evidence mutation.

## 3. Verification and lifecycle preparation

- [x] 3.1 Independently review the implementation and tests; document results and remaining risks.
- [x] 3.2 Execute applicable repository checks and record individual results; the aggregate gate result is recorded separately after it actually completes.
- [x] 3.3 Verify formal spec/PM synchronization and archive plan; execution of archive, commit and push occurs after the aggregate gate and applicable approvals.

## Post-gate and separately authorized observation

Focused synthetic tests and scoped OpenSpec validation passed. The isolated aggregate gate passed all 11 stages (exit 0; `/tmp/a147-gate.log` and `/tmp/a147-gate.exit`); archive and commit/push have not run. New session collection, model/API calls and scheduling are not authorized by this design. Future separate approval and a frozen preregistration are required; implementation completion alone does not mean the planned five sessions per market were measured, or that the MFI association is validated.

## Base reanchor approval — 2026-10-18

Manager approved reanchoring `936d7e0d23510277224f5b16655169b1c80a07be` to `055f47f78f9478d4a7522a70bb461789bd7b937e` after the zero-own-Go-commit and 14 sibling-function attribution audit recorded in `review.md`. A base-only commit and current-HEAD gate are required before the final lifecycle steps.
