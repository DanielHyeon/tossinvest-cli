## 1. Prepare implementation

- [x] 1.1 Obtain implementation authorization and independent design reviews; verify Story, predecessor/spec synchronization order and repository READY requirements.
- [x] 1.2 Freeze the implementation baseline and obtain fresh CodeGraph evidence; produce AST and function/branch maps before editing existing functions, as applicable.

## 2. Offline input-quality evaluation

- [x] 2.1 Implement snapshot manifest, integrity/link validation and deterministic input-quality sidecar without modifying original evidence.
- [x] 2.2 Implement same-cycle, same-K comparisons and market/session-separated reporting, including per-quality-stratum counts, exclusion/selection rates, mean and label gaps, incomplete labels, overlapping strata, dependent windows and inconclusive samples.
- [x] 2.3 Add offline tests for missing versus empty versus unsupported input, observation-time freshness, invalid hashes/joins, cycle exclusion, K=0 and fewer-than-ten selection counts, same K, ties and state deduplication.

- [x] 2.4 Repair advisory stats timeout handling and test both hard-success and hard-failure outcomes.

## 3. Separate J2 comparison

- [x] 3.1 Implement the fixed r2 binary wording in design and a provenance manifest; keep r1, J1 question and original J1 scores unchanged in the paired replay.
- [x] 3.2 Implement offline pairing and explicit BLOCKED status for absent/incompatible saved r2 responses; synthetic fixtures exercise plumbing only.
- [x] 3.3 Test incomplete-pair cycle exclusion, insufficient K, mismatch, no-response, future-label leakage and contaminated snapshots, fixed-J1 and no-network behavior using synthetic fixtures; do not make real model calls.
- [x] 3.4 Perform a real empirical r2 comparison only if compatible already-saved responses exist; otherwise record BLOCKED. Any new model request requires separate authorization and is outside current scope.

## 4. Review and completion

- [x] 4.1 Independently review evidence, statistical limitations and unchanged order/operating boundaries.
- [x] 4.2 Run applicable offline checks and individual gate components; record failures, corrections and targeted rechecks accurately.
- [x] 4.3 Review and verify the formal spec synchronization and archive/PM path plan; perform archive only after the full gate passes.

Current evidence: the actual r2 comparison is BLOCKED due to absent original request anchor/model-version evidence. In the isolated checkout, hard index freshness and individual SDD checks passed; the full test passed after environment repair. The initial seams failure and successful targeted rechecks are recorded in review; corrected full seams and the final aggregate gate passed.

Full aggregate gate: PASS, 11/11, exit 0 in the isolated checkout; see review and external raw log.

Post-gate lifecycle: formal spec synchronization, archive and PM path update completed; commit and push are not yet claimed.
