# KR/US Same-Wave Independent Review P1 Remediation

Date: 2026-08-04

All changes below are implemented and tested as paired KR/US cases. No market is allowed to graduate before its peer.

1. `Attempt.DispatchStrategyVerified` MUST reject a nil official verifier before state or transport mutation.
2. A confirmed terminal fill, including terminal zero or partial fill observed during the ACK window, MUST release
   the aggregate remainder and each of the five monetary bucket remainders in the same terminal transaction.
3. A post-`Prepare`, pre-transport refusal MUST close the core attempt, exact strategy lease, aggregate reservation
   and five monetary reservations in one transaction. Failure injection MUST restore the whole preimage.
4. Production `bindApplyHooks` MUST bind `Campaign: journal.ApplyPositionCampaignFill` in the same one-time hook
   literal as Position, Exit and Costs. Strategy terminal backfill MUST continue to invoke campaign exactly once.

No item grants LIVE order authority or activation authority. The production assembly remains dormant.
