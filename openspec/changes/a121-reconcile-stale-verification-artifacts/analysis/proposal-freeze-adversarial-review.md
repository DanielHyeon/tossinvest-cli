# a121 proposal-freeze adversarial review

**Verdict: BLOCK — do not begin RED or implementation until the reconciliation proof is made decidable and safe.**

Read-only review of a121 proposal/design/tasks/delta specification/base/hard evidence, the a063 cleanup-404 diagnosis, and current record, cleanup, verification, and official-reader source. No live CLI, account, order, engine, systemd, survey, or configuration command was invoked.

## What the proposal gets right

The proposal correctly keeps a063 operational work separate, preserves the failed DELETE, rejects 404 and generic operator observation as terminal proof, prohibits reuse of cleanup/resume/abort/redo, preserves verdict/attestation boundaries, and requires an append-only, idempotent result. The hard-evidence inventory correctly identifies Outstanding/cleanup as shared projections and marks existing edited functions for maps.

## Blocking contract gaps

1. **The absence proof is not safe against a lifecycle race.** ConditionalOrdersRaw exposes only independently fetched OPEN and CLOSED pages with cursors; it carries no server snapshot revision, read-cut timestamp, or cross-group consistency token. A conditional can be replaced or trigger between pages. In particular, disappearance of the parent conditional is compatible with a triggered child order or successor artifact. An OPEN+CLOSED scan that does not find the old opaque ID therefore cannot by itself safely permit the record to stop protecting that artifact.

   The delta requirement must define an authoritative, consistent proof: a broker-supported snapshot/cut token if available, or a fail-closed alternative that includes the necessary official child/successor/order-state evidence and rejects any response set whose time/identity consistency cannot be established. “Fresh” also needs an explicit bounded measurement (start/end clock, maximum duration, and refusal on expiry); current raw-page values contain no freshness or snapshot field. A partial page, cursor failure, or state transition during the scan must leave the artifact outstanding.

2. **The append-only event cannot currently be both exact and free of raw broker IDs.** The design requires exact record-owned artifact matching while the delta spec prohibits retaining raw account or broker identifiers, but neither document defines the event’s identity encoding, version, collision/domain separation, or projection lookup rule. A plain event cannot safely make one selected artifact disappear; a raw ID violates the stated boundary.

   The design/spec must define a versioned, domain-separated digest of the exact record-owned tuple at minimum (artifact kind, opaque ID, selected profile/record binding, account binding, market and symbol as required), its storage, comparison, duplicate rule, and redacted display. It must specify the selection rule too: reject zero or multiple eligible outstanding conditionals unless a non-arbitrary record-owned selector resolves one exact artifact. Never accept a caller-supplied raw identifier or broad account scan.

3. **Account/profile binding needs an executable equality rule.** The current record holds a masked account reference and the current verification resolver returns the selected official account reference/sequence. The proposal says “chosen profile/account context” but does not require all reconciliation candidates to have a single matching record account/profile/market or specify how a masked record reference is compared with the current selected account. Define those equality checks and reject missing, mixed, or mismatch records before any official read or local append.

These are safety requirements, not implementation details. Without them, a reconciled-absent event could hide a trigger/replacement state, remove a cleanup barrier for the wrong artifact, or turn a non-atomic paginated observation into terminal evidence.

## Required freeze amendments

- Define the consistent official-read proof and bounded freshness semantics, including trigger/replacement/child-order ambiguity. If the official API cannot provide the necessary consistency, the command must refuse reconciliation rather than treating absence as sufficient.
- Define the non-raw, versioned artifact identity/basis representation and exact-selector behavior.
- Make account/profile/market equality and single-candidate ambiguity refusal explicit in the delta specification and RED tasks.
- Add RED cases for: parent disappears while a successor exists; parent triggers a child; status transition during OPEN/CLOSED pagination; read duration/snapshot failure; multiple candidate artifacts; digest/domain mismatch; and mixed/masked-account or profile mismatch.
- Add a structural no-mutation test that the reconciliation dependency exposes only official GET reads and a local append, plus an assertion that no event affects SucceededEndpoints, BuildReport measured attributes, soak proof, or engine interlock.

## Maps and gates

Task 1.2 must name every selected existing function before RED and produce its AST, Function Logic Map, Branch Test Map, and risk report. At minimum this includes the actual record projection function(s), newVerifyCmd if registering the new command, any edited record loader/account resolver, report/status projection, and any official-reader function that changes. The existing-function table is useful, but “functions that will be edited” is not sufficient until the implementation boundary is fixed.

Current a121 hard evidence records both make sdd-check failure (stale/missing CodeGraph hard evidence) and PM tracker check failure. Under the a121 document’s own stated freeze boundary, those failures currently do not permit implementation. Independently, make sdd-check is a mandatory SDD completion gate; the PM check is normally a handoff/archive concern, but it remains an explicit unresolved a121 freeze condition until the Manager records an authorized resolution. Neither failure may be papered over or bypassed.

a063 remains unarchived and operationally pending. This proposal review authorizes neither a121 implementation nor any reconciliation operation.
