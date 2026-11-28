# a063 execution-baseline adoption — independent adversarial review

## Verdict

**CLEAR for the adoption evidence and the H2 map normalization, subject to rebinding this refreshed review in a successor metadata-only commit.** This review does not approve an operational action or declare final gates complete.

## Bound snapshot

- Review worktree: `/tmp/tossos-a063-execution-adoption`, detached `H2` `3aa52ec2560c112c42849df83b93efb8b32f03e0` (parent `H` `8e97d2a3fc34af0386bb29ad3e00d9d872bec7c8`); frozen source `S` `c727ad12a42dcd15c494c1997e92816edaf17b6b`, tree `2d608c945d4b2a4a2fd87824bf06a42815d256a1`.
- Immutable bases: planning `P` `da80ce31b6a1ab5d443016768f970a82bab102db`; execution `E` `e65e394bf84b3c6e4559a219e816af96d341d75d`.
- `execution_baseline.py` SHA-256: `e9d8a55c718fe3568d09facb355e507c06c565f5bed915ef1b0f780d0104e1c3`.
- Draft ledger SHA-256: `3d9f0743cbbc9d30939582e19680e2c3ae58f9c968b54312d76f34aad4478675`; unbound record SHA-256: `171fa084c3d0041c89fedecbbe8a3ac6512a00f705812ed1300b0f98faa1fe3c`.
- The nine active map pairs comprise 18 files. Their reproducible aggregate SHA-256 is `40604d2abd8c2f5564d7c2448a06a1b65bd5b7d9eb1ba42d12cd7136f15545de`, computed with:

  ```sh
  find openspec/changes/a063-align-attestation-renewal-profile/analysis/function-logic \
    -mindepth 2 -maxdepth 2 -type f \
    \( -name function-logic-map.md -o -name branch-test-map.md \) -print0 \
    | LC_ALL=C sort -z | xargs -0 sha256sum | sha256sum
  ```

## Independent checks

- Recomputed canonical ledger payload from immutable Git history: exact equality. `P..E` has 315 commits and 329 modified-existing function rows; `E..S` has 3 commits and 9 rows. Both inventories match their recorded SHA-256 values. The `E..S` 12-path Go inventory is exactly the fixed allowlist.
- The record binds the exact `S` commit/tree and ledger digest. All 15 entries in `analysis/gstack/implementation-reviewed-digests.json` match the source snapshot. All four review fields are empty, as required before actual review evidence exists.
- Worktree Go files equal `S`; no source change is hidden in the metadata preparation. The historical planning-base directory retains 11 original bundles (45 files including its provenance README) and explicitly leaves the 329-row historical debt unclaimed.
- Reviewed all nine E-based map pairs. Each now names source-derived conditions, outcomes, calls/effects, and focused test evidence or an explicit unmeasured limit. The three deleted E tests are labeled immutable historical behavior only, avoiding a false current-runtime claim.
- `git diff --check` exits 0. The prior three risk-report EOF blank lines are gone.
- No repository `__pycache__` or `.pyc` remains. Re-running the pre-bind checker under `PYTHONPYCACHEPREFIX=$(mktemp -d /tmp/a063-final-prebind-pycache.XXXXXX)` created no repository cache. This prefix remains required for every later Python/make verification command.
- `python3 tools/logic-map/check_analysis.py --change a063-align-attestation-renewal-profile` exits 1 only because the untracked draft `execution-baseline.json` is absent from committed `S`; its fail-closed diagnostic is expected before bindings and `H` exist.

## H2 normalization re-review

- All nine active Function Logic Maps contain the five exact checker headings: Inputs and invariants; Branches and early returns; Calls and live bindings; State mutations and fallbacks; Safety conclusion. The normalized prose preserves the source-derived branch effects, call/error limits, and safety conclusions reviewed above.
- The branchless `newSoakAttestCmd` branch map now has a B1 happy-path row. It is limited to command registration/read-only metadata and explicitly does not claim flag-binding or `RunE` execution.
- `Console.readAttestation` retains the repaired claim: blank trimmed `Attestation` returns before either `readRenewalStatus` call; only nonblank load-error and loaded-success flows run that diagnostic.
- `PYTHONPYCACHEPREFIX=$(mktemp -d /tmp/a063-h2-final-pycache.XXXXXX) python3 tools/logic-map/check_analysis.py --change a063-align-attestation-renewal-profile` exited 0 and printed that the execution-baseline adoption evidence is complete. It left no repository `__pycache__` or `.pyc`; `git diff --check` also exited 0.
- H2 changes only OpenSpec map-format metadata and remediation evidence. Every Go path remains byte-identical to S.
- The H2 record currently binds the prior adversarial-review digest. Because this report is refreshed for H2, its new digest must be written to `execution-baseline.json` and committed in a successor metadata-only detached commit before final validation.

## Required next sequence

Rebind this refreshed adversarial-review SHA-256 in `execution-baseline.json` and commit a successor metadata-only detached commit. Then rerun final adoption validation and the separately required gates. This review does not approve an operational action or declare those later checks complete.
