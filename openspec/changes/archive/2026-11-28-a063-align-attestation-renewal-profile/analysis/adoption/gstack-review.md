# a063 execution-baseline adoption — independent gstack review

## Verdict

**CLEAR for clean H2 and the refreshed, still-unbound adversarial review.** The
remaining record mismatch is intentional: H2 binds the previous adversarial
report digest, while the refreshed H2 adversarial review exists as the only
uncommitted worktree change. A successor metadata-only H3 must bind the actual
current adversarial and gstack review file hashes. This review does not approve
final gates, runtime work, task completion, or archival.

## Bound review state

- Planning base `P`: `da80ce31b6a1ab5d443016768f970a82bab102db`.
- Execution base `E`: `e65e394bf84b3c6e4559a219e816af96d341d75d`.
- Frozen source `S`: `c727ad12a42dcd15c494c1997e92816edaf17b6b`, tree
  `2d608c945d4b2a4a2fd87824bf06a42815d256a1`.
- Clean normalization commit `H2`:
  `3aa52ec2560c112c42849df83b93efb8b32f03e0` (parent H
  `8e97d2a3fc34af0386bb29ad3e00d9d872bec7c8`).
- Ledger SHA-256:
  `3d9f0743cbbc9d30939582e19680e2c3ae58f9c968b54312d76f34aad4478675`.
- H2 record SHA-256:
  `1614c804efae6ad35660513108a35725f02073c5c3131d1fb10cff826fe656e2`.
- Refreshed adversarial report SHA-256:
  `9597e26762d33258f6c6ba4a8d65a271880c008b556eaeb99359382aec5d9d44`.
- The current map-pair aggregate is
  `40604d2abd8c2f5564d7c2448a06a1b65bd5b7d9eb1ba42d12cd7136f15545de`.
  It was independently reproduced from the worktree root with:

  ```sh
  find openspec/changes/a063-align-attestation-renewal-profile/analysis/function-logic \
    -mindepth 2 -maxdepth 2 -type f \
    \( -name function-logic-map.md -o -name branch-test-map.md \) -print0 \
    | LC_ALL=C sort -z | xargs -0 sha256sum | sha256sum
  ```

## Independent checks

1. A clean detached clone at H2, using an external
   `PYTHONPYCACHEPREFIX`, ran:

   ```sh
   python3 tools/logic-map/check_analysis.py \
     --root /tmp/a063-gstack-h2-clean \
     --change a063-align-attestation-renewal-profile
   ```

   It exited `0` and printed `execution-baseline adoption exception evidence
   complete`. Neither the clean clone nor the reviewed worktree contained a
   repository `__pycache__` or `*.pyc` after the check.
2. Every tracked Go worktree blob equals `S`, and every worktree Go file has
   mode `0644`. `git diff --quiet S -- . ':(exclude)openspec/**'
   ':(exclude).sdd/**'` succeeded. H2 and the refreshed report therefore make
   no source, test, service, timer, or runtime change.
3. All nine active Function Logic Maps contain the exact five headings:
   `Inputs and invariants`, `Branches and early returns`, `Calls and live
   bindings`, `State mutations and fallbacks`, and `Safety conclusion`.
   Each paired Branch Test Map remains present with its AST and risk report.
4. The branchless `newSoakAttestCmd` has an explicit B1 happy-path map row
   limited to command registration/read-only metadata. It does not claim flag
   binding or `RunE` execution.
5. `Console.readAttestation` retains the repaired behavior: a blank trimmed
   path returns the initialized unknown view before either
   `readRenewalStatus` call; only nonblank load-error and loaded-success flows
   invoke that diagnostic. Source and both maps agree.
6. Generic-label/TODO scanning was clear, historical P artifacts remain
   byte-preserved, and `git diff --check` was clear.

## H2 to H3 rebinding condition

The committed H2 record contains the prior adversarial-review SHA-256
`5de872ba232e5bbd806d6e9d9024416b3cca30ea7c68bddabeaaa977e3abd763`,
not the refreshed H2 report hash above. H2 also contains the prior gstack
review digest; replacing this report necessarily makes that digest stale too.

The direct checker against the current worktree returns its clean-worktree
protection failure because the refreshed adversarial report is deliberately
uncommitted. That is expected while preparing H3; it is not a source, map,
cache, or validation failure. The clean H2 checker result above establishes
that the evidence set itself validates. H3 must write the two actual review
paths and final SHA-256 values, commit metadata only, and then rerun final
validation and required gates.

## Audit context

The earlier gstack BLOCK identified an overbroad `readRenewalStatus` map claim
and an opaque map aggregate. The map was corrected, and the refreshed
adversarial report now records the exact root-relative digest recipe reproduced
above. Those defects are closed; the record rebind is the only intended
pre-H3 condition.
