# a152 gstack implementation review

Current verdict: scoped implementation review PASS after timestamp corrections and rereview. No unresolved actionable finding remains. This is not a full-gate or empirical-data approval. Codex-host gstack code/security/QA review, not a Claude or cross-model execution. Reviewer executed no implementation, tests, CLI, gate or lifecycle operations. Only this assigned report edited. Full gate remains pending.

## Initial finding — resolved (history preserved)

P2 reported by the independent reviewer and acknowledged by Terra: `stamp` accepts unrestricted fractional timestamp precision through `datetime.fromisoformat`. Precision beyond Python datetime microseconds can be discarded, making a received timestamp slightly after as_of compare equal and be accepted. Exact temporal fail-closed qualification must reject unsupported precision before parsing or preserve it exactly. Source inspection confirms no precision guard in the reviewed snapshot. Await implementer behavioral RED/GREEN and final source hash; no reviewer-run reproduction claimed.

## Scope inspected

Read complete new Python implementation, 25-test source, README, approved design/spec/tasks, synthetic fixture/output and research/verification evidence. Initial held source SHA-256 `d6854d5610a558e3353a8c39591a376774d6e10cd28878c6e009dd486a6e0743`; test SHA-256 `de51e8af2a1e294a5a69ff4ceccc63e25521c2f75f937b7d8251b3e903ea3810`. These are a held snapshot, not a frozen approval.

Other inspected boundaries are consistent with the narrow contract:

- Exact bounded decimal volume sum and zero-origin grid; explicit bin interval/width/zero-inclusive coverage. Empty and all-zero evidence remain not_ready without manufactured LVN/HVN.
- Distinct cutoff and information as_of, prefix/prior date bounds, bin build/receipt ordering and trade execution/receipt ordering. Received ordering permits delayed older executions. Final strict timestamp guard below resolves the precision defect.
- Duplicate IDs, corrections/cancellations and unknown/unresolved correction policy never become eligible or silently deduplicated. Incomplete coverage and source/access mismatch fail closed; no partial total on rejected evidence.
- Classification stays separate from empirical authority. Synthetic eligible fixtures keep vp_input_eligible=false; authenticity_verified, coverage_verified and production_authorized remain false. Missing actual data completes unsupported and leaves non-profile research unblocked.
- Strict envelope/field/key checks, bounded records/datasets/strings/decimal input, isolated dataset failures, stable error reasons and no raw event disclosure. No network, shell, subprocess, live trading or authentication imports in the implementation.
- Exclusive O_EXCL/O_NOFOLLOW output creation and no-follow parent traversal prevent existing-path and symlink overwrite. Input hash/preservation and no crash-atomic durability overclaim are explicit.
- Validated output now retains market/symbol/currency/basis/session/bin method and scope metadata. Manager's earlier missing-metadata concern is corrected in this snapshot.
- Stdlib leaf reuses local patterns without modifying/importing a151 or production paths. Research report names inspected sources and explicitly skipped external channels; no whole-disk absence or source-authentication claim.

## Evidence read, not reviewer execution

`analysis/verification/focused-green.log` records 25 PASS tests, run separately from aggregate gate. Initial RED is explicitly a missing-module import failure, not behavioral TDD. CLI stored output SHA-256 `5c1fc88ebafce240872fbc7f0451d21ca717e8e9ed527022e2d90e37113fd7e2` verified by file read/hash; output input hash matches synthetic manifest bytes. Five classifications and exact trade/bin totals of 5 shares confirmed; every synthetic empirical flag remains false. These initial receipts do not cover the precision defect; corrected receipts below supersede them.

## Corrective rereview

Verified TIME.fullmatch runs before datetime parsing for every timestamp field. Full seconds and optional 1–6 fractional digits are required; fractional/second-bearing offsets, excess precision and alternate separators are rejected. Calendar validity remains checked by datetime and session IANA checks. Offset HH is bounded 00–23 and MM 00–59, preventing normalization of +08:60 to +09:00.

Read precision RED (two actual failures), precision GREEN (27 PASS), then offset RED (invalid offset wrongly eligible) and final offset GREEN (28 PASS). Both reviewers identified the offset follow-up. Tests cover future submicrosecond trade/bin receipts, unsupported lexical variants, valid microseconds, and invalid session/trade/bin offsets. No implementation/test execution by this reviewer. Full gate has not run as part of this review.

Final source SHA-256 `34c09bf0c7b250aa9fd639bd1d637fdc572f0cd18dae9e137db94f45cf22aa63`; tests `5316db61f95c0d6abb955a1c9da810429990cdc3b8d87ff41ee3c6d7b9c7ab8b`. Fresh CLI output after offset fix retains SHA-256 `5c1fc88ebafce240872fbc7f0451d21ca717e8e9ed527022e2d90e37113fd7e2`; file hash inspected and matches the previously inspected synthetic output. Evidence distinguishes initial import RED, behavioral failures and superseded receipts. Approved design and README document the timestamp subset.

Actual inspected tape/bin availability remains unsupported, independently of synthetic structural success. No runtime authorization, data collection, trading or profitability claim follows from PASS.

## Final metadata refresh

Compared all 15 manifest entries after Manager acceptance: only tasks.md changed. Completion checkboxes explicitly describe preparation; actual sync/check, aggregate gate and lifecycle success still require receipts. Manager review.md acceptance and evidence were read; no code/test/design change detected. Earlier pending-review sentences in implementation evidence are historical and superseded by this report and Manager acceptance. No gate completion is inferred.

## Reviewed manifest

Source/tests/docs/fixture and relevant evidence; reviewer reports excluded. Later edits require refresh.

Manifest SHA-256: `53d9904e239b518b31dd21f946b5e923073d21df7498846992fa1c9e1305c310`

```text
34c09bf0c7b250aa9fd639bd1d637fdc572f0cd18dae9e137db94f45cf22aa63  tools/a152-qualify-volume-profile-evidence/analysis.py
5316db61f95c0d6abb955a1c9da810429990cdc3b8d87ff41ee3c6d7b9c7ab8b  tools/a152-qualify-volume-profile-evidence/test_analysis.py
3679037629821408b6838694f6892e7b45d8591a76b2b09c37417b392de6c66d  tools/a152-qualify-volume-profile-evidence/README.md
1f9c0086e6df75681be2d1e410f9a9e90dd443f0886310a28ec1997ed2e2ccf3  tools/a152-qualify-volume-profile-evidence/testdata/synthetic.json
4f227cebf2eafd86f42bf90a7936733ff78413808d080f6cfcfea08050599bf8  openspec/changes/a152-qualify-volume-profile-evidence/proposal.md
9c33cca33d8b0323eb71fbf11129a80a22ba9f650bf5ccdb3caea0dd251497f3  openspec/changes/a152-qualify-volume-profile-evidence/design.md
3b164f0fa793fe3aa670148a742cf18a7ba960f844b74550fcae3ae140f4e896  openspec/changes/a152-qualify-volume-profile-evidence/specs/candidate-discovery/spec.md
b724012b19cdddfa1e13bb45cf6c6b2a354d7e70d230c6182be7cc67b84b8481  openspec/changes/a152-qualify-volume-profile-evidence/tasks.md
5679baa0fddb8370ccab6ec4186f366affc5e6ba1eb78f5f12e1a25894a7e155  openspec/changes/a152-qualify-volume-profile-evidence/analysis/research-availability.md
6878c4b151710ffb5f88065ad0a2760cf140a9eb5112c2ccc6896ad492e8b5e9  openspec/changes/a152-qualify-volume-profile-evidence/analysis/verification/implementation-evidence.md
e6c7e7e5d905e305ddd83095bc08d87343ca02097fcfdfd7f06a6a9e845e47d4  openspec/changes/a152-qualify-volume-profile-evidence/analysis/verification/precision-red.log
7277bd63d4fe6063c27dae0daf36c2368f76a59ba02270f14fe9120811390b52  openspec/changes/a152-qualify-volume-profile-evidence/analysis/verification/precision-green.log
9224d34e23545c466572a03ac0fb84c471ccea0f9746f0ab0917665dfe2fed25  openspec/changes/a152-qualify-volume-profile-evidence/analysis/verification/offset-red.log
5d78e41521f16b3128b98c5056e10e3146c96f0d12510a84e38411de6f19696d  openspec/changes/a152-qualify-volume-profile-evidence/analysis/verification/offset-green.log
5c1fc88ebafce240872fbc7f0451d21ca717e8e9ed527022e2d90e37113fd7e2  openspec/changes/a152-qualify-volume-profile-evidence/analysis/verification/synthetic-cli-after-offset.json
```
