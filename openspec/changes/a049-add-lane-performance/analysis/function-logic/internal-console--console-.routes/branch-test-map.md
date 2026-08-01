# Branch Test Map: `Console.routes`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | token-authenticated remote registers login/logout; trusted-network mode does not require application login | `TestRemoteLoginIssuesADistinctBoundSecureSession`; `TestTrustedNetworkConsoleNeedsNoApplicationSession` | no — unchanged existing branch | yes |
| B2 | non-nil remote returns the security-wrapped complete mux | `TestRemotePeerHostOriginAndCSRFAreIndependentGates`; `TestRemoteResponsesHaveSecurityHeadersAndHealthIsMinimal` | no — unchanged existing branch | yes |
| fallthrough | local mux keeps session gates and the new performance path stays method-read-only | `TestEveryRouteRefusesARequestWithoutTheSessionToken`; `TestPerformanceHistoryIsMethodReadOnlyMobileAccessibleAndCSPCompatible` | yes — performance route absent before implementation | yes |
