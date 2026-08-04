# Status — a069-add-kr-us-weekly-value-lanes

- Date: 2026-08-04
- Implementation: KR and US peer lanes complete and adversarially hardened in the same release unit
- Runtime state: both desired/effective OFF
- Integration boundary: private-sealed dormant pure port only; no runtime, scheduler, journal, broker or source client wiring
- Targeted, race, property/fuzz and selected regression tests: PASS
- OpenSpec strict validation: PASS
- Full-worktree regression: a069 and strategyrouter now pass; the run is blocked only by concurrent journal work causing `internal/position` eligibility-spelling checks on `internal/journal/position_campaign.go` and `internal/journal/strategy_evidence.go`
- SDD: `make sdd-sync` and one `make sdd-check` PASS; after final verification-document edits the shared fingerprint check became stale again while `codegraph status .` remained up to date at HEAD `23794f8626a20691431d5452b76e800255b0ee74`
- Hardening: cap/FX freshness, exact stable week, scoped CAS/count/ordinals, full evidence digest, atomic fill-risk aggregate, sealed private RiskState, sealed stop and complete lineage verified
- Gate: invoked and stopped at step 4/8 only because concurrent `internal/journal/**` functions lack their owning workstream's Function Logic Maps; all a069 maps pass and steps 1-3 passed
- Remaining: final `make sdd-sync`/shared gate/full regression rerun after concurrent journal changes settle

KR does not gate US and US does not gate KR. Each evaluator carries market-scoped disclosure,
revision, model, calendar-week, reservation, immutable plan and risk lineage.
