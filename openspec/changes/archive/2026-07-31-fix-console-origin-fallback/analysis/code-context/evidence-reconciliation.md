# Evidence reconciliation

| Question | CodeGraph / HEAD hard evidence | CodeGraphContext candidate | Reconciled conclusion |
|---|---|---|---|
| Is the URL path compared? | `sameOrigin` uses `Origin` verbatim or rebuilds only `u.Scheme + "://" + u.Host` from Referer | No specific candidate | Path is not compared and must remain irrelevant |
| Why do all setting forms fail together? | Every state-changing route passes `Console.mutating`, which calls the same `sameOrigin` function before parsing CSRF | No specific candidate | One common origin gate is the defect boundary |
| Can missing headers safely fall back? | Outer middleware already checks peer CIDR and exact Host; remote listener is TLS; `mutating` still requires process-local CSRF | No contradiction | Only when both headers are absent, require direct TLS and exact canonical Host |
| Can an explicit wrong Origin fall through? | Existing function returns the explicit value comparison; threat model requires fail-closed | No contradiction | Explicit Origin/Referer mismatch remains final and must not reach request-host fallback |
| Does this affect trading? | Impact set is console construction/routes/login only | No candidate | No LIVE order, engine, Guardian, exit or journal behavior changes |

No unresolved evidence conflict remains. Production edit boundary is
`internal/console/remote.go` plus its focused tests.
