## Proposal Freeze Review

### Reviewers

- Independent test/contract reviewer: `/root/http2_hotfix-test-review`
- Independent security reviewer: `/root/http2_hotfix_security_review`

### Findings and disposition

1. **P0 — unsafe reconcile resolution surface.** `Tracker.Resolve` does not bind an authoritative stable re-query and releases rows in separate transactions before clearing tracker/gate. `Tracker.Blocks()` is not the complete journal reconcile set. **Resolved:** remove resolution capability, POST routes and all journal mutation from a052. A future change must design stable re-query evidence plus atomic journal/tracker/gate transition.
2. **P0 — wrong projection priority.** `exclude > eligible` would mislabel existing managed positions; `block > candidate` would mislabel unselected positions. **Resolved:** candidate-first predicate and `unknown > managed > excluded > candidate+covering-block > candidate > unmanaged` table is normative.
3. **P0 — raw exit evidence conflict.** Hiding every raw exit-state value conflicts with the operator-console visibility contract, while promoting raw values to effective would violate canonical snapshot truthfulness. **Resolved:** raw t0/initial-stop/baseline/high-water is shown only as `stored evidence · effective unknown`; actionable fields continue to require the canonical persisted effective snapshot.
4. **P1 — US normative regression missing.** **Resolved:** add an exit-policy delta with explicit AAPL/US include-only and blocked scenarios.
5. **P1 — API schema underspecified.** **Resolved:** fix six stable enums, labels/reason/booleans/nullable block and effective-known contract.
6. **P1 — symbol-only price mapping risk.** Current adoption price fan-out is symbol-only and does not prove market/currency binding. **Resolved in a052:** accept only KR/KRW or US/USD provenance; empty/mismatch currency and cross-market duplicate symbol candidates fail closed with no adoption record.
7. **P1 — tracker projection is not the full journal reconcile authority.** **Resolved by semantic narrowing:** adoption status intentionally uses the same `Tracker.Blocks()` projection as `ReconcileDriver.blocked()`, labels its source `adoption-blocking tracker projection`, and does not claim to enumerate every journal reconcile cause.
8. **P1 — runtime-unavailable booleans could masquerade as false.** **Resolved:** status/designation knownness and typed runtime-unavailable reason are explicit; desired remains visible only in the settings summary and cannot promote an effective row status.

### Freeze decision

**APPROVED AFTER NARROWING** by test-contract disposition and independent security re-review. The security re-review found no remaining P0/P1 proposal issue after tracker-semantic narrowing, runtime-knownness and US currency hardening. Reconcile resolution remains out of scope and no operating state may be changed during implementation or canary.

## Post-implementation requirement re-review

### Reviewers

- Independent implementation test/maintainability reviewer: `/root/a052_test_maint_review`, final re-review via `/root/a051_http2_body_debug`
- Independent implementation security reviewer: `/root/a052_security_review`

### Findings and disposition

1. **P1 — Compose namespace made loopback runtime unreadable.** The API sidecar cannot reach the console/engine container's `127.0.0.1` command endpoint. **Resolved:** keep lifecycle List/Preview/Apply on its existing loopback server and add a separate authenticated 0700/0600 Unix runtime endpoint in the shared engine directory. Its client exposes only `Runtime`; Preview/Apply/reconcile paths are absent.
2. **P1 — optimization returned contradictory legacy and nested values.** Static OFF/5% legacy fields could disagree with actual ON/3% nested fields. **Resolved:** legacy desired/effective summaries now derive from the same actual values; unavailable effective is explicitly `알 수 없음` and remains nullable/known=false.
3. **P1 — released lifecycle was projected as managed and could expose an old canonical exit line.** Immutable adoption provenance survives operator release. **Resolved:** both pages and the API read authoritative lifecycle, give durable `RELEASED` precedence as `UNMANAGED/OPERATOR_RELEASED`, and suppress old actionable snapshots to raw `실효 미확인` evidence.
4. **P1 — virtual lifecycle default looked like an operator release.** A version-0 row may default to `RELEASED` without a durable release command. **Resolved:** only `StatusReleased && Version>0` maps to `OPERATOR_RELEASED`; version 0 remains `UNMANAGED/NOT_SELECTED`.
5. **P2 — non-Unix engine startup regression.** A Unix-only runtime server initially made Windows startup fail. **Resolved:** non-Unix builds keep the engine available with runtime unknown; Windows cross-compilation passed.
6. **P2 — no-resolution invariant needed a direct public-route pin.** **Resolved:** public router tests now assert `/api/v1/reconcile` and `/api/v1/reconcile/resolve` are unreachable for every mutation verb.
7. **P2 — startup-only runtime client could remain stale across late engine startup or restart.** **Resolved:** API/runtime readers keep only the descriptor path and re-open its strictly validated descriptor on each read. A Unix integration test pins late-start recovery, fail-closed downtime and recovery through a restarted server with a replacement token and value.
8. **P2 — per-read transports could retain idle Unix connections.** **Resolved:** the runtime-only transport disables keep-alives, so the retryable descriptor strategy cannot accumulate idle connections under read bursts.
9. **P2 — canceled reads still performed descriptor filesystem validation.** **Resolved:** `DialRuntime` returns the request context error before touching the descriptor, with a direct cancellation regression test.
10. **P2 — late-start recovery comment exceeded the console lifecycle seam's guarantee.** **Resolved:** the comment now limits late-start recovery to the API sidecar and restart recovery to already-wired readers; lifecycle commands remain deliberately startup-bound on loopback.

### Final decisions

- Security: **PASS, P0=0/P1=0/P2=0** after focused race tests and static authority review.
- Test/maintainability: **PASS, P0=0/P1=0/P2=0** after released/virtual lifecycle, optimization parity, runtime transport and Windows portability re-review.
- Requirement change approval: the runtime-only Unix transport and durable-release priority are accepted because they close production-topology and truthfulness defects without widening trading or reconciliation authority.
