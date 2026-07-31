## 1. Evidence and contract freeze

- [x] 1.1 Capture the implementation base, CodeGraph/CodeGraphContext reconciliation, Function Logic Maps, Branch Test Maps, risk reports, and the High-risk Pre-Edit declaration.
- [x] 1.2 Strict-validate the Story-linked OpenSpec artifacts and record proposal-freeze review with an adversarial engineering voice.

## 2. RED regression contracts

- [x] 2.1 Add failing command/integration tests with two databases proving an explicit config directory selects only its journal, the default profile still selects the data-directory journal, missing/incompatible selected journals never fall back, and a v8 selected journal is classified before the `policy_id` query.
- [x] 2.2 Add failing console tests proving the adoption control has no inline script, the remote response keeps the existing CSP, deterministic rendering covers non-default and legacy off-step values, every legal 2%..20% half-step converts correctly, adjacent invalid values plus empty/NaN/infinities/stale fraction submissions write nothing, and an isolated real config seam round-trips `7.5 → 0.075`.

## 3. Minimal implementation

- [x] 3.1 Resolve the console journal with the engine's active-profile rule and extend the read-only compatibility check for the required v9 column without migration, fallback, or write authority.
- [x] 3.2 Replace the script-dependent range/output pair with a native percentage control and add finite percentage parsing, half-step validation before division, deterministic formatting, and fraction conversion.

## 4. Verification and delivery

- [x] 4.1 Run focused tests, race tests for affected packages, full tests, vet, strict validation, logic-map checks, and security/static guards without starting an engine, soak, toggle, or LIVE order.
- [ ] 4.2 Complete independent implementation review, refresh SDD indexes, and pass `make sdd-check` plus `make gate CHANGE=fix-adoption-console-truth`.
- [ ] 4.3 Commit and push the feature branch and build the Docker image. Record production recreation as a separate human-authorized operation because it stops and may autostart the engine; do not POST the live setting.
- [ ] 4.4 Sync the operator-console spec, archive the change, update Story/PM state, and retain a secret-free episodic learning.
