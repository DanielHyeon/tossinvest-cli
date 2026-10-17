# a147 independent gstack review

Verdict after independent re-review: NO UNRESOLVED FINDINGS in reviewed a147 scope. The reproduced P2 denominator-contract defect is fixed and independently retested. This is not aggregate-gate approval or archive authorization.

Independent reviewer: Codex child context, separate from implementer. Scope only tools/a147-jev-mfi-offline and openspec/changes/a147-validate-jev-mfi-trend; imported a145 timestamp helper inspected. No repository changes, API/model calls, collection, runtime actions, commit/push/archive performed by this reviewer.

## Historical finding — resolved

[P2] Move invalid state-digest identities into selection-unknown. `analysis.py`, branch immediately following `digest = first.get('state_digest', '').removeprefix('sha256:')` (lines 385–389 at reproduction). A paired J1/J2 row with `state_digest="invalid-digest"` and `screen.finalist=false` returns a `nonselected` cohort candidate with `mfi_excluded.invalid_identity=1`. This violates the spec requirement that ambiguous identity SHALL NOT be classified as nonselected, and distorts cohort missingness. Fix this branch to set selection unknown and identity exclusion for labels, matching the missing-state and state-mismatch handling. Add evaluator regression with malformed digest and a valid pair. This was FIXABLE and blocked declaring this explicit contract complete until corrected; no trading side effect was involved.

Synthetic repro run from temporary snapshot containing only paired judgment rows; observed result: selection=nonselected, candidates=1, mfi_valid=0, label_valid=0, mfi_excluded={invalid_identity:1}, label_excluded={missing_label:1}. No state/capture files are required because this early branch exits first.

## Coverage and executed evidence

- Read gstack review SKILL.md, required checklist.md, security/testing/red-team specialist checklists; applied substantive checklist and adversarial reasoning locally. No Claude pass, independent second model pass, or automatic gstack review-log pass claimed. Spawned host setup/telemetry/routing/autofix omitted per delegated read-only scope. No additional agents spawned.
- Read all 626 original analysis source lines, focused tests, README, proposal/design/spec/tasks. Source was mutable concurrently; hashes below bind report snapshot, and final changes require recheck.
- Ran `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tools/a147-jev-mfi-offline -v`: 19 tests PASS, 2.282 seconds. Includes subprocess CLI artifact assertions on synthetic inputs and network-blocked happy path.
- MFI: checked 18-bar endpoints, 14 signed flows per value and lag-three comparison, typical-price sign, equal-price neutrality, zero-flow handling, numeric validation, age/future/gap rejection, epsilon. Formula 100*P/(P+N) is equivalent to specified formula for finite positive sums. Existing tests cover monotonic/negative/zero flows and temporal boundaries.
- Temporal/provenance: checked receipt <= observation, bars <= receipt, session windows, source name/timezone, state byte digest and capture byte digest. These validate supplied evidence consistency, not external source/clock authenticity, correctly disclosed in README/output. No live provenance verified.
- Cohorts: checked pair deduplication, duplicate cycle/symbol collapse into unknown, missing-state/mismatched-state unknown handling, per-session/selection reporting, matched cycle equal weighting, paired-label timing/value checks; found malformed-digest exception above.
- Preservation/security: checked exclusive output creation, component-wise O_NOFOLLOW directory traversal, sidecar flock, nlink refusal, protected-input inode checks, SHA-256 filename allowlist, no shell/eval/network code in implementation. Tests cover existing-file refusal and symlink/hardlink append refusal. Parent-directory replacement race, malicious filesystem mount changes, crashes/disk-full atomicity and hostile special input files were not stress-tested; no unsupported security guarantee made.
- Session accounting: planned sessions remain in output, supplied complete status alone does not confer eligibility, failed/partial need reasons, output always inconclusive. Calendar truth and candidate-pool completeness remain externally supplied limitations. No performance inference or five-session collection claim.
- SQL/auth/API/frontend/migrations/distribution specialist concerns: not applicable to this local offline Python tool. Go AST/CodeGraph internal-function map not applicable to this Python-only read-only review, no Go behavior edited. Aggregate gate, PM lifecycle synchronization and archive legality are parent/worker responsibilities, not independently certified here.

Recommendation after re-review: proceed to the parent-managed aggregate gate and lifecycle checks; no remaining independent-review blocker was found. Existing test method now includes both malformed-digest cohort regressions.

## Fixed snapshot SHA-256 (independently rechecked)
- `tools/a147-jev-mfi-offline/analysis.py`: `c8d802e3f6423e100353310412923c845e50464bb637e391c81c33eaeb395c75`
- `tools/a147-jev-mfi-offline/test_analysis.py`: `ee3b03145f2fdd21e0fda36c1f27ab3bc9e417289a67f5529d6ed623eb870bbc`
- `tools/a147-jev-mfi-offline/README.md`: `27fa3952f1a5f107b1acfcc72e4adb52780b660aa3f7c3c4730fce7216bf526a`
- `openspec/changes/a147-validate-jev-mfi-trend/specs/semantic-judgment-shadow/spec.md`: `373adca1d8ea7fdf00d5d65c438fbf25bf7e8bec66c1a9bd6d55a676a8481574`

Report time UTC: 2026-10-17T18:29:05.126753+00:00

## Independent fix verification

Important snapshot correction: the initial finding and synthetic reproduction above describe the PRE-FIX implementation, where the branch only assigned mfi_reason. No pre-fix source hash was captured at reproduction time. The original report was written while the worker was changing the files, so its hash appendix already contained the FIXED version; those hashes must not be treated as hashes of the vulnerable version. This section explicitly supersedes the initial verdict and separates those observations.

Current analysis.py lines 385–389 now use `record.update(selection='unknown', mfi_reason='invalid_identity', label_reason='invalid_identity')`. This resolves the demonstrated denominator leakage and preserves candidate count while attributing both exclusions to identity.

Current test_analysis.py lines 170–180 run the actual CLI for screen.finalist=True and False, each with an invalid state digest and otherwise valid paired fixture. Both assert unknown selection plus identity exclusions for MFI and labels; original ledger bytes are then restored for later fixture scenarios. Independently read this implementation and test change, then reran the complete focused suite: 19 test methods PASS in 0.725 seconds, including these two subtests. No live inputs or calls.

Final source hashes were independently rechecked against the worker-provided values before rerun and after rerun:
- `analysis.py`: `c8d802e3f6423e100353310412923c845e50464bb637e391c81c33eaeb395c75`
- `test_analysis.py`: `ee3b03145f2fdd21e0fda36c1f27ab3bc9e417289a67f5529d6ed623eb870bbc`

Fix-review time UTC: 2026-10-17T18:29:52.105293+00:00
