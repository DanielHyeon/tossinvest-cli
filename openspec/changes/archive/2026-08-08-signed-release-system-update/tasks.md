## 1. Contract and Evidence

- [x] 1.1 Capture the pre-implementation base commit and strictly validate the change
- [x] 1.2 Complete proposal-freeze review with adversarial engineering and supply-chain security voices
- [x] 1.3 Sync CodeGraph and record definition/caller/impact evidence plus Function Logic Maps for every edited existing function

## 2. Release Provenance

- [x] 2.1 Add RED tests for fixed latest-release selection, bounded HTTPS redirects/responses, and unsupported releases/platforms
- [x] 2.2 Add RED tests for Sigstore digest, workflow/tag identity, SLSA predicate, transparency evidence, and fail-closed bundle handling
- [x] 2.3 Implement the pinned Sigstore verifier and GitHub release/attestation client
- [x] 2.4 Update the release workflow to pin actions, minimize job permissions, attest every completed platform archive, forbid clobber/rerun mutation, and add a static workflow regression test

## 3. Safe Candidate Staging

- [x] 3.1 Add RED tests for official auth-helper directories plus archive traversal/link/duplicate/multistream/PAX/sparse/size rejection and exact regular binary extraction
- [x] 3.2 Add RED tests for atomic fixed candidate publication, wrong binary rejection, sync/rename restoration, no-prior-candidate uncertain-state cleanup, and preservation of an earlier candidate
- [x] 3.3 Implement verified archive extraction and `localupdate` fixed candidate staging without executing candidate code

## 4. Console Integration

- [x] 4.1 Add RED route/race tests for session/CSRF, ignored selector fields, success/refusal notices, no install/relaunch, install-commit refusal, and reverse-completion two-tag serialization
- [x] 4.2 Add the signed-release settings model, download handler, fixed wiring, and separate review/install UI
- [x] 4.3 Add a real CLI assembly regression test proving the production console wires the official release verifier and fixed sibling publisher

## 5. Documentation and Verification

- [x] 5.1 Document signed-release prerequisites, first attested release behavior, trust identity, network endpoints, failure recovery, and the one-time pre-menu bootstrap
- [x] 5.2 Run focused tests, race tests, full tests, vet, strict OpenSpec validation, `make sdd-sync`, and `make sdd-check`
- [x] 5.3 Run `make gate CHANGE=signed-release-system-update`, obtain independent implementation review, and retain a verified episodic learning
