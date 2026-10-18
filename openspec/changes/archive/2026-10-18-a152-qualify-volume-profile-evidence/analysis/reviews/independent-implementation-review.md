# Independent a152 implementation review

Verdict: IMPLEMENTATION REVIEW PASS. Initial HOLD findings below are resolved in the final reviewed source. This is not aggregate gate acceptance or proof of real evidence availability.

## Finding history

Initial source used unrestricted datetime.fromisoformat. A supplied receipt one nanosecond later than as_of was silently truncated to equality and could qualify. Independent review reported this P2; writer confirmed. Work paused on Manager instruction, then resumed after collision check.

The revised source dca781a2e8ff2d0d598b5fdf62a1b3f2be8db9e1b10d5224218415ea0a74d6e1 adds strict pre-parser syntax: full seconds, optional 1–6 fractional digits, Z or HH:MM offset. Inspected behavioral precision-red.log (two failures) and precision-green.log (27 tests PASS). Excess fractional precision and fractional/second offsets are now rejected; the original precision defect is resolved.

## Follow-up P2: offset minute overflow (resolved)

TIME accepts arbitrary two-digit offset hours/minutes. Offset +08:60 can be normalized by datetime to +09:00 and then pass KR session offset validation, despite the newly explicit RFC3339/no-normalization contract. Requested writer reproduction and pre-parser HH00..23/MM00..59 bounds. Reviewer did not execute tests.

## Other reviewed boundaries

Read complete implementation and tests. Exact trade/bin sums, event and receipt bounds, independent cutoff/as_of, coverage declarations, correction refusal, synthetic downstream exclusion and non-profile research freedom are present. Eligible results retain interpretation metadata; errors avoid raw record payloads and preserve previously validated metadata where available. Malformed datasets are isolated from valid siblings. Input bounds, duplicate identities/keys, exclusive/no-follow output creation and parent symlink refusal are implemented. No collection/network/runtime invocation or LVN/HVN computation appears in source.

Only this review report edited. Source/tests remain writer-owned. Current receipt proves synthetic test coverage only; no actual tape authenticity/completeness or profitability claimed. Final rereview and hashes follow below.

## Final resolution and evidence

Final TIME grammar bounds offset hours to 00–23 and minutes to 00–59 before datetime parsing, retaining the 1–6 fractional second limit and full timestamp syntax. All timestamp fields use stamp(). The final offset regression covers session, trade receipt and bin receipt; accepted six-digit precision remains tested. Neither normalization nor silent truncation can bypass those supplied grammar bounds.

Inspected offset-red.log: one behavioral failure reproduces the +08:60 false eligibility. Inspected offset-green.log: 28 tests PASS. Earlier two behavioral precision failures and 27-test success remain historical evidence. Final CLI output was read independently: five synthetic classifications, 5-share trade/bin totals, empirical eligibility false and authority flags false. Input hash and recorded CLI output hash agree with supplied evidence. No remaining actionable finding identified in reviewed source.

Reviewer performed no tests or CLI execution; writer receipts and test/source inspection support this verdict. Aggregate gate remains pending. Caller assertions remain unverified and actual available tape/bin evidence remains unsupported in the documented research scope.

## Final reviewed SHA-256

- `tools/a152-qualify-volume-profile-evidence/analysis.py`: `34c09bf0c7b250aa9fd639bd1d637fdc572f0cd18dae9e137db94f45cf22aa63`
- `tools/a152-qualify-volume-profile-evidence/test_analysis.py`: `5316db61f95c0d6abb955a1c9da810429990cdc3b8d87ff41ee3c6d7b9c7ab8b`
- `tools/a152-qualify-volume-profile-evidence/README.md`: `3679037629821408b6838694f6892e7b45d8591a76b2b09c37417b392de6c66d`
- `tools/a152-qualify-volume-profile-evidence/testdata/synthetic.json`: `1f9c0086e6df75681be2d1e410f9a9e90dd443f0886310a28ec1997ed2e2ccf3`
- `openspec/changes/a152-qualify-volume-profile-evidence/design.md`: `9c33cca33d8b0323eb71fbf11129a80a22ba9f650bf5ccdb3caea0dd251497f3`
- `openspec/changes/a152-qualify-volume-profile-evidence/specs/candidate-discovery/spec.md`: `3b164f0fa793fe3aa670148a742cf18a7ba960f844b74550fcae3ae140f4e896`
- `/tmp/tossos-a152-offset-red.log`: `9224d34e23545c466572a03ac0fb84c471ccea0f9746f0ab0917665dfe2fed25`
- `/tmp/tossos-a152-offset-green.log`: `5d78e41521f16b3128b98c5056e10e3146c96f0d12510a84e38411de6f19696d`
- `/tmp/tossos-a152-offset-cli.json`: `5c1fc88ebafce240872fbc7f0451d21ca717e8e9ed527022e2d90e37113fd7e2`
