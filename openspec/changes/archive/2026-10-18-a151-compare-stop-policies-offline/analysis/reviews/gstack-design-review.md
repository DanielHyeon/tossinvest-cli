# a151 gstack design review

Verdict: DESIGN FREEZE PASS after rereading the corrected timing and entry-gap contracts. No unresolved actionable finding remains in reviewed design scope. Design-only Codex-host gstack review; no cross-model/Claude execution, implementation, tests, gates or runtime actions claimed. Read proposal, design, strategy-optimization delta, tasks and review. Only this report edited.

## Resolved findings and decisions

### Resolved P1 — Entry-bar receipt and hypothetical execution separated

Design chooses the first scheduled one-minute open strictly after signal confirmation, but also says “received_at보다 앞선 open으로 소급 진입하지 않는다.” The inputs describe complete OHLC bars with close_at and received_at. A normal complete entry bar arrives after its open; applying that condition to its receipt makes the declared next-open model unusable. Ignoring it silently would violate the contract.

Minimum recommendation: explicitly model an order scheduled after signal confirmation for the next minute open. Treat its entry price as a hypothetical simulation fill, not a contemporaneously observed quote; later complete-bar receipt is outcome evidence only. Require received_at >= close_at for completed bars and close_at/received_at <= confirmation for signal inputs. If actual executable receipt is required instead, add a separate open quote/event and its availability timestamp; never infer it from full-bar receipt. Choose one contract, not both.

### Resolved P2 — Signal features separated from entry-time sizing

Spec requires “Only information available at signal confirmation SHALL determine stops and quantity.” Design computes entry from a later minute open, then computes fixed/ATR absolute stops, target and quantity from that entry. Both statements cannot hold literally.

Recommendation: freeze signal low, ATR, eligible feature history, cost/tick and risk/cap snapshots at confirmation. At the declared simulated entry, use only its open and frozen slippage to derive absolute stop/target and quantity. Entry-bar high/low/close or subsequent receipts must not influence these values. Add a scenario with differing next opens but identical signal history, proving frozen features and explicitly entry-dependent geometry/sizing.

Resolution verified: design now explicitly models a pre-scheduled hypothetical next-open fill; subsequent complete OHLC receipt is outcome evidence, not retrospective live execution. Entry/outcome bars arrive after close. Signal inputs remain available by confirmation. Structural low, ATR and all manifests/budget/caps freeze at confirmation; only next scheduled open supplies entry-dependent absolute stops, target and quantity. The spec states the same separation and forbids entry-bar HLC from affecting those values.

Additional independent finding verified: the entry bar itself follows open-gap precedence. If adverse entry slippage places entry above a structural stop while raw open is already below it, immediate simulated exit uses raw open less exit slippage, not favorable stop price. Explicit scenario added.

The old review.md batch-level DESIGN FREEZE PASS link is historical. Its later clarification correctly says updated KR+US design awaits review; retain that distinction as historical context; current verdict is this corrected-contract design review.

## Contracts satisfactory after revision

- KR/KRW and US/USD units explicit; FX unsupported; no cross-currency account-return aggregation. Caller-supplied calendar/tick provenance is unverified, including US tick assumptions and calendar holidays.
- Same-session 15-bar ATR seed and Wilder continuation specified; closed/available signal inputs, five-to-one bar consistency and contiguous path policy provide implementable anti-leakage boundaries under the corrected receipt contract.
- Exact decimal limits, size/candidate/bar limits, duplicate/unknown JSON rejection, OHLC validation, positive bounded tick grid, fee rounding and zero-cap behavior are explicit. Local reads and exclusive new output/symlink refusal constrain side effects. Implementation still needs to honor these contracts, including rejecting invalid timestamp order and not following output symlinks during creation.
- Three-arm pre-eligible versus complete cohort, exclusion reasons, censoring, empty denominator and separate mark-to-market reporting resolve denominator ambiguity. Arm-specific sensitivity endpoints are not claimed as jointly achievable paired bounds.
- Same entry/target, fixed per-candidate budgets and caps, no compounding/portfolio simulation, and always-inconclusive profitability avoid unsupported performance conclusions. Synthetic arithmetic correctness is separate from empirical policy merit.
- EMA comparison explicitly unsupported and deferred. No new framework, database or live connectivity required; single manifest plus pure bounded calculator remains the smallest useful scope.
- grill-me/grilling decisions are recorded; user-selected US expansion is explicit. The two remaining timing decisions above are now explicit Manager choices; no unresolved design frontier remains in this review scope. No renewed approval of settled scope was requested. impeccable is correctly N/A because no UI work is proposed.

This PASS freezes the specified offline design only. It does not certify implementation, test execution, actual data provenance, profitability, or live operation. Implementation still requires fresh source mapping, fixtures, independent review and actual gate receipts.

## Reviewed document snapshot

SHA-256 at final rereview; later substantive edits require another review.

```text
380f9ca931c674484a064752e1dc6da72bea38cc442a2567c44975c796618468  proposal.md
80bdb4c959e04451548bad2e718eb65bfa5a895ea4d6fdebab22a1f9f06bd21d  design.md
2bd0ee278f988450e6a319315917942a9551d4716600b7bc5291b549e3fd4662  specs/strategy-optimization/spec.md
4ccf85422f211c9ef6edbfc9ef518f3a1525c6ad3b2dc2d4a6a83afffdd77d82  tasks.md
ee5e1deb93c8240e0a3313962f079273520e896fa7934c27f02f2d89a48f9ece  review.md
```
