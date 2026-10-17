# a142 additional gstack review

Date: 2026-10-16. Reviewer `/root/gstack_compatibility_reviewer`, independent Codex context, separate from Terra implementation and dedicated a142 adversarial reviewer. Applied `/home/daniel/.agents/skills/gstack/review/SKILL.md`, critical/informational checklist, specialist and red-team perspectives. Baseline `b36798f0ec47ee9b0d0d26d5492d0eafe8dd5172`; production `c8ab315a`; final tests `3aeabc4c`.

## Verdict

**PASS REVIEW for amended a142 scope; both P2 findings CLOSED.** No demonstrated unresolved defect in the reviewed adapters and narrow watchlist probe amendment. No tests or product files executed/edited by this reviewer. Full gate and completion remain separate. Closure evidence follows below; original findings are retained as history.

## Findings

1. **P2, confidence 10/10 — final evidence descriptions disagree with current AST.** In `analysis/function-logic/`, `internal-client--client.listwatchlist/function-logic-map.md` calls an if/range function a branchless wrapper and prints old source hash `9da13...` while current AST/source is `fc544...`. `internal-client--client.listwatchlistgroups/function-logic-map.md` prints `e9c350...` versus current `de925...`. `internal-client--testlistwatchlistfromfixtures/function-logic-map.md` invents B4 although AST has only B1–B3. The removed `internal-client--testlistwatchlistsendstypesfilter/branch-test-map.md` names replacement TestListWatchlistRejectsInvalidCollections despite its AST representing removed SendsTypesFilter. Generic existing-coverage GREEN yes and per-branch RED yes must not stand for unnamed/unmeasured branch execution; distinguish focused-test success, retrospective baseline replay, historical unretained RED and static unchanged-path evidence. Pending B2 links consumer `{}` cases that actually reach nil guard B3. Correct documents only; AST itself matches current source where designated final.

2. **P2, confidence 10/10 — watchlist monitoring follows the retired read contract.** `internal/ops/wts_operations.go:579–600` still declares POST `/api/v2/dashboard/asset/sections/all` with WATCHLIST body/check, while its handler at line 605 now reaches the dedicated bulk GET through changed ListWatchlist. `internal/monitor/probes.go:61` actually imports these registry probes. This can miss a failure of the used bulk endpoint or alert on an unused dashboard contract. This source is outside the approved four-file adapter scope; sent to Manager for an explicit narrow follow-up or authorized scope extension, not silently changed. The groups probe's different query is pre-existing and not a new standalone blocker asserted here.

## Verified scope and checks

- Full production/test diff and callers read. Pending/completed collection guards distinguish nil from explicit empty arrays and preserve decoder errors. Completed aggregation returns nil/error on the second-market failure; local tests exercise a populated first market. HasPendingOrder/FindOrder propagation tests prevent an unknown collection being treated as authoritative absence.
- Dedicated watchlist uses existing bulk GET with includePrice=true/lazyLoad=false, validates required watchlists/items arrays, stably sorts numeric ordering and retains code identity, group/name, currency and optional price defaults. Current tests cover group/item unequal ordering plus ties and positive/zero/null/omitted price states.
- No changed mutation method body, operation ID, engine, token-sharing, journal, risk-control or operating configuration. Existing mutation-test receipts are narrow preservation evidence, not exhaustive mutation validation.
- No new SQL, concurrent persistent state, shell execution, LLM execution boundary, migration or UI. New decoder types are internal and only used by this bulk read. Slice sorting operates on the local decoded response; no shared mutable state was found.
- Fresh independent `/root/gstack_compatibility_reviewer/a142_specialist` read testing/security/API-contract/maintainability/performance/red-team checklists and returned NO FINDINGS for scoped adapter changes. It independently noticed the ancillary monitor mismatch. Perspectives were combined in one context due concurrency limits; not represented as six separate reviewers or cross-model consensus.
- Read focused verification receipt and dedicated reviewer test receipts. Retrospective replay is explicitly distinguished from historical TDD RED. Reviewer did not duplicate tests/full suites or contact live services.

## Execution limits

Local scoped review; no PR/remote fetch, Greptile, release queue, telemetry, setup, external sync, persistent memory writes, Claude CLI or separate Codex CLI review. Actual participants are Codex agent contexts. The gstack checklist/specialist/adversarial workflow is adapted to user-mandated Manager/Terra separation. No claim of literal Claude-plus-Codex CLI pipeline, live WTS compatibility, deployment or full gate PASS.

## Reviewed SHA-256

```text
internal/client/account.go cda0d5d3cb12e91e55641aa66cc54f02f7e77ed48b57a4052b83c09d39b0a90f
internal/client/completed_orders.go 2c389d043a94552ec5dd4e0179f3f5c3cdbd1eb26eaefec8cbbfcae5959cc486
internal/client/watchlist.go fc5440b66ddc096a9daeb2a8a125cf132d69bb69b7c191aaab864a44eb74f33b
internal/client/watchlist_manage.go de925d43fbf32a246efaf6c6df56fbf6a2f7158d00188818dcbc6dbb39dca4f1
internal/client/orders_parse_test.go 68c536332b40ea4d407232486ec7adf43f3dc902b3ec84e4a36d002b120fc182
internal/client/completed_orders_test.go edbc76c6f47787af02a153444a21ef7cc8b84ceff1c7835419fb1e2d8f67bf1a
internal/client/watchlist_test.go de3ec778a43b365513e8d9ae4070f3bd19a039c6ac5c5bf397d46af9288e7bf7
```

## Final closure re-review at 10be99cf

**P2-1 CLOSED:** Re-read Manager's final client FLM/BTM corrections. Watchlist wrapper now enumerates B1 error and B2 range; both watchlist source hashes match current AST. Fixture map has only actual B1–B3, with full-field equality in B3. Removed types-filter test is explicitly historical, not the replacement test. Pending consumer failures are mapped to B3. Named-test success, retrospective baseline failure and static unchanged paths are distinguished without claiming isolated branch execution. The new helper's evidence/coverage scope is documented. Current AST IDs and final descriptions align.

**P2-2 CLOSED:** Manager approved the OpenSpec amendment before implementation. `10be99cf0c857a40dfb44f270030b9ed718a2786` changes only the existing watchlist ProbeSpec in the registry: GET exact bulk URL/query, empty body, existing status requirement, and missing/null/wrong-type checks for watchlists and each items array. Empty arrays remain accepted. Operation ID/handler, other probes, scheduling/configuration and monitor-runner policy are unchanged. `TestWatchlistProbeDedicatedReadContract` exercises the registered callback locally, including invalid JSON/status and malformed later groups. Its populated valid-item path is not separately asserted; the intentionally opaque item payload and populated adapter have separate coverage, so this is not a demonstrated defect.

Reviewed current registry AST hash and B61–B65 (status:582, decode:592, collection:595, range:598, items:599), matching final FLM/BTM. Read independent reviewer's local focused PASS receipt; did not duplicate its tests or run a live probe. Final Go scope is five production files and four test files; the original seven client hashes remain unchanged. New final SHA-256:

```text
internal/ops/wts_operations.go e622a677b56a55caaf4eec189086219112c9b48ba3d4a49e22f690d00cdf1fd2
internal/ops/wts_operations_watchlist_probe_test.go a5e6a64d65490c0f395e80a8a5ac16c14be04e0c950554ba412372232109a0dd
```

The independent additional specialist also re-read the narrow 10be99cf amendment and returned NO FINDINGS, without writes/test execution/network calls.

No remaining finding from this additional gstack review. Manager's mandatory full-suite/gate verification and archive/integration remain separate completion requirements; this verdict does not imply those checks passed.
