# Branch Test Map: `evaluateEvidence`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | caller literal and mutated decoded evidence fail closed | TestDecodedEvidenceSealRejectsLiteralAndMutation | yes | yes |
| B2 | financial/revision/PIT mutation changes or invalidates digest | TestDecisionDigestCoversFullImmutableEvidence | yes | yes |
| B3 | valid KR/US replay deterministic | TestStrictOpenDARTAndEDGARSchemasAcceptExactPointInTimeReplay | existing | existing |
| B4 | duplicate JSON rejected | TestStrictSchemasRejectDuplicateUnknownFutureRevisionAndBrokenPreimage | existing | existing |
| B5 | config seal mismatch | TestStrictSchemasRejectDuplicateUnknownFutureRevisionAndBrokenPreimage | existing | existing |
| B6 | evidence market mismatch | TestKRAndUSWeeklyEvaluationAreSourceAndMarketIndependent | existing | existing |
| B7 | evidence source mismatch | TestKRAndUSWeeklyEvaluationAreSourceAndMarketIndependent | existing | existing |
| B8 | schema/model/config mismatch | TestStrictSchemasRejectDuplicateUnknownFutureRevisionAndBrokenPreimage | existing | existing |
| B9 | bounded required identities | TestStrictSchemasRejectDuplicateUnknownFutureRevisionAndBrokenPreimage | existing | existing |
| B10 | revision chain invalid | TestStrictSchemasRejectDuplicateUnknownFutureRevisionAndBrokenPreimage | existing | existing |
| B11 | PIT timestamp missing | TestStrictSchemasRejectDuplicateUnknownFutureRevisionAndBrokenPreimage | existing | existing |
| B12 | PIT ordering invalid | TestStrictSchemasRejectDuplicateUnknownFutureRevisionAndBrokenPreimage | existing | existing |
| B13 | evidence freshness invalid | TestKRAndUSWeeklyEvaluationAreSourceAndMarketIndependent | existing | existing |
| B14 | dilution cutoff/status invalid | TestStrictSchemasRejectDuplicateUnknownFutureRevisionAndBrokenPreimage | existing | existing |
| B15 | currency/unit mismatch | TestStrictSchemasRejectDuplicateUnknownFutureRevisionAndBrokenPreimage | existing | existing |
| B16 | diluted shares absent | TestStrictSchemasRejectDuplicateUnknownFutureRevisionAndBrokenPreimage | existing | existing |
| B17 | financial vector cardinality invalid | TestStrictSchemasRejectDuplicateUnknownFutureRevisionAndBrokenPreimage | existing | existing |
| B18 | input identity/unit duplicate invalid | TestStrictSchemasRejectDuplicateUnknownFutureRevisionAndBrokenPreimage | existing | existing |
| B19 | input arithmetic invalid | TestStrictSchemasRejectDuplicateUnknownFutureRevisionAndBrokenPreimage | existing | existing |
| B20 | equity/fair arithmetic invalid | TestStrictSchemasRejectDuplicateUnknownFutureRevisionAndBrokenPreimage | existing | existing |
| B21 | fair-value preimage mismatch | TestStrictSchemasRejectDuplicateUnknownFutureRevisionAndBrokenPreimage | existing | existing |
