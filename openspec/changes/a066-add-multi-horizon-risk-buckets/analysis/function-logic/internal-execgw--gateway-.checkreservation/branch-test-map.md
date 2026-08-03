# Branch Test Map: `(*Gateway).checkReservation`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | risk-reducing exit/cancel bypasses all reservation checks | existing reservation gate tests | existing | GREEN |
| B2 | legacy reservation read error fails closed | existing reservation gate tests | existing | GREEN |
| B3 | iterate durable aggregate reservations | existing gateway tests | existing | GREEN |
| B4 | HELD aggregate reservation triggers q_final revalidation | `TestGatewayRefusesQFinalMarkedDecisionWithoutExactAdmissionBeforeBroker` | q_final absence bypassed | GREEN |
| B5 | revoked decision during revalidation keeps Guardian-missing reason | `TestRevokedDecisionIsRefusedAtTheLastMoment` | reason regression | GREEN |
| B6 | q_final owner/five-reservation mismatch refuses | `TestGatewayRefusesQFinalMarkedDecisionWithoutExactAdmissionBeforeBroker` | broker boundary lacked check | GREEN |
