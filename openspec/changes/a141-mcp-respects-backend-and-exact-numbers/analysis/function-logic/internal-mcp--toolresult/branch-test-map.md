# Branch Test Map: `toolResult`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Non-marshallable payload | Existing RPC invalid-params behavior | no | pending |
| B2 | Small payload | `TestToolResultKeepsSmallPayload` | no | existing green |
| B2+B3+B4 | Oversized array payload | `TestToolResultCapsOversizedPayload` | no | existing green |
| B2+B3+B4 | Oversized nested payload retaining 9007199254740993, negative integer, and long decimal | New exact-token regression test | pending | pending |
| B3 | Exact decoder rejects impossible internal JSON and retains original text | Characterization through decoder seam | not-applicable | pending |
| B4 | Re-encoding failure preserves original marshalled text | Existing error-safe behavior; retain | no | pending |
