# a063 execution-baseline adoption preparation

Status: draft evidence only. This report does not bind reviews, validate a final H, complete task 4.0, or authorize an operational action.

## Provenance and source snapshot

- P: `da80ce31b6a1ab5d443016768f970a82bab102db`
- E: `e65e394bf84b3c6e4559a219e816af96d341d75d`
- accepted a120 tool commit: `71d1017223b536f797041c94e20b6c9febdb02b2`
- S: `c727ad12a42dcd15c494c1997e92816edaf17b6b`
- S tree: `2d608c945d4b2a4a2fd87824bf06a42815d256a1`

All fifteen reviewed implementation files matched the recorded SHA-256 manifest before and after the restricted copy. Their modes in S are ordinary Git `100644`; no NTFS executable mode was transplanted. The E→S Go diff contains exactly the 12 fixed allowlist paths, with no other Go path.

## Immutable inventory

`tools/logic-map/execution_baseline.py` ran once with full S and wrote the fixed non-overwriting drafts:

- `analysis/execution-baseline-ledger.json`: SHA-256 `3d9f0743cbbc9d30939582e19680e2c3ae58f9c968b54312d76f34aad4478675`
- `execution-baseline.json`: SHA-256 `171fa084c3d0041c89fedecbbe8a3ac6512a00f705812ed1300b0f98faa1fe3c`

P→E contains 315 commits and 329 net modified-existing function rows (inventory digest `16a6742aadc59467243ebcdf0153a5431ea459ce824eecec639e16dddaa25e38`). This remains historical debt. E→S contains 3 commits, 12 changed Go paths, and 9 modified-existing rows (inventory digest `a7102ca5f7406b1a37f282fe000a52f170c519e41aab0b25e1268188528f2b7b`). The nine fresh E-based bundles are five current and four base-revision obligations; three base-revision functions were deleted after E.

## Evidence and checks

- AST/risk evidence was regenerated from current S or immutable E blobs as required. All maps label themselves retrospective and do not assert pre-edit capture.
- Historical P bundles were moved byte-for-byte to `analysis/historical-planning-base/function-logic`; the recorded before/after hash inventories compare equal.
- `go test ./cmd/tossctl ./internal/soak ./internal/console` exited 0: tossctl 56.929s, soak 0.209s, console 118.004s. This is isolated test-suite evidence, not per-branch coverage or operational acceptance.
- Accepted-component recomputation passed: P→E ledger exact, E→S ledger exact, and all 9 E-based bundle hash/revision/branch-map bindings passed. The command/result log is `/tmp/tossos-a063-execution-adoption-checks.log`.
- `python tools/logic-map/check_analysis.py --root /tmp/tossos-a063-execution-adoption --change a063-align-attestation-renewal-profile` exited 1 as required at this draft stage: the record is uncommitted and therefore cannot be read from S. No validation bypass or success claim was made.

## Deliberate remaining blocker

The draft record retains empty adversarial and gstack review paths/digests exactly as generated. Therefore final adoption validation and `check_analysis.py` must remain blocked until the dedicated independent reviews are written, bound in a later metadata-only H commit, and verified in a clean detached H worktree. No final gate, archive, task checkbox, runtime/service action, or source modification was performed.
