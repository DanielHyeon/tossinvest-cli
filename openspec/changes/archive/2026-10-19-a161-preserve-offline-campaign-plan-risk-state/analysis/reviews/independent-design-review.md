# a161 independent design review

Verdict: PASS — design scope only. No implementation or test result is asserted.

Reviewed baseline: 67810038f21dffd519a3d44edb9e255238b5b74a. Reviewed proposal, design, delta spec, tasks and review against current continuationlane, reversallane and weeklyvaluelane plan/risk source. No production changes, tests, live calls or existing evidence edits performed.

## Assessment

The detached inspection-only return contract resolves the principal authority risk: export may inspect existing native validation, but decode may not mint native plan, FX or RiskState. Caller-provided provenance and unverified ledger currentness are mandatory data, not only documentation. Initial budget and recorded filled/held remain separate; no NewRiskState reset, inferred legacy state or a160 availability upgrade is permitted. Weekly ReservationState/cancel history exclusion is explicit and machine-readable. External expected scope and content hash are integrity constraints, not authenticity proof. Caller-owned temporary-file roundtrip is appropriately bounded and does not claim crash durability.

Existing digest compatibility is preserved by retaining native digests separately from the envelope hash. This matters because continuation allocation.go planDigest uses length prefixes, reversal plan.go planDigest uses NUL-delimited original FX fields, and weekly plan.go planDigest includes the frozen FX seal with its own time canonicalization. No algorithm consolidation is authorized.

No blocking contradiction found. Implementation review must verify these already-specified boundaries:

- Required top-level filled/held and budget strings must be present and valid; native incomplete per-fill records must remain representable. For example reversallane/risk.go ApplyFillRisk records Applied=true with empty RiskMinor for unknown risk. A blanket nonempty check on every nested RiskMinor would violate the unknown-preservation scenario.
- weeklyvaluelane/risk.go AppliedFill.Known is initialized from knownIdentity before later time/FX/arithmetic failures. Preserve this flag without deriving an actual-risk-known claim from it. continuation/reversal public mutable risk states cannot establish ledger completeness; weekly validRiskState seal cannot establish it either.
- Preserve all family-owned fields, false latch entries and conflict/missing-ID keys, including continuation cancel Applied/Fingerprint and reversal cancel string values. Weekly absent cancel data must not become an invented empty cancellation history.
- Capture must detach maps/FX after completion; caller synchronization during capture is an explicit precondition. No concurrency guarantee should be inferred from race tests using immutable inputs.
- Strict canonical roundtrip must reject omitted required fields, duplicate keys and trailing input rather than silently canonicalizing them; malformed input must return no usable partial record. Mutable DTO re-encoding needs the same checks.

## Reviewed hashes

proposal.md 06920ddf31d620ad7f85dc9d67425661289936a7f5464046ba36ded3c5be04fe

design.md 994f36ed006002c0be5eae7a8039dec049080a89aa05460c6961cfe9a94992cb

specs/campaign-state-evidence/spec.md d1b665ec8f0dc4fe67cc530dbc6dc7676b48a5630aa267257843d68841ee6037

tasks.md bbec5e3ba0e87314151644d1eb295fd75a2341722f7d46d974a20a5f1f970365

review.md 2904c0405c80bd784de0f61268ce0a003ac96bf0f3af4e60d117bf7c090bda43
