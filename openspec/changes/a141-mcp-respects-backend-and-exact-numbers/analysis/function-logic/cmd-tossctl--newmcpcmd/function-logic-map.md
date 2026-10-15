# Function Logic Map: `newMCPCmd`

- Source: `cmd/tossctl/mcp.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input | Contract |
|---|---|
| cfg.OpenAPI and opts.backend | Existing resolver determines preference before constructing clients |
| Credentials | Loaded through official.LoadCredentials; errors return; absent is permitted with WTS session |
| WTS session | Missing session is ordinary; unexpected read error returns |
| Official capability | Requires credentials AND Enabled AND non-WTS preference; otherwise client/trading dependencies remain nil |

## Branches and early returns

| Branch | Condition (AST source line) | Effect | Focused evidence |
|---|---|---|---|
| B1 | config load error (47) | return error | unchanged error path; no individual branch-hit claim |
| B2 | backend resolution error (51) | return before paths/client construction | resolver semantics unchanged; moved earlier |
| B3 | credential/token path error (58) | return error | unchanged error path |
| B4 | credential load error (62) | return error | unchanged error path |
| B5 | mcpOfficialEligible (67) | construct official and trading dependencies only when true | TestMCPCmdWTSDoesNotWireOfficialDependencies; TestMCPCmdDisabledAndEnabledBackendWiring |
| B6 | unexpected session load error (81) | return error | unchanged error path |
| B7 | session exists (85) | construct authenticated WTS client | synthetic session in guarded command tests |
| B8 | both effective clients absent (89) | return login guidance error | unchanged; no focused branch-hit claim |
| B9 | routedWTS nil (99) | construct sessionless client for hybrid embedding | unchanged; no focused branch-hit claim |
| B10 | update-cache path resolves (114) | bounded/cached update probe | guarded child denies external transport |
| B11 | latest version is newer (119) | append initialization instructions | unchanged; no focused branch-hit claim |

## Calls and live bindings

| Call | Purpose / error contract | AST evidence |
|---|---|---|
| resolveBackend | determine preference before construction; error returns | B2 |
| official.LoadCredentials | load configured material; error returns | B4 |
| official.New / trading.NewService | eligible direct dependencies; existing lineage/confirmation retained | B5 |
| FileStore.Load / tossclient.New | optional WTS session; unexpected load error returns | B6/B7/B9 |
| hybrid.New | bind routed reads to permitted official pointer and WTS client | call after B9 |
| mcp.NewServer / SetAuthStatus | publish dependencies and auth snapshot | calls after hybrid.New |
| updatecheck.New(...).LatestStable | cached update metadata; 2-second context; silent network failure | B10/B11 |
| Serve | serve stdio until EOF; return its error | terminal return |

## State mutations and fallbacks

Constructs local clients and lineage service. Construction does not place an order. The new gate controls direct official-read/trading reachability; existing WTS fallback, journal and trading checks are unchanged. Update metadata probing remains bounded and is denied external access in guarded tests.

## Safety conclusion

High-risk capability boundary. Focused tests cover denied and allowed dependency publication; mutations demonstrate enabled/WTS predicates are required. Unchanged error branches above are not falsely claimed as individually executed. Full-suite and gate receipts are separate completion requirements.
