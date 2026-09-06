# Status — a064-add-multi-market-strategy-evidence

- Date: 2026-09-07 (previous entries 2026-08-04 and the 2026-09-07 completion pass)
- State: implementation landed 2026-08-04; the completion pass found it not complete; the remediation pass
  below closed every finding. Remaining work is the gate run and archive.
- Completed in evidence-source slice: frozen SEC/OpenDART fixture contracts, KRX zero-call unavailable gate,
  sealed deployment-policy minting, bounded official pagination, shared rate/concurrency budget, secret
  boundary, immutable complete-batch commit gate, trusted ingestion clock and adversarial evidence-core
  hardening.
- Completed in Wave 1C: v21 nullable snapshot-ID/digest-only journal lineage, exact immutable replay,
  SELECT-only read ports, full-Header snapshot digest binding, scope/dual-cutoff replay revalidation,
  database-level partial-reference guards, and independent KR/US failure scope.

## Completion pass, 2026-09-07

- Comparison base re-pointed `c57915dd` → `bc03c4d4`. 307 unrelated commits had made `check_analysis.py`
  attribute 341 foreign functions to a064; a064's own two commits modify exactly three pre-existing Go
  functions, all byte-identical at HEAD. See review.md.
- Repository-wide gates all passed, but two independent reviews returned **BLOCK**: the load-bearing
  properties were enforced by no test, one checked task's artifact was never built, and one `SHALL` had no
  implementation while a test pinned the opposite. Findings are in `issues.md` (I1-I27).

## Remediation pass, 2026-09-07

The human chose full remediation (A-1) and narrowing the Requirement to the dormant scope (B-1).

- **I1-I9, I11-I14 (untested or unmeasurable properties)** now have tests that fail when the property is
  broken. Every one was falsified with `go test -overlay` and the mutation result is recorded in review.md
  and issues.md; each instrument carries a positive control.
- **I15-I19 (production defects)** are fixed: transport errors carry no credential or authenticated URL, a
  spoofed SEC response cannot choose the next request path, every policy field is bounded by the frozen
  official contract, `NewAdapter` is unexported so no adapter can skip the shared rate budget, and a
  missing optional fatal fact is recorded instead of vanishing.
- **I10** — the named broker spy now exists as two instruments that can fail: a transitive import-closure
  walk and a whole-journal row-count digest across the dormant read. Task 6.2 is checked again.
- **I20** — the Requirement was narrowed to what this change implements. Two SHALLs replace the one that
  had no implementation, and the Requirement now states in the open that forcing a snapshot reference as an
  order-entry precondition belongs to the lane-wiring change.
- Eleven Function Logic Map bundles cover the modified existing functions, so I27 ("the logic-map step is
  vacuous for a064") is no longer true.

## Corrections to earlier entries

- **"dormant / no runtime activation" is now pinned rather than asserted.**
  `internal/strategyevidence` is imported by six production files, so the earlier claim had decayed. What
  holds and is now tested is narrower and true: the package's transitive import closure inside this module
  is `internal/clock` and nothing else — no broker, dispatch, execgw, Guardian, journal, order or toggle
  package, and no `net/http`.
- **"KRX performs zero calls"** is true by mechanism (four independent hardcodes; `SourcePolicyConfig`
  carries no endpoint fields) and is now also bound to `testdata/official_contracts.json`, which declares
  the KRX contract unfrozen.
- **"a064 journal/evidence race tests PASS"** and the other transcribed timings remain hand-copied prose
  with no stored artifact (issues.md I26). The 2026-09-07 gate and mutation tables in review.md are
  re-runnable.

## Safety

- No broker/order/toggle path was added and no LIVE authority was granted. The claim is now backed by a
  test that fails if it stops being true, not by import analysis done once by hand.
- The credential-leak fix (I15) is verified against a synthetic `*url.Error` shaped like `net/http`'s; the
  module still contains no `Transport` implementation, which is why the defect was latent rather than live.
- Two fail-closed rules were added. Both name what they reject: the SEC page-name rule accepts exactly the
  frozen `CIK<requested CIK>-submissions-<3 digits>.json` form, and the policy caps admit both minted
  policies. Each has a positive control proving it does not refuse normal input.
