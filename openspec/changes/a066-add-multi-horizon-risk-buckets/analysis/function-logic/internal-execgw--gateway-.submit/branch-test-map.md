# Branch Test Map: `(*Gateway).submit`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | prepare request invalid | existing gateway validation suite | existing | GREEN |
| B2 | symbol claim conflict | existing symbol gate suite | existing | GREEN |
| B3 | symbol-free read error | existing symbol gate suite | existing | GREEN |
| B4 | symbol-free refusal branch | existing symbol gate suite | existing | GREEN |
| B5 | active symbol mutation refusal | existing symbol gate suite | existing | GREEN |
| B6 | decision-unspent read error | existing replay suite | existing | GREEN |
| B7 | decision-unspent refusal branch | existing replay suite | existing | GREEN |
| B8 | spent decision refusal | existing nonce suite | existing | GREEN |
| B9 | durable Prepare failure | existing journal failure suite | existing | GREEN |
| B10 | protection refusal | existing protection suite | existing | GREEN |
| B11 | entry gate refusal | existing entry gate suite | existing | GREEN |
| B12 | optional preflight present | existing preflight suite | existing | GREEN |
| B13 | preflight refusal | existing preflight suite | existing | GREEN |
| B14 | initial decision load refusal | existing decision suite | existing | GREEN |
| B15 | initial decision mismatch | existing decision suite | existing | GREEN |
| B16 | initial reservation/q_final refusal | `TestGatewayRefusesQFinalMarkedDecisionWithoutExactAdmissionBeforeBroker` | q_final missing reached broker path | GREEN |
| B17 | final fresh decision load refusal | `TestRevokedDecisionIsRefusedAtTheLastMoment` | existing | GREEN |
| B18 | final decision mismatch | existing expiry/tamper suite | existing | GREEN |
| B19 | final idempotency key mismatch | existing key tamper suite | existing | GREEN |
| B20 | final protection refusal | existing paired readiness suite | existing | GREEN |
| B21 | initial q_final check passes, then aggregate hold is released during the second protection checkpoint before the final q_final barrier | `TestGatewayLastMomentQFinalBarrierRefusesHoldReleaseAfterInitialAdmissionCheck` | missing last-moment check | GREEN; zero broker calls |
| B22 | dispatch transaction error | existing retry suite | existing | GREEN |
| B23 | nonce already spent | existing nonce suite | existing | GREEN |
| B24 | non-nonce dispatch error | existing retry suite | existing | GREEN |
| B25 | successful/settled outcome | existing gateway suite | existing | GREEN |
