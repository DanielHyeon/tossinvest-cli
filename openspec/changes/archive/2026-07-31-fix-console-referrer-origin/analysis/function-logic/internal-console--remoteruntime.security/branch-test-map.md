# Branch Test Map: `remoteRuntime.security`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| header setup | every remote response uses exact security headers and `Referrer-Policy: same-origin` | `TestRemoteResponsesHaveSecurityHeadersAndHealthIsMinimal` | failed: got `no-referrer` | passed |
| opaque origin boundary | explicit `Origin: null` with canonical TLS/Host and valid CSRF is origin-refused and cannot invoke the wrapped handler | `TestExplicitOpaqueOriginCannotReachMutationHandler` through `Console.mutating` | safety behavior passed while policy tests were RED | passed |
| B1 | unparseable or disallowed direct peer is refused after headers are set | existing remote peer rejection coverage | unchanged | passed in console suite |
| B2 | loopback GET `/healthz` reaches minimal handler while POST stays method-rejected | `TestRemoteResponsesHaveSecurityHeadersAndHealthIsMinimal` | unchanged | passed |
| B3 | wrong Host or disallowed CIDR cannot reach a console route | existing trusted-network remote tests | unchanged | passed in console suite |
| final | canonical console route reaches downstream handler | existing trusted dashboard coverage | unchanged | passed in console suite |
| browser contract | canonical Chrome form POST supplies canonical origin and reaches CSRF, using an invalid token | safe deployed Playwright probe recorded in `verification.md` | reproduced origin refusal | deployed probe sent canonical Origin/Referer and reached CSRF refusal |
