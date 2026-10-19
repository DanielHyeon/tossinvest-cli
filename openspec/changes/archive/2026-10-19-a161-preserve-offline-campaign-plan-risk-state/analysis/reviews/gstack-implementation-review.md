# a161 gstack implementation review

**PASS — scoped code/security/QA review.** No unresolved code blocker found in the snapshot below. Full repository gate attempts 1–3 did not pass; a fresh aggregate gate remains required. This reviewer has not executed gates. Metadata preparation, archive, commit/push and live readiness are not covered by this PASS.

Reviewer: Codex-host gstack review using installed review skill and two-pass checklist. No Claude/cross-model run claimed. Read source and implementer receipts; did not execute tests or modify production/test files. Reviewed all new codec/export/test files and native field/parser/binding definitions. Existing tracked diff is PM metadata; implementation is new files.

## Findings and resolution

The initial supplied codec snapshot was superseded during review by the UTF-8 preservation correction. Ordinary JSON marshaling replaces invalid native strings and keys; the retained behavioral RED demonstrates acceptance of an invalid key. The intermediate Encode marshaled then decoded and compared typed Evidence with reflect.DeepEqual, refusing a lossy result; this mechanism is superseded by the gate3 repair below. Both invalid key and invalid field cases are covered by the added continuation test. Final5 focused/race receipts cover the final corrected source. This report does not treat earlier passing receipts as proof of that correction.

Documentation base typo was reported and Terra corrected it; no unresolved code/authority finding remains.

Independent review subsequently found sum-preserving ceiling redistribution accepted by the initial codec. Retained ceiling RED proves this defect; current validation checks exact native 8:4:2, 2:4:8 or seven-leg remainder allocation. Quotient/remainder arithmetic stays within uint64 limits. Added lane tests compare native allocators for small quantities and MaxUint64 and reject redistribution. Final5 receipts cover this correction; previous report snapshot is superseded.

## Code and security assessment

- Exporters validate native plan/binding and weekly native risk seal, then copy arrays/maps/FX into detached data. They never call constructors to recreate native execution objects, change source state, recalculate qFinal or touch journal/runtime. Nil maps stay nil; empty maps remain empty; false latch entries and incomplete fill records survive. Weekly identity knowledge is explicitly named KnownIdentity, with reservation not captured and cancel not applicable.
- Exact amount strings retain reversal's native whitespace semantics. Numeric validation is bounded before Rat parsing: reversal's binary exponent is limited, and other families use decimal lexical checks. No global parser change was introduced. Filled/held overbudget values are not clamped or interpreted as permission.
- Decode checks 4 MiB size and external SHA before parsing, rejects unknown fields and compares canonical re-encoding byte-for-byte. Duplicate keys, alternate/missing fields, trailing content and noncanonical encodings cannot silently normalize into acceptance. Scope uses exact struct equality; errors return a wholly zero Evidence. Mutable DTOs are validated on Encode.
- Native digests/seals are preserved, not authenticated or reminted. This is caller-provided inspection evidence: capture time and unverified ledger currentness remain explicit. Codec structural validation is not a claim that every arbitrary caller-authored payload came from native capture.
- The new production leaf imports standard library only. Import/capture-reference guard scans internal/cmd/tools and rejects runtime imports and aliased/raw/escaped capture references; dot imports of lanes are rejected. Guard does not prove an eternal absence of every future same-package unqualified call. Current reviewed source has no runtime caller; this distinction is sufficient for this change.
- File roundtrip uses caller-owned temporary files via standard library. There is no storage daemon, path overwrite API, network/shell side effect, fsync durability promise or authority restoration.

## QA evidence and limits

Implementer final5 focused and race logs each report all four packages passing with EXIT_CODE=0. Tests include KR/US per lane, native fill/cancel transitions where supported, nonzero/unknown/overbudget risk, native digest/FX copying, nil/empty distinctions, exact zero result on error, strict format/scope boundaries and close/reopen. Assertions sample preserved fields; this is not an exhaustive proof for every possible native state or concurrent caller mutation. Caller synchronization remains required.

Scope mutant exits 1 with `foreign scope accepted`; latch mutant exits 1 with `snapshot lost`. These are behavioral checks of their specific protections, not proof of broker safety, complete risk reconstruction or universal mutation coverage. UTF-8 RED exits 1 with `invalid UTF-8 native map key accepted`. No reviewer rerun or live-market verification is claimed.

Final test-only follow-up explicitly rejects omitted false/nil fields, nested duplicate keys and a trailing object with an entirely zero error result. Final5 focused/race receipts include these tests. Native currency/digest/unused FX strings remain inspection data; this review does not impose a new three-letter currency policy or claim cryptographic provenance verification.

## Gate3 compatibility repair — scoped PASS

Read actual attempt3 failure: existing `TestNoModulePackageInTheShadowProductionClosureImportsUnsafeOrReflect` rejected codec production reflect in ordinary/tagged dependency closures. The guard was not relaxed or edited. Existing tracked source/test changes remain absent; new codec removes reflect and replaces the intermediate equality roundtrip with `validText` before validation/marshal.

Direct field inventory confirms all Evidence, Scope, Plan, Risk and optional FX strings are checked, including all six maps' string keys and every string value/nested record field. Booleans, uint64 fields, slices of uint64 and bool pointers cannot suffer JSON invalid-UTF-8 replacement; their nil/false distinctions remain unchanged. No native authority type, numeric rule or caller change was added.

New test uses reflection only in `_test.go` to produce one corrupted string at a time across structs, nonnil FX and populated maps, calling validText directly. The deliberately combined-family fixture does not make this a false-green semantic validation test: direct helper invocation isolates UTF-8 coverage from family rejection. It covers current fields; future new optional container types would require updating the generator and typed checker. Existing Encode/capture positive tests and native invalid-string negatives retain integration coverage.

Reviewed repair-focused and repair-race original logs: all five packages, including the previously failing strategyshadow guard package, report PASS and EXIT_CODE=0. These supersede final5 as current-source receipts. Retained scope/latch mutants describe the prior final4 snapshot only; they were not rerun by this reviewer and are not claimed to prove this repair. Attempts 1/2 history and renamed evidence bytes remain historical, not a successful aggregate gate.

## Reviewed snapshot SHA-256

| File / receipt | SHA-256 |
|---|---|
| `internal/campaignstateevidence/authority_guard_test.go` | `79f219f533050f6c1c16727e6563f42b6a8d92a80831b5c8b8bced07d8a4f714` |
| `internal/campaignstateevidence/codec.go` | `209ef42a97ad73e0fcae187bec505edd487b2602f861f2480808b1c6597b26ee` |
| `internal/campaignstateevidence/codec_test.go` | `4a97df9da2a76822c8f571c383166934e8a6e1a4081e70dc1d5c93f379d24840` |
| `internal/continuationlane/offline_evidence.go` | `34f73baf0abe41be85f3fd086a16983b999451f6fc140eb7cc2b131476c51414` |
| `internal/continuationlane/offline_evidence_test.go` | `47a692626ad3f046730498c64a4f7d6b44ad19061c19f190a1818c134f9ddb2b` |
| `internal/reversallane/offline_evidence.go` | `5f2cd3c5d23c0a7b9273fad7a50ddaa98536e41f98f701a63676ace64a548117` |
| `internal/reversallane/offline_evidence_test.go` | `f3cb1da69d190dcddaf74dbf71fbccf9d87b135b2dfd85ba071d0f7d4e09215b` |
| `internal/weeklyvaluelane/offline_evidence.go` | `7da848ccaea063fd9664096f6e49402a89b9c83310b74609343830c0f426f639` |
| `internal/weeklyvaluelane/offline_evidence_test.go` | `d46d9754fa6fc01fe6de46da60c4a2f808c761b05672c51e5d3c782f99e286dd` |
| `openspec/changes/a161-preserve-offline-campaign-plan-risk-state/analysis/field-matrix.md` | `5756b2fe9ccd97fdc85732854d11860b22f3be8cca61f219c000db49a6c4e66b` |
| `openspec/changes/a161-preserve-offline-campaign-plan-risk-state/design.md` | `994f36ed006002c0be5eae7a8039dec049080a89aa05460c6961cfe9a94992cb` |
| `openspec/changes/a161-preserve-offline-campaign-plan-risk-state/specs/campaign-state-evidence/spec.md` | `d1b665ec8f0dc4fe67cc530dbc6dc7676b48a5630aa267257843d68841ee6037` |
| `/tmp/tossos-a161-focused-final5.log` | `f6ee830b3056e81275053071999d7adea578432a0370c435623c8645d69eebd3` |
| `/tmp/tossos-a161-race-final5.log` | `2ee31c70fb05bb3653ba7321f99ee3350f5ea487fec7705ccd58547eb6ed24f9` |
| `/tmp/tossos-a161-scope-mutant-final.log` | `4ebb60fbebbb7f46c22d163886a48638c0c67ec7d7c21257371d3aef9f4fe95c` |
| `/tmp/tossos-a161-latch-mutant-final.log` | `603fa60cb5f130d884b12b8f64de8c6cef55cd29cd4a3a1dcb16aeab56a0b552` |
| `/tmp/tossos-a161-utf8-red.log` | `88eb3479f1f1f3de95dc91d35f25125071a447b0b2da92de34d506c2e622fe24` |
| `/tmp/tossos-a161-ceilings-red.log` | `df0b044249329b5b8a9cdb05184f9bd3a554efdfb9fa91c5634cd5cbea7f2402` |
| `/tmp/tossos-a161-repair-focused.log` | `17f3bc18fc5e1ddbcf1485ee9e5f0d560fc76197371b4457577e38fc2e5bc664` |
| `/tmp/tossos-a161-repair-race.log` | `87d8a372590dae6fac7cc350bf945197a8eff0ba8c861ff27813e2e5851c56cf` |
| `/tmp/tossos-a161-attempt3-gate.log` | `69483f66d3e7713acbf0bb2df9d578dc911d2328a540863c959410205e7442fb` |
