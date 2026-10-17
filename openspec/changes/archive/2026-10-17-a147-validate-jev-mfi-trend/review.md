# Design review — 2026-10-18 (historical, superseded for implementation authorization)

- Scope: design and review only. Proposal, design, delta and tasks describe a prospective observational MFI evaluation with unchanged r1 and separate authorization for future implementation and collection.
- Independent design review: PASS, including whole-study preregistration, as-of provenance, unknown-selection denominators, MFI boundary cases and mandatory age/delay limits. Record: `outputs/mfi-jev-20261018/a147-design-review.md` in the ChatGPT task workspace.
- gstack design review: PASS on the same final contract, including offset, closed-bar and input-preservation boundaries. This is a design verdict only.
- OpenSpec strict validation and PM generator check: PASS for the design artifacts; implementation tasks remain unchecked.
- Implementation is NOT READY: user has authorized neither implementation nor additional collection, and the repository's implementation READY checks have not been completed. No sessions were measured under this plan; no performance improvement is established. No code, tests, model/API calls, scheduling, operation, commit, push or archive were performed for this change.

## Implementation verification — 2026-10-18

The user subsequently authorized the offline implementation. Additional observations, API/model calls, scheduling, runtime changes, commit, push and archive remain outside that authorization. No session has been collected under this plan.

The implementation and 19 focused synthetic tests run in an isolated checkout based on `936d7e0d`. Raw test log: `/tmp/a147-primitives.log`; exit 0. Offline synthetic CLI evidence: `/tmp/a147-cli-evidence/report-frozen.json`, `/tmp/a147-cli-record-frozen.log`, `/tmp/a147-cli-evaluate-frozen.log`; exits 0. A duplicate event is rejected (`/tmp/a147-cli-duplicate.log`, exit 1). Input ledger SHA remains `eac5d9cd35fbb242ce2110bd444e8f7ef217864027256a2177d3be1280033a05`. Source SHA-256 `2cf098c6a44f84135892dbfc74fd54ee11e42bffe3084e36cd09c45139a600ee`; test SHA-256 `f4bb3dc796cc174f2543c80dc306550f91467bdc2ff870481245ca3317329096`. No existing function was modified; Go AST tooling does not apply to the new Python leaf.

Candidate pool completeness, capture origin and receipt clocks, and independent evidence that a manifest was fixed before observation remain unverified from supplied files alone. Session completeness is review eligibility, not verified market coverage. There is no empirical MFI performance conclusion.

The aggregate gate passed all 11 stages with exit 0 in the isolated checkout based on `936d7e0d`; log: `/tmp/a147-gate.log`, exit receipt: `/tmp/a147-gate.exit`. This included test, test-seams, race, vet and validation. `GOFLAGS=-buildvcs=false` and `SDD_PYTHON` were scoped to the gate process. The preceding official sync succeeded for hard CodeGraph and CodeGraphContext; GBrain was deliberately skipped during sync and its stale advisory warning remained. This is not a claim that the original dirty worktree passed the full gate.

Independent and gstack implementation reviews: PASS on source `2cf098c6`, test `f4bb3dc7`, README `27fa3952`; the reviewers separately checked the final added test cases. Scoped OpenSpec validation and PM tracker check passed before this task update. Archive preparation: the delta adds requirements to the existing `semantic-judgment-shadow` capability; a future authorized formal archive must synchronize the delta and update Story/PM paths, without modifying the active a128 evidence or bypassing spec synchronization. No archive was executed.

Function Logic Map: not-applicable — a147 adds a Python leaf; no existing function is changed, and the Go AST mapper does not process Python.

## Manager-approved base reanchor audit — 2026-10-18

- Frozen base: `936d7e0d23510277224f5b16655169b1c80a07be`; requested replacement: `055f47f78f9478d4a7522a70bb461789bd7b937e`.
- Attribution: no commit touches the a147 change directory or tool; consequently zero a147 own non-merge Go commits after the old base. No earlier a147 directory name was identified.
- The old-base gate requested 14 existing Go functions. All six affected paths were changed by sibling a146 commits `f6038a784` and/or `34ffe3c93`, not a147. Evidence: `/tmp/a147-current-gate.log` and `git log 936d7e0d..055f47f78 -- <six paths>`.
- Offline implementation SHA-256 remains `2cf098c6a44f84135892dbfc74fd54ee11e42bffe3084e36cd09c45139a600ee`; tests `f4bb3dc796cc174f2543c80dc306550f91467bdc2ff870481245ca3317329096`; README `27fa3952f1a5f107b1acfcc72e4adb52780b660aa3f7c3c4730fce7216bf526a`. These match the prior isolated gate inputs.
- Manager explicitly approved this base-only reanchor after attribution audit, as a prerequisite for a current-HEAD gate. It does not approve archive or final commit; existing evidence is retained. The earlier gate remains evidence only for the old baseline.

## Independent-review correction — 2026-10-18

Invalid state digests formerly retained selected/nonselected status and mislabeled the absent outcome as missing. The candidate now remains in the unknown cohort with `invalid_identity` for both MFI and label. End-to-end selected and nonselected fixtures both failed before the correction (`/tmp/a147-invalid-identity-red.log`, exit 1) and all 19 focused tests passed afterward (`/tmp/a147-invalid-identity-green.log`, exit 0).

Corrected source SHA-256 `c8d802e3f6423e100353310412923c845e50464bb637e391c81c33eaeb395c75`; corrected tests `ee3b03145f2fdd21e0fda36c1f27ab3bc9e417289a67f5529d6ed623eb870bbc`. Earlier hashes and gate results above remain historical and do not validate this correction. Independent re-review and a new gate are pending.

Independent fixed-snapshot review: PASS with no unresolved findings, by a separate Codex-host context applying the gstack checklist and security, testing and red-team checklists. The report is `analysis/reviews/independent-gstack-review.md`. No separate Claude or second-model pass and no automated gstack review-log pass were performed. This review does not establish aggregate-gate or archive completion.

## Current-baseline aggregate verification — 2026-10-18

On isolated HEAD `21e4a08a75e12e4c6c0aff5620f84cf42bedbad1`, with corrected source `c8d802e3f6423e100353310412923c845e50464bb637e391c81c33eaeb395c75` and tests `ee3b03145f2fdd21e0fda36c1f27ab3bc9e417289a67f5529d6ed623eb870bbc`, aggregate gate passed all 11 stages, exit 0 (`/tmp/a147-fixed-gate.log`, `/tmp/a147-fixed-gate.exit`). Full tests, test-seams, race, vet and OpenSpec validation passed. `GOFLAGS=-buildvcs=false` and `SDD_PYTHON` were scoped to the gate. CodeGraph and CodeGraphContext sync succeeded; GBrain was not synced and its advisory stale warning remains. Manager authorized formal archive and spec synchronization following this gate. This establishes offline implementation verification only; no market sessions were collected and no MFI performance conclusion follows.
