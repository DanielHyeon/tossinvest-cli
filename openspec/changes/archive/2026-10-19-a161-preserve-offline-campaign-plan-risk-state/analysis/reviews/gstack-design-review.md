# a161 gstack design review

Current disposition: **PASS — scoped design only**. No unresolved P1/P2 design findings. Implementation, tests, gate and lifecycle are not verified by this receipt.

Reviewer: Codex-host gstack review. Applied installed `/home/daniel/.agents/skills/gstack/review/SKILL.md` and two-pass checklist to the proposed boundary, with native source inspection. No Claude or cross-model execution claimed. No test execution, production edits, telemetry, configuration changes or live operations performed.

## Reviewed conclusions

- Minimal dependency direction is sound: lane-owned one-way export can access private fields; a stdlib-only detached codec cannot import native execution types. No native restore/mint API, runtime caller, journal migration or authority expansion is authorized. Existing a160 unavailable semantics stay intact.
- Current risk means caller-provided capture-time values. Required inspection-only, source/currentness disclaimers distinguish checksum validation from authentication, journal completeness and real-time risk. Caller synchronization is explicit; no concurrent-map safety or atomic journal capture is claimed.
- Source checks confirm continuation public fill records can legitimately have empty RiskMinor and false flags, reversal has a different record shape, and weekly Known tracks fill identity rather than verified risk. Weekly cancellation history belongs elsewhere. Design explicitly preserves incomplete records, false map entries, unknown latches and overbudget values instead of resetting or clamping them.
- Complete native plan preimages, private FX/seal data and original digest algorithms remain evidence; artifact digest is separate. Allocation ceilings and initial budget are not inferred from journal digests. Source inspected: continuation allocation/risk, reversal plan/risk/types, weekly plan/risk.
- Bounded canonical decoding, external digest and complete scope binding, rejection without partial result, and revalidation of mutable DTOs form an adequate inspection trust boundary. No SQL, network, shell or broker input path is introduced.
- Caller-owned temporary-file close/reopen is sufficient for stated persistence scope. A new storage service, lock hierarchy or fsync protocol would add unrequested complexity; no power-loss durability is promised.
- Proposed negative/roundtrip/deep-copy tests cover six family/market combinations, uncertainty and source immutability. Behavioral mutation/fault evidence and separate focused/race receipts remain implementation obligations.

## Implementation review checkpoints (not additional design blockers)

Canonical re-encoding must reject omitted required zero/false fields and duplicate keys without rejecting legitimate empty unknown-risk record values. Avoid omitempty-based loss of required presence; preserve nil/empty representation where native fidelity requires it. Bound bytes before untrusted decoding, reject unsupported enum/family combinations and compare scope exactly. Test tampering with externally retained digest, a valid artifact paired with foreign expected scope, and no usable result on every error. Native structural validation must not remint authority or mistake stale FX evidence for a fresh quote.

## Reviewed design snapshot

| File | SHA-256 |
|---|---|
| `proposal.md` | `06920ddf31d620ad7f85dc9d67425661289936a7f5464046ba36ded3c5be04fe` |
| `design.md` | `994f36ed006002c0be5eae7a8039dec049080a89aa05460c6961cfe9a94992cb` |
| `tasks.md` | `bbec5e3ba0e87314151644d1eb295fd75a2341722f7d46d974a20a5f1f970365` |
| `specs/campaign-state-evidence/spec.md` | `d1b665ec8f0dc4fe67cc530dbc6dc7676b48a5630aa267257843d68841ee6037` |
