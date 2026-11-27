# Manager acceptance worksheet — a063

Status: code implementation and both independent review stages verified; final acceptance blocked.
Manager independently compared all 15 gstack-reviewed source/test/unit/operations SHA-256 values with
the current worktree: all matched. Actual gate and operational conditions below remain mandatory.

| Contract | Required evidence | Current disposition |
|---|---|---|
| One operating profile | Unit argv, shared full-path status derivation, default/custom/same-directory isolation tests | Verified in independent code/tests; actual installation pending |
| Existing behavior with diagnostics OFF | No status file, same issuer result/message/attestation and `ErrIncomplete` compatibility | Verified in independent code/tests |
| Single qualification decision | Typed issue/refusal flow with unchanged predicates/messages and no second evaluation | Verified by independent predicate/message comparison and tests |
| Failed renewal stays failed | Refusal/input/publication/status-write failure behavior; last good attestation preserved | Refusal, override rejection and status-write failure verified in isolated tests, including service stub exit 23. Source review supports error propagation; not every resolver, attestation reread, mkdir or attestation-save failure was directly injected (see branch maps). |
| Bounded diagnostic file | Atomic owner-only write; 4096-byte/16-code/schema/timestamp/regular-file validation matrix | Verified in independent code/tests |
| Operator can see the problem | Existing console render tests and isolated visual inspection; no raw error/account disclosure in diagnostics | Manager inspected refused-mobile.png and unknown-desktop.png on 2026-09-06: readable refusal/expiry warning, unknown without zero timestamp. Synthetic fixture proof only; running console deployment pending. |
| Advisory does not change safety authority | Exact 72h/12h boundaries, future/mismatch/absent/stale cases; unchanged usability and engine controls | Verified in independent code/tests; controlled warning-removal regression fails as expected |
| Current code evidence | Generated current and required base AST/Function Logic/Branch Test Maps | Eleven scoped bundles directly validated. Three base-revision test ASTs are retrospective evidence only. Global checker exits 1 with 327 missing rows outside scope; I1 remains. |
| Separate implementation review | Terra adversarial code review after implementation, findings resolved | Code clear after path-snapshot correction and independent reruns |
| Post-adversarial gstack review | gstack code/security/QA review on the final scoped diff | Code clear; 2 UI findings fixed and independently rerun; exact 15-file digests match |
| Engineering checks | Actual focused/full/seams/race/vet/validate/sdd-sync/sdd-check results | Recorded focused/full/seams/race passes; full/seams/race preceded final UI correction, which has focused console regression proof. Earlier administrative vet/sdd-sync/sdd-check passed; all-change validate failed for a119. Continuation sdd-check and diff-check exit 0; continuation sdd-sync final exit unavailable, recorded UNKNOWN, not a pass. |
| Concrete deployment approval | Template/binary/console scope, digest, profile, backup/reload/rollback and explicit human approval | Bounded proposal reviewed: all three procedural findings corrected and independently re-reviewed. Not executed; same-profile evidence and explicit operational approval remain pending I2. |
| Real operational proof | Actual three consecutive qualifying days, same-profile fresh attestation and true service outcome; no synthetic substitute | Pending I2 |
| Final gate | Actual `make gate CHANGE=a063-align-attestation-renewal-profile` exit 0 after prerequisites | Blocked I1/I2 |
| PM and archive | Correct Story/registry/generator, official spec-syncing archive, post-archive strict validation | Pending final acceptance |

Preserve the original base and unrelated checkout work. No global goal completion, next a0xx processing,
or archive follows engineering-only success while any required acceptance row remains unresolved.

## Continuation review — 2026-09-06

Manager independently recomputed all 15 recorded review digests again: no mismatch.
Browser screenshots are isolated synthetic fixtures, not production UI acceptance. The browser
shutdown failure recorded in `ui/qa.md` does not invalidate the saved images, but does not establish
a clean browser lifecycle either. Terra completed the scoped function-map verification; its separate
adversary confirmed the eleven-bundle result and verified corrections for all three deployment-plan
findings in `adversarial-continuation-review.md`. The follow-up gstack review independently matched the fifteen
implementation digests again. These results do not close the global gate or operational acceptance.
