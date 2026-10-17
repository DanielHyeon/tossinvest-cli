**Freeze remains blocked by inaccurate boundary claims and a contradictory new spec sentence—not simply by deferred projector wiring.**

Read-only review; no files changed, tests run, network calls, or engine execution. “RESOLVED” below means adequately addressed for proposal freeze, not implemented or verified by tests.

References: `D`, `T`, `Δ`, `P` mean this change’s `design.md`, `tasks.md`, `specs/engine-safety/spec.md`, and `proposal.md`. `AD` = `internal/app/engine/alertdelivery.go`; `EG` = `internal/app/engine/gateway.go`; `N` = `internal/obs/notifier.go`; `O` = `internal/journal/outbox.go`; `J` = `internal/journal/alert_claim.go`; `M` = `internal/journal/operating_mode.go`; `G` = `internal/execgw/retry.go`.

Earlier findings, individually:

| ID | Status | Why; evidence |
|---|---|---|
| F1 | RESOLVED | Transaction-local attempts read before commit is specified; existing return lacks it. D:40–46; J:330–334. |
| F2 | PARTIAL | Epoch ordering is repaired; operational protection remains narrower than some boundary claims. D:269–279,303–304; AD1–AD2 below. |
| F3 | RESOLVED | Row-specific claim/record-error counters and list-error counter cover failures without durable increments. D:341–378; AD:168–171,222–225. |
| F4 | RESOLVED | Serial batch service included; 214/316-second calculations follow stated assumptions. D:153–168; AD:133–161. |
| F5 | RESOLVED | Selection guarantee is narrowed and residual starvation recorded; “already blocked” needs AD1’s qualification. D:124–133; AD:178–192. |
| F6 | RESOLVED | Deployment escalation requires selection and another failed attempt; startup separately restores pending-alert latch. D:500–502; EG:153–167. |
| F7 | RESOLVED | OFF boundary is startup refusal, not missing Notifier. D:488–490; `cmd/tossctl/engine.go:220–221`; EG:323. |
| F8 | RESOLVED | Fixed details, allowlisted new logs and escalation-error fallback specified. D:397–413; `internal/obs/log.go:162–165`. |
| F9 | RESOLVED | First detail remains; repeated `Block` does not update it. D:78; G:526–533. |
| F10 | RESOLVED | Error precedes interpretation of zero-valued `SettleApplied`. D:49; J:109,322–334. |
| F11 | RESOLVED | Like-for-like single-row timing replaces mixed failure modes. D:86–96; N:428–433,525–528. |
| F12 | RESOLVED | Missing-publisher write cost has an explicit contention measurement obligation. T:32–39; `internal/journal/journal.go:174`. |
| F13 | RESOLVED | Approximately immediate versus 4–6-second missing-publisher response is disclosed. D:106–107; N:429–431,570–573. |
| N1 | RESOLVED | Executor takes no Notifier mutex; journal/logging remain outside gate lock. D:269–279,322–332. |
| N2 | RESOLVED | Error judgments use explicit conservative ordering rather than unreliable acknowledgement-state suppression. D:282–285,302; Δ:39–43. |
| N3 | RESOLVED | Published-but-unsettled failure immediately judges block/escalation and retains lease. D:64–76; N:465–491. |
| N4 | RESOLVED | Counter lifetime, rearm carryover and pruning specified. D:341–378; O:334–340. |
| N5 | RESOLVED | General upper bound withdrawn; conditional assumptions and unbounded waits named. D:144–177. |
| N6 | RESOLVED | Failed-attempt `NotFound` blocks without escalation, matching synchronous outcome propagation. D:61; N:519–523,309–315. |
| N7 | PARTIAL | Failure policies are covered, but new boundary requirement contradicts retained exceptions. Δ:13–16 versus :33–43; AD2. |
| N8 | RESOLVED | Deployment wording no longer promises unconditional first-cycle escalation. D:500–502. |
| R1 | RESOLVED | Executor-held `n.mu` across database waits is removed. D:322–332; N:254–255,851–852. |
| R2 | RESOLVED | New logs exclude account and raw-error strings. D:404–408; M:393–394,448–449. |
| R3 | RESOLVED | Other-row success and zero-write settlements cannot reset the affected row. D:341–350; J:337–365. |
| R4 | RESOLVED | Global pending-count fence removed; reason-specific clear epochs replace it. D:239–263,359–375. |
| R5 | RESOLVED | Lock contract consistently excludes journal, logging and remote work. Δ:45–48; T:86–87. |
| R6 | RESOLVED | Duplicate suppression explicitly ends when lease expires or is replaced. D:74–76; J:162–165; T:65–68. |
| R7 | RESOLVED | Expiry alone does not invalidate settlement; state/token CAS governs it. D:179–181; O:455,472; J:349–365. |
| R8 | RESOLVED | Per-row counting prevents one batch’s different rows reaching the threshold. D:380; AD:133,154–161. |
| Q1 | RESOLVED | Executor no longer creates `n.mu → g.mu` dependency. D:322–332; `internal/execgw/strategy_entry_gate_authority.go:60–72`. |
| Q2 | RESOLVED | Epoch advances at `Clear`, not acknowledgement start. D:244–249; N:851–875. |
| Q3 | RESOLVED | Failed/partial acknowledgements leave epoch unchanged because they do not clear. N:864–875; T:52–59. |
| Q4 | RESOLVED | Delivery/rearm without clear retains old judgment conservatively. D:299,356–357; O:334–340. |
| Q5 | RESOLVED | Complete-list pruning and truncated-list retention are explicit. D:346,353–354. |
| Q6 | RESOLVED | Formula includes listing/escalation costs and six assumptions; handoff replaces old formula. D:153–185; T:107–108. |
| Q7 | RESOLVED | F2 disposition identifies B′ rather than obsolete mutex design. D:522. |
| V1 | RESOLVED | Superseded by explicit conservative error policy; acknowledgement timestamps no longer suppress judgments. D:282–285,302. |
| V2 | PARTIAL | Count–clear race is acknowledged, but cited production trace already carries another blocking reason. D:303,455–458; AD1. |
| V3 | RESOLVED | Documents agree on post-settlement sampling, after release on failed-send path. D:272–274; T:86. |
| V4 | RESOLVED | Contract permits bounded gate-map operations instead of forbidding every shared lock. Δ:45–48; G:526–543. |
| V5 | RESOLVED | Rearm carryover and list-counter reset behavior specified. D:356–378; T:59–62. |
| V6 | RESOLVED | Final listing cost and corrected handoff formula present. D:159–185; T:107–108. |
| W1(a) | RESOLVED | Delayed application does not uniquely lose the alert latch: the same later `Clear` removes an on-time latch. D:303–304; N:875; G:537–543. |
| W1(b) | NOT RESOLVED | Revision 7 honestly withdraws mode-based protection, but neither current row persistence nor a broadly described count–clear repair proves B-specific protection. D:459–462; AD1, AD3. |
| W2 | RESOLVED | List failures use explicit epoch/counter transitions without nonexistent row lookup. D:359–378. |
| W3 | RESOLVED | Retry-count-based unconditional blocking is gone; fallback is specifically escalation-write failure. D:232–235,319–320. |
| W4 | RESOLVED | Block-only judgments remain block-only through application. D:227; T:43,53; N:519–523. |
| W5 | RESOLVED | Current epoch requirements and implementation tasks align. Δ:reason-specific epoch requirement; T:70–73,86–87. |
| X1 | RESOLVED | Deferred-judgment map removed. D:319–320. |
| X2 | RESOLVED | Clear ordering handles delivered-then-manually-cleared rows without requiring impossible ACK state. D:305; O:494–503. |
| X3 | RESOLVED | Deferred map removed; pruning exceptions and release exclusion enumerated. D:341–351; Δ:50–57. |
| X4 | RESOLVED | Journal contention distinguished from mutex isolation; no unchanged-dwell claim. D:322–326; T:32–39. |
| X5 | RESOLVED | Current disposition and B8 map match baseline’s log-and-release behavior. D:558,572; AD:205–212. |
| Y1 | RESOLVED | Escalation survives late clear as durable state; revision 7 expressly retracts enforcement inference. D:226–230,453–454; M:468–476. |
| Y2 | RESOLVED | Timestamp ordering removed; acknowledgement timestamp precedes actual update. D:282–285; O:485,494. |
| Y3 | RESOLVED | Spec includes settled-claim, complete-list and successful-list pruning. Δ:50–57; D:341–351. |
| Y4 | RESOLVED | Isolation assertions and shared-connection measurements separated. T:32–39; `internal/journal/journal.go:174`. |
| Y5 | RESOLVED | Failed-send release precedes sampling; retained-lease expiry risk disclosed. D:273,287–289. |
| Y6 | RESOLVED | B8 baseline and count–clear chronology corrected; whole-gate inference remains AD1. P:34–36,132–136; AD:205–212. |
| Z1 | RESOLVED | Threshold judgment precedes epoch reset and preserves escalation. D:359–375; T:58–61. |
| Z2 | RESOLVED | Escalation failure triggers unconditional fallback block. D:232–235,279; Δ:33–34. |
| Z3 | RESOLVED | Judgment/action table preserves block-only outcomes and applicable escalation. D:224–230; Δ:30–34. |
| Z4 | RESOLVED | SELECT/commit fault tests required for all settlement callers. D:40–44; T:40–41. |
| Z5 | RESOLVED | Fixed tolerance and deliberately failing control replace self-expanding budget. T:35–39. |
| AA1 | RESOLVED | Fallback independent of earlier conditional-block success. D:232–235,279; T:54. |
| AA2 | RESOLVED | Threshold-first transition and three expected outcomes agree. D:359–375; T:58–61. |
| AA3 | RESOLVED | Fixed 250 ms margin reused. T:36; `internal/app/engine/a098_the_backlog_does_not_delay_protection_test.go:61–65`. |
| AA4 | RESOLVED | Additive delivery-list method preserves existing API/callers. D:111–113; O:517–529; N:855. |
| AB1 | RESOLVED | Full-backlog query cost acknowledged; size measurements and stop condition mandatory. D:115–122; T:38–39. |
| AB2 | RESOLVED | Conservative exceptions remain explicit, although new boundary paragraph conflicts with them. D:309–317; Δ:33–43,93–105; AD2. |
| AC1 | PARTIAL | Missing production wiring is now honestly disclosed; D10 still overstates whole-gate outcomes and Δ overstates post-clear behavior. D:417–469; AD1–AD2. |
| AC2 | RESOLVED | Adequately recorded as an unfixed dependency defect, with correct code coordinates and required atomic/ordered projection. D:475–480; `internal/execgw/modegate.go:35–50`; M:468–476. |

New findings:

| ID | Severity | Finding | Evidence | Suggested fix |
|---|---|---|---|---|
| AD1 | **P1** | **D10 conflates absence of a124’s alert latch with an open production entry gate.** Its specific `parkAlert` example sets `ReasonUnresolvedInDoubt` **before** enqueueing B; alert acknowledgement does not clear that reason. Fresh production gates also reject unobserved required queries. Thus “production entry opens” and “HEAD only has startup protection for B” are not established by the cited path. | D:303–304,420,442–445,455–462; P:132–136. `internal/execgw/replay.go:534–551`; N:875; G:565–590; EG:249,261–269. | State “a124 supplies no alert/mode block here; other checks remain authoritative.” Include producer-specific unresolved blocking, sender-down, reconciliation and freshness in the boundary’s exclusions. Qualify D4:130–131 similarly. |
| AD2 | **P1** | **New normative post-release claim contradicts unchanged judgment rules.** Δ says this requirement blocks no entry after operator release. Yet the same requirement permits settlement→clear→epoch-read reblocking and mandates block→clear→escalation-error reblocking. This contradiction needs no projector or new failure policy. | Δ:13–16 versus :33–43,:83–84,:93–105; D:269–279,309–311; T:45–46,54. | Limit the statement to a quiescent, successfully completed judgment with no permitted reblocking exception or later judgment. Prefer “the durable mode row adds no enforcement” over an unconditional post-clear outcome. Add both exception variants to 2.14. |
| AD3 | **P2** | **The proposed W1 closers need narrower sufficiency claims.** Atomic count–clear excludes the exact enqueue-between-count-and-clear trace. It does not establish that every delivered B was covered by the acknowledgement: B can arrive after acknowledgement’s ID snapshot, fail, then be delivered before the final atomic count–clear. Projector wiring also requires the already-recorded AC2 fixes and restoration before admission. | D:455–468; N:854–875; O:453–456,494–503; a092 delta:72; M:468–476,574–588. | Specify which trace each closer eliminates. For general B-specific acknowledgement protection, require acknowledgement/producer-generation coverage beyond merely atomic count–clear. Describe operational enforcement as wiring **plus** safe projection/startup ordering. |
| AD4 | **P3** | **Reader/provenance statements are imprecise.** No production enforcement reader is bound, but transitions themselves read `operating_modes`. The cited canonical lines 1030–1037 do not contain the missing-binding explanation in this supplied base. | D:304,417–424; P:116; M:398,676–685; `openspec/specs/engine-safety/spec.md:995–1000`. | Say “no production admission/startup consumer”; distinguish transition-internal readers. Correct canonical coordinates to 995–1000. |

The central D10 discovery is correct. Repository-wide non-test searches find no calls binding `SetModeProjector` or invoking `RestoreOperatingModeProjection`. Transitions project only when a projector exists, after commit (`M:468–476`). Startup restores the alert latch when `UndeliveredCount > 0`; count failure rejects startup, while zero pending rows produces no alert latch (`EG:153–167,269–270`). Tracer supplies `risk.ModeNormal` (`internal/app/engine/tracer.go:473`), so its risk-mode check does not read the durable row.

That does **not** mean these are the only entry controls. Besides the unresolved latch above, sender death directly blocks entry (`internal/app/engine/auxiliary.go:209`); entry checking includes all account latches and freshness (`G:565–590`), reconciliation refresh and symbol blocks (`internal/execgw/symbolgate.go:229–250`). D10 should describe a124’s contribution, not infer overall admission from two absent reasons.

W1(a)’s equivalence argument therefore survives. W1(b) is honestly labelled unresolved in revision 7, but “B’s alert protection is absent” and “production admits entry” must remain distinct. Count–clear repair suffices for the **exact** stated interleaving only when it covers every relevant producer; it is not a general acknowledgement-coverage proof. Correctly ordered, atomic mode enforcement and startup restoration provide the separate durable protection, assuming successful escalation and no approved relaxation.

AC2’s coordinates are correct. Its manifestation condition should say **overlapping transition calls with reordered post-commit projections**, not literally simultaneous commits: SQLite serializes writes, but that does not serialize the callbacks after commit. The delete/unlock/reinsert gap is directly visible at `modegate.go:36–50`.

Tasks 2.14(a–b) appropriately prevent fixture-only enforcement. Task 2.14(c) is implementable, but must distinguish “neither alert nor mode reason exists” from `CheckEntryFor == nil`; a production-built gate starts with required queries unobserved. Keep production thresholds, provide legitimate fresh observations for a controlled admission case, and separately assert both reasons and the durable row. Include AD2’s reblocking cases rather than treating the happy-path pin as universal.

Calling projector work an **operational-effect precondition rather than code-landing precondition is defensible** for the narrowed change. a124 can add conservative alert-latch judgments without removing HEAD’s existing blocks or changing acknowledgement. It cannot claim durable mode enforcement or complete W1 protection. The fixed-budget journal-contention checks remain implementation gates. Also, ownership is presently a handoff commitment: the supplied a092 proposal still explicitly calls wiring an unassigned follow-up (`a092/proposal.md:499–503,582`).

No judgment-rule change was identified in D1’s outcome tables, D7’s action table/order, or D8’s threshold-first transitions. They agree with the stored round-12 review’s description of revision 6 fifth printing. **Exact revision-to-revision verification is unavailable:** this export has no Git metadata or complete prior design snapshot; the review’s assertion alone cannot establish a textual diff.

D2 and D3 remain adequately justified. Limit **3** matches `N:45`; the 4/34-second comparisons are correctly limited to the single-row failure patterns. Counting missing publishers matches existing synchronous conservative behavior (`N:429–431,570–573`) and a092’s explicit requirement (`a092/specs/engine-safety/spec.md:76`). The approximately immediate versus 4–6-second difference is disclosed, not disguised as timing equivalence.

D10 must distinguish missing alert/mode protection from overall entry admission, and the new post-release SHALL must acknowledge the unchanged reblocking exceptions.

VERDICT: REJECT