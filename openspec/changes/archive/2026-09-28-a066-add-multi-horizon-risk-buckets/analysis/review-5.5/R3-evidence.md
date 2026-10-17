# R3 — evidence (general-purpose), reviewed at 90e5170d

Verdict: core journal safety claims pinned by behavioural tests whose mutants go RED; many D8 claims only declared or
pinned by the exit census; committed ledger stale. No P0 in code; P2 evidence debt + one stale operator document.

| Sev | Where | Finding | Evidence |
|---|---|---|---|
| P2 | mutation-5.5-relaxation/journal-ledger.tsv | stale (pre-fix HEAD, old format, no E*/C* rows) | re-run: M11/M20/M21 CAUGHT, E10 and C04 SURVIVED |
| P2 | harness run_tests | build failure counted as CAUGHT; M22 false CAUGHT; compiling M22b SURVIVED | M22.log `[build failed]` |
| P2 | relaxation.go:308 | M19/M19b caught only by the exit-count census | census message only |
| P2 | v35 SQL triggers + UNIQUE | MX01–MX03, MX05–MX07 SURVIVED (only lock-release DELETE tested) | ledger-all.tsv |
| P2 | transport :36-39 | route auth unpinned (EX13/EX14 SURVIVED) | |
| P2 | command :101 | latch path skipping relaxationAuditor survives (EX16) → typed-nil unaudited commit | |
| P2 | relaxation.go:316 | generation filter (MX17), overage_minor kept (MX18), released owner (MX20) unpinned | |
| P2 | CLI :80; readers | show read-only unpinned (CX21 journal.Open survives); ReadOnly LastEvent (MX11) / digest (MX12) | |
| P2 | audit census | misses local variable shadow, method value, function-local const sorting earlier | probes P2/P3/P4 rc=0 |
| P2 | status.md:8, tasks.md:37 | stale: CLI opens journal.Open; "relaxation not implemented" | |
| P3 | relaxation_test.go:105 | M03 caught only by FOREIGN KEY; no other-scope real lock test | |
| P3 | CLI :255-256 | ErrAuditUnavailable refusal mapping unpinned (CX19) | |
| P3 | command :158 | WithoutCancel unpinned (E10) | |
| P3 | D8 no engine lock / nothing delays | declared only; commit failure after audit leaves "released" | not probed |
| P3 | D8 scope latch rows untouched | no test | not probed |
| P3 | FLM/BTM | four ast.json byte-identical to regenerated; nit: audit field line 103 → 102 | |

Pinned (mutant RED): M01, M02, M04, M05–M08, M21, M09, M16, audit after commit (MX08, MX09), extra audit line (MX10),
M13, M14, M15, M17, M12, M11, E01–E08, C06/CX15. Copy deleted; the modified files seen at the end were the implementer's
concurrent repair, not R3's.
