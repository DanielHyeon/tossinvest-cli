# R2 — engine endpoint + CLI (code-reviewer), reviewed at 90e5170d

Verdict: no P0/P1. Single writer, auth, descriptor ordering, ErrUnwired, stopped-engine refusal and the
refused / unknown / 「완화됨·통지 실패」 wording hold.

| Sev | file:line (90e5170d) | Scenario | Evidence |
|---|---|---|---|
| P2 | risk_relaxation_command.go:119,158-165 | alert Title/Body carry `<AccountRef>` (raw broker account number in production) to the external transport; invariant 8 | Probe 5 published text includes `acct-7` |
| P2 | risk_bucket_relaxation.go:65 | typed-nil `(*audit.Log)(nil)` passes the journal guard → unaudited commit | Probe 1 `err=<nil> open-locks-after=0`, no audit line |
| P2 | transport :76 + journal :154,189-192 | tx on request ctx, no server deadline; stalled audit holds the single connection until the client's 5s timeout; afterwards a "released" audit line for a rolled-back tx | Probe 7: other writer `waited=4.5s`; `open-locks-after=1 audit-released-lines=1` |
| P3 | same | cancel between audit and Commit → audit says released, journal did not | Probe 2 |
| P3 | command :63-64,158 | `s.mu` held across tx and enqueue; slow enqueue stalls quarantine list/release | Probe 3 `quarantine-list-stalled=5.5s` |
| P3 | client :82-85 → CLI :259 | audit failure / 401 map to `internal` → CLI "outcome unknown" (conservative) | Probe 4 |
| P3 | command :167-170 | engine does not log a notice failure | read |
| P3 | readonly.go:78 | v35 tables not in the read-only check; show on pre-v35 → raw "no such table" | not probed |

Held: no `journal.Open` in the CLI release path; show uses mode=ro + query_only; both routes behind `server.auth` and
registered before descriptor publication; EventKey `engine.risk_relaxation|<kind>|<seq>` distinct (Probe 5); undelivered
notice latches entries (conservative); stale descriptor → "not running … nothing was released" (Probe 6); client timeout
after commit → "outcome unknown" and the release stood (Probe 3). Copy deleted; repo `internal cmd` status unchanged by R2.
