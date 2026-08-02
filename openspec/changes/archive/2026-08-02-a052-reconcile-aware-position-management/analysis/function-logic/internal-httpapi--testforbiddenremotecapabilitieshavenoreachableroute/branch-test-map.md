# Branch Test Map: `TestForbiddenRemoteCapabilitiesHaveNoReachableRoute`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | router setup succeeds; construction errors fail immediately | `TestForbiddenRemoteCapabilitiesHaveNoReachableRoute` | no new route was present | focused HTTP API suite passes |
| B2 | every forbidden path is enumerated | `TestForbiddenRemoteCapabilitiesHaveNoReachableRoute` | negative contract added for a052 | focused HTTP API suite passes |
| B3 | GET/POST/PUT/PATCH/DELETE are all denied | `TestForbiddenRemoteCapabilitiesHaveNoReachableRoute` | negative contract added for a052 | focused HTTP API suite passes |
| B4 | anything other than 404/405 fails | `TestForbiddenRemoteCapabilitiesHaveNoReachableRoute` | negative contract added for a052 | focused HTTP API suite passes |
