# Branch Test Map: `Console.routes`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | All routes are registered with session; update install has CSRF; start routes are exclusive | static route tests and `TestSystemUpdateSerializesAConcurrentEngineStartThroughCommit` | update route absent/unserialized | pass |
