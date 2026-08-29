## 1. Evidence and proposal freeze

- [x] 1.1 Record StockOS policy-number provenance, CodeGraph impact evidence, and pre-edit Function Logic/Branch Test Maps for every existing high-risk function to be changed.
- [x] 1.2 Validate the OpenSpec delta strictly and record the required adversarial engineering proposal-freeze review in `review.md`.

## 2. RED contracts

- [x] 2.1 Add failing registry/evaluator tests for all three policy profiles, HYBRID_50 half-retention, T4 high-water trailing, monotone floor, and breach precedence.
- [x] 2.2 Add failing config tests for empty legacy behavior, registered-policy validation, surgical byte preservation, and unknown-policy refusal.
- [x] 2.3 Add failing v8→v9 migration/journal tests for nullable policy snapshots, legacy LADDER interpretation, and adoption crash recovery.
- [x] 2.4 Add failing engine tests for self-opened and external-adopted policy snapshots, mixed active policies, unknown snapshot refusal, and no automatic rebind.
- [x] 2.5 Add failing console/cmd tests for `/optimization`, session+CSRF POST gating, minimum-capability seam, audit before/after, and no gate/trading mutation.

## 3. Core implementation

- [x] 3.1 Implement immutable BALANCED/RUNNER/HYBRID_50 registry and HYBRID_50 runner protection in `internal/exitpolicy`.
- [x] 3.2 Add config `engine.exit_policy.common_policy`, fail-closed merge validation, and locked surgical load/save operations.
- [x] 3.3 Add journal schema v9 and policy ID fields to exit/adoption records without rewriting historical rows.
- [x] 3.4 Resolve and persist the common policy when self-opened and external-adopted positions begin exit management, including crash recovery from adoption.
- [x] 3.5 Evaluate each active LADDER position with its stored registry policy while preserving existing proposal, cancel-first, Guardian, and submission paths.

## 4. Optimization console

- [x] 4.1 Add the `/optimization` GET page with current/recommended policy, exact rung descriptions, external-adoption behavior, and restart/effective-scope notices.
- [x] 4.2 Add the session+CSRF protected policy POST, dedicated config seam, cmd wiring, and save-time audit.
- [x] 4.3 Extend static route/capability tests so optimization cannot gain broker, gate, trading-toggle, or journal-write authority.

## 5. Verification and handoff

- [x] 5.1 Refresh Function Logic Map AST/risk evidence and mark every RED/GREEN branch result.
- [x] 5.2 Run focused tests, race tests for affected packages, full `make test`, `make vet`, strict OpenSpec validation, and `make validate`.
- [x] 5.3 Run independent adversarial review, resolve findings, then run `make sdd-sync`, `make sdd-check`, and `make gate CHANGE=add-common-exit-optimization`.
- [x] 5.4 Record verification evidence and episodic memory without changing LIVE gate, starting the engine, or submitting an account mutation.
