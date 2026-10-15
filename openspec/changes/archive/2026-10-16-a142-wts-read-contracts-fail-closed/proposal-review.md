# a142 independent proposal review

Date: 2026-10-15. Reviewer: `/root/a142_adversarial_reviewer`, inherited manager model; one independent agent, not four independent models. Initial status: CHANGES_REQUIRED. Revised-artifact status: **APPROVED FOR PROPOSAL FREEZE**; implementation READY still requires baseline/evidence prerequisites.

Method: read repository safety/SDD instructions, gstack `autoplan/SKILL.md`, proposal/design/delta/tasks, Terra intake and its pre-edit AST, current domain/fixtures, upstream source/test fixtures at `75485727fd0fda6a62378624feac0b3f488167f3`. Four perspectives below are explicit review lenses of this reviewer. This record does not claim the full dual-voice autoplan pipeline or the later implementation gstack review ran. No production code/tests/settings, live endpoint, credentials, global gstack configuration or external publication changed. Manager owns spec dispositions.

## Findings requiring spec disposition

| ID | Severity | Finding | Required disposition |
|---|---|---|---|
| P1 | High | “Preserve identity” does not decide whether `Symbol` remains product code or becomes ticker. Upstream introduced `ProductCode` separately; local `WatchlistItem` has only `Symbol`. Copying ticker preference changes the local JSON meaning. | Preserve local `Symbol = item.code` for both existing list interfaces. Do not add domain fields or silently prefer `item.symbol`. Test fixture with different code/ticker values. |
| P2 | Medium | “Server-defined ordering” could mean wire order or explicit `ordering` numbers. | Define ascending numeric group ordering then ascending item ordering, stable ties. Omitted ordering remains Go zero default; test shuffled input and equal keys. |
| P3 | Medium | “Required collections” is unspecified while upstream has both bulk and metadata-only APIs. | For the existing bulk route require `result.watchlists` and each group's `items` arrays. Absent/null/wrong-type fails, explicit empty arrays succeeds. No adoption of `/groups/simple` or per-group lazy APIs. |
| P4 | Medium | Optional currency/price defaults are left to implementation. Fixtures establish USD/KRW but not arbitrary missing-value financial meaning. | Copy currency verbatim, retain empty currency when absent; preserve existing nullable `base`/`close` conversion to zero without currency inference or price substitution. Test omitted/null prices and legitimate numeric zero separately. Zero is presentation compatibility, not a new trading-authority assertion. |

Missing/invalid item identity is worth a follow-up, but do not import unrelated upstream positive-folder-ID, required-itemCount/count-mismatch or mutation-reconciliation contracts under this narrow collection fix. If Manager adds such checks now, enumerate them explicitly and re-review the affected requirement.

## Upstream and local fixture evidence

- Upstream `fixtures/responses/auth-sanitized/new-watchlists.json`: groups have ordering -100/0/100; items have -2/-1. NVDA item has code `US19990122001`, symbol `NVDA`, USD base 120.5/close 125.3. Samsung has code `A005930`, symbol `005930`, KRW base 72000/close 73500. Third item has empty symbol but nonempty `FALLBACK_CODE`.
- Upstream `internal/client/watchlist_test.go`: symbol mapping/fallback, currency, ordering, explicit-empty nonnil result, omitted items/null items and missing collection tests. These prove intended upstream adapter behavior; they do not establish live WTS guarantees or authorize adopting the entire upstream domain/mutation implementation.
- Local `fixtures/responses/auth-sanitized/asset-sections-v2.json`: WATCHLIST `stockCode` examples are product codes such as `NAS20200101036`; price currency is explicit. Local `internal/domain/models.go:82` has `Symbol`, no `ProductCode`; its `WatchlistGroup` has no Ordering output field. Ordering can stay an internal wire field.
- Existing AST source hashes and enumeration: `/tmp/tossos-a142-intake/ast/list-pending-orders.json`, `list-completed-orders-range.json`, `list-watchlist.json`. This review uses that intake for existing branch claims. New shared-decoder internals must get their own AST/maps before edits; no ungenerated shared-function branch claim is treated as proven here.

## CEO perspective

Premise challenge: Does a read-contract bug justify replacing the client or adding new APIs? No. User value is avoiding false absence and restoring a stable watchlist read, not increasing trading authority. Existing read session, GET helper and dedicated endpoint already exist, so use their narrow contract and preserve engine/journal/Guardian/token behavior. Dream-state delta is reliable distinction between empty and malformed responses; websocket/all-dates/group-management redesign is outside scope. Scope accepted after P1–P4 disposition.

## Design perspective

No browser/UI layout scope; visual scoring is not applicable. Output contract still matters: user-facing JSON identity must remain stable while internal endpoint changes. Empty results must serialize as existing successful arrays rather than error-like null outputs. No new configuration knob, operator toggle or migration is justified. Main design failure risk is P1's ticker substitution.

## Adversarial engineering perspective

Minimal nil guards at the two order collection boundaries fit the root cause. They must execute for every required market response, and an error in a later market must return an error rather than accumulated earlier orders. Wrong-type JSON errors already carry decode failure; do not replace them with empty slices. Shared watchlist read mapping is acceptable only with read-only ownership and no mutation payload edits. Do not change retry, absence, terminal-status or reconciliation policy.

```text
CLI / WTS operation consumers
  -> pending GET -> required result[] -> existing order parser -> result or error
  -> completed market GETs -> required result.body[] per market -> aggregate or error
  -> ListWatchlist -> existing bulk GET -> required watchlists/items -> stable flatten
                                                        |
                                       same local domain and mutation methods
```

| Failure mode | Rescue / required result | Evidence required before acceptance |
|---|---|---|
| Omitted/null/wrong-type pending collection | Descriptive schema/decode error, no authoritative empty result | Table tests including explicit [] success |
| KR valid then US malformed or HTTP failure | Error, no successful partial aggregate | Local server asserts both requests and failed aggregate |
| Empty both markets | Successful nonnil zero-length list | Exact value assertion |
| Watchlist envelope/items absent/null/wrong type | Error | Bulk route table tests, including second-group failure |
| Reordered groups/items/equal ordering | Stable expected flat sequence | Shuffled multi-group fixture |
| Code differs from ticker | Product-code Symbol preserved | Exact JSON/domain assertion |
| Optional currency/prices omitted/null | Existing empty/zero defaults | Exact field assertions; no inferred currency |
| Shared decoder affects mutations | Existing request contracts unchanged | Existing watchlist mutation suite |
| Network or session error | Existing error propagation | Existing affected tests, no account traffic |

Test plan: Terra first produces RED fixtures against baseline, then minimal implementation, focused client tests, affected CLI/WTS consumer tests, mutation regressions, final AST/maps, independent implementation review, separate gstack review and mandatory gate. Reviewer will inspect receipts rather than infer test success from checkboxes. No tests were executed by this proposal reviewer.

## DX perspective

Developer journey: existing Go checkout -> same dependencies -> local fixture client -> exercise existing list method -> inspect descriptive collection error or familiar domain result -> run focused tests -> inspect AST/map receipts -> independent review -> mandatory gate. No new install/run steps, flags, authentication flow or runtime setup. Time-to-first-fixture remains the existing Go test workflow; no measured duration claim. Error text should name `result` or `result.body`/`result.watchlists`/`items` so the failed boundary is discoverable without printing response/account data. Documentation must state that stricter errors are intentional schema-drift detection. Ticker migration would undermine safe upgrades, reinforcing P1.

## Decision audit and completion

| Decision | Classification | Reason | Rejected alternative |
|---|---|---|---|
| Keep product-code Symbol | Compatibility requirement | Preserve existing local JSON semantics | Copy upstream ticker preference without its domain contract |
| Stable numeric ordering | Mechanical clarification | Explicit upstream order fields and tests | Depend on incidental array order |
| Bulk collections required | Mechanical clarification | Existing includePrice=true/lazyLoad=false route returns full items | Adopt metadata-only omissions |
| Preserve zero/empty optional defaults | Compatibility requirement | Existing domain cannot represent price unknown separately | Invent currency or price fallback |
| Skip visual review and extra model claims | Not applicable / transparency | No UI; only this independent agent performed these lenses | Fabricated multi-model consensus |

Cross-perspective concern: identity compatibility affects product, design, engineering and DX. No challenge to the user's stated scope. Required subsequent stages: a141 landing/baseline evidence, Terra implementation, this reviewer's separate adversarial implementation pass, additional gstack review, Manager verification. This document is not implementation completion evidence.

## Revised-artifact re-review

Manager revised design decisions 2–3, resolved Open Questions, the watchlist requirement, and added the optional-values/sorting-ties scenario. Independently re-read those on 2026-10-15: P1–P4 are all accepted and resolved. Bulk-only collection checks, local code identity, explicit currency/price defaults and stable ordering are now unambiguous; no unresolved proposal finding. Approve the bounded proposal freeze. This does not waive baseline commit, fresh shared-decoder AST/maps, offline RED/GREEN receipts, later independent adversarial review, additional gstack review or completion gate. Full multi-model autoplan consensus remains unclaimed; this artifact supplies the explicitly requested independent four-perspective proposal review.

## Narrow probe amendment review (2026-10-16)

Reviewed amended proposal, design decision 5, task 2.4 and health-probe scenario. **GO for bounded implementation after pre-edit evidence/RED prerequisites.** Product perspective: probe/read endpoint disagreement undermines the compatibility repair. Design perspective: no new user control, output format or operational setting. Adversarial engineering perspective: one existing watchlist ProbeSpec only; exact GET/query/no-body plus required bulk/per-group arrays and non-success status must be tested offline. DX perspective: retain operation identity/handler and other probes to avoid unrelated contract drift. No unresolved scope question or further user authorization needed within these stated boundaries.

Implementation verification must assert the exact request definition, valid populated/explicit-empty responses, missing/null/wrong-type watchlists and items (including malformed later group), invalid JSON and non-success status. Excluded: schedules/configuration/live probe invocation, other probe definitions including watchlist-groups, operation IDs/handlers and shared monitor-runner policy. Subsequent independent code re-review and additional gstack re-review remain mandatory; earlier PASS CODE does not cover this new implementation.
