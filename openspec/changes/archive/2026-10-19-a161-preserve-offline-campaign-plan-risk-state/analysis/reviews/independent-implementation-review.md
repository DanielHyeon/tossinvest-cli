# a161 independent implementation review

Final verdict: PASS for the bounded offline implementation. Initial HOLD below is preserved as review history; all actionable findings are resolved. Gate attempt 3 subsequently failed its unchanged shadow dependency guard; see repair addendum below. Full gate retry is pending. This review is source/receipt inspection; reviewer did not modify implementation or execute tests.

## Findings history

1. P2: mutable evidence validation accepts sum-preserving ceiling redistribution. `codec.go` checks ceiling sum and length plus fixed weights, but continuation Q=14 `[14,0,0]` passes despite 8:4:2 allocation; analogous reversal and weekly examples pass. Existing negative test changes total, so does not cover this case. Native capture validates its original plan; issue is public Encode/Decode validation after DTO mutation. Sent to Manager and Terra; disposition pending.
2. Manager UTF-8 finding: Go JSON replaces invalid string/key bytes with U+FFFD. Revised Encode marshals, unmarshals and requires whole-record DeepEqual, rejecting lossy input. Source inspected; `/tmp/tossos-a161-utf8-red.log` shows actual native-key regression failure, final3 focused/race logs show EXIT_CODE=0. Resolved in current source.
3. Documentation: field-matrix reversal cancel values were labelled fingerprints; native ApplyCancelRisk stores ReleaseHeldMinor strings. Export source copies these strings correctly. Correction requested; no implementation defect.

## Source assessment

All native plan fields, family-specific FX fields and seal data, risk amounts, latch map values, native dedupe records are copied by family-owned exporters. No native object is restored. Risk binding is checked before capture; weekly native risk seal is checked. Continuation/reversal publicly mutable native state is not provenance authentication. Original plan digest functions and existing execution functions are unchanged.

Nil/empty maps remain distinct; maps, ceilings, and FX values/boolean pointers are detached. Unknown nested fill risk may remain empty; required aggregate amounts cannot become zero by omission. Weekly Known is represented as KnownIdentity, alongside unknown latch and explicit reservation-not-captured/cancel-not-applicable status. No remaining-budget arithmetic, fresh FX mint, qFinal, journal write or runtime consumer is introduced.

Decoder bounds input at 4 MiB, requires external exact content hash and complete expected scope, rejects unknown fields and requires byte-identical canonical re-encoding. This also rejects omitted fields, duplicate keys, trailing data and alternate encodings. Errors return zero Evidence. Integrity and currentness limitations remain mandatory fields. Capture relies on caller synchronization; tests do not prove concurrent map mutation safety.

## Evidence inspected

- `/tmp/tossos-a161-focused-final3.log`: four packages PASS, EXIT_CODE=0.
- `/tmp/tossos-a161-race-final3.log`: four packages PASS, EXIT_CODE=0.
- `/tmp/tossos-a161-scope-mutant.log`: foreign-scope behavior test fails, EXIT_CODE=1.
- `/tmp/tossos-a161-latch-behavior-mutant.log`: source latch-copy mutation detected by snapshot preservation test, EXIT_CODE=1.
- `/tmp/tossos-a161-utf8-red.log`: native invalid-key counterexample fails before correction, EXIT_CODE=1.

KR/US roundtrip fixtures cover all three families; real native fill/cancel calls supplement seeded unknown/overbudget states. Weekly unknown Known=true fixture is explicitly seeded and resealed, not real-account evidence. File roundtrip uses caller-owned temporary files only; no crash/power-loss durability is claimed. Existing native digest is compared to captured/restored digest; this is preservation evidence, not authentication or a new independent digest algorithm proof.

## Final disposition and receipt

Fixed-allocation finding resolved. Codec now verifies exact native ceiling distribution using bounded quotient/remainder arithmetic (8/4/remainder, 2/4/remainder, six equal seventh floors plus remainder). No multiplication overflows at MaxUint64. Three lane tests compare codec acceptance against existing native allocator output at small and maximum quantities; redistribution negatives reject equal-total wrong distributions. `/tmp/tossos-a161-ceilings-red.log` records the original behavioral failure. Original native digest algorithms remain untouched.

UTF-8 fix and reversal cancel documentation correction remain present. Latest codec tests additionally reject omitted false/nil fields, nested duplicate keys and a trailing object with whole-zero result. These tests were inspected directly. Current focused/race final5 logs both show all four packages PASS and EXIT_CODE=0. Final scope/latch mutant logs still fail for the intended behaviors; those mutants precede only the final extra malformed-input test cases, with implementation unchanged.

Full-field preservation is supported by source-by-source field inspection plus selected roundtrip/assertion tests, not a claim that every individual field has a separate mutation test. No live evidence, latest-account risk, native authority restoration or crash durability is established. Full gate remains pending and must pass independently before lifecycle completion.

Current receipts:

- focused-final5 SHA256 f6ee830b3056e81275053071999d7adea578432a0370c435623c8645d69eebd3
- race-final5 SHA256 2ee31c70fb05bb3653ba7321f99ee3350f5ea487fec7705ccd58547eb6ed24f9

Reviewed source/test SHA256:

```
79f219f533050f6c1c16727e6563f42b6a8d92a80831b5c8b8bced07d8a4f714  internal/campaignstateevidence/authority_guard_test.go
ba42e8bb209b8d779d5adcf445d947068f8867ec27e9ca5f09591fd2aa56a11e  internal/campaignstateevidence/codec.go
f03e613f25eebc3067693a41c07cded273cfcc3763525aad105495dbf379b750  internal/campaignstateevidence/codec_test.go
34f73baf0abe41be85f3fd086a16983b999451f6fc140eb7cc2b131476c51414  internal/continuationlane/offline_evidence.go
47a692626ad3f046730498c64a4f7d6b44ad19061c19f190a1818c134f9ddb2b  internal/continuationlane/offline_evidence_test.go
5f2cd3c5d23c0a7b9273fad7a50ddaa98536e41f98f701a63676ace64a548117  internal/reversallane/offline_evidence.go
f3cb1da69d190dcddaf74dbf71fbccf9d87b135b2dfd85ba071d0f7d4e09215b  internal/reversallane/offline_evidence_test.go
7da848ccaea063fd9664096f6e49402a89b9c83310b74609343830c0f426f639  internal/weeklyvaluelane/offline_evidence.go
d46d9754fa6fc01fe6de46da60c4a2f808c761b05672c51e5d3c782f99e286dd  internal/weeklyvaluelane/offline_evidence_test.go
```

## Gate attempt 3 repair addendum — PASS

The original aggregate failure is preserved: `/tmp/tossos-a161-attempt3-gate.log` reports `TestNoModulePackageInTheShadowProductionClosureImportsUnsafeOrReflect` rejecting direct production reflect imports in both dependency closures, then EXIT_CODE=2. No guard exception or weakening was introduced; `internal/strategyshadow/guard_test.go` has no diff against HEAD.

Latest codec removes production reflection and replaces marshal/unmarshal DeepEqual with typed `validText` before validation and encoding. Direct review matched every present string field in Evidence, Scope, Plan, Risk and optional FX, plus every family map key and string/record-string value. Numeric, boolean and pointer fields retain exact JSON representations; no custom marshaler is present. Therefore the current data model retains the original invalid-UTF8 fail-closed guarantee without production reflection. Decode still re-encodes and compares exact canonical bytes, rejecting external invalid UTF8 rather than accepting JSON replacement text.

`TestTextValidationCoversEveryStringField` uses reflection only in test code. It walks each struct/string, populated FX pointer and populated maps, independently replacing a string or map key/value with invalid UTF8; shallow outer clones receive separately copied changed maps/pointers. It calls validText directly, so unrelated family-validation failures cannot falsely satisfy the negative cases. Existing successful Encode/roundtrip tests prevent an always-false checker from passing the suite. The generator does not traverse slices, but present slices contain only uint64, so no present text field is omitted. Future string slices or other new shapes need coverage expansion; this is not a current defect.

Latest five-package focused and race receipts include strategyshadow and end EXIT_CODE=0. Reviewer inspected source and receipts only, did not rerun tests. This is a scoped repair PASS, not overall gate PASS. Earlier pre-gate receipts and historical descriptions remain above; current implementation no longer uses DeepEqual in production.

Current hashes superseding codec entries above:

```
209ef42a97ad73e0fcae187bec505edd487b2602f861f2480808b1c6597b26ee  internal/campaignstateevidence/codec.go
4a97df9da2a76822c8f571c383166934e8a6e1a4081e70dc1d5c93f379d24840  internal/campaignstateevidence/codec_test.go
506f18bc413e11f5d91c20fa8f6ec474199472a4b131bace52fc76e526d68ea6  internal/strategyshadow/guard_test.go
69483f66d3e7713acbf0bb2df9d578dc911d2328a540863c959410205e7442fb  /tmp/tossos-a161-attempt3-gate.log
17f3bc18fc5e1ddbcf1485ee9e5f0d560fc76197371b4457577e38fc2e5bc664  /tmp/tossos-a161-repair-focused.log
87d8a372590dae6fac7cc350bf945197a8eff0ba8c861ff27813e2e5851c56cf  /tmp/tossos-a161-repair-race.log
```
