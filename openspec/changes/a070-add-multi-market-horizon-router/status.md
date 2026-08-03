# Status — a070-add-multi-market-horizon-router

- Core package: GREEN (`internal/strategyrouter`)
- KR/US delivery: paired release descriptors, both default `OFF` / `UNOBSERVED`
- Routing: exact owner generation plus exact sealed market record/revision; no mutation authority
- Scheduler: independent KR/US records, CAS/locks, OFF-only rollback, fail-closed idempotent migration
- Quota: one physical endpoint/reset-generation counter with market/horizon anti-replay subscopes
- Authority boundary: official owner/quota/ON-record attestation constructors are package-private
- Remaining: a066–a069 integration, safety-loop regression, `make sdd-sync`, `make sdd-check`, root gate and independent final review

No market was selected, no lane was enabled, and no live order/toggle/activation was created.
