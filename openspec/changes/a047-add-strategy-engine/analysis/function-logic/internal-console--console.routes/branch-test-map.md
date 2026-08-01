# Branch Test Map: `Console.routes`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | untrusted remote mode preserves login/logout and security wrapper | existing remote console tests | baseline | baseline |
| B2 | loopback returns authenticated application mux | existing session/route tests | baseline | baseline |
| A047 | `/strategy-runtime` is authenticated GET/HEAD-only, with POST/PUT/PATCH/DELETE 405 | `TestStrategyRuntimeStatusIsAuthenticatedReadOnlyAndHasNoControls` | yes | yes |
| A047 | nil/error reader renders blocker-complete OFF/not_configured state | `TestStrategyRuntimeStatusFailsClosedOnNilOrReaderError` | yes | yes |
