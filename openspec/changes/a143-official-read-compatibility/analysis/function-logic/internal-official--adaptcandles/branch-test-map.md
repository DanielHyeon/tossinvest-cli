# Branch Test Map: `adaptCandles`

| Branch | Condition / scenario | Named test or static evidence | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range over wire candles; empty skips range | TestAdaptCandlesOrdersNewestFirstWireChronologically; TestAdaptCandlesEmpty | no conversion defect claimed | named tests PASS |
| B2 | nonempty timestamp enters parse | TestAdaptCandlesOrdersNewestFirstWireChronologically; TestAdaptCandlesBadTimestamp | no (preserved predicate) | named tests PASS; empty-timestamp branch not isolated |
| B3 | successful time.Parse assigns exact instant; failure leaves zero | TestAdaptCandlesOrdersNewestFirstWireChronologically; TestAdaptCandlesBadTimestamp | no (preserved parsing) | named tests PASS |
| B4 | local slice reversal loop; singleton/empty skip swaps | TestAdaptCandlesOrdersNewestFirstWireChronologically; TestAdaptCandlesUnit; TestAdaptCandlesEmpty | initial RED and independent reversal-removal mutant FAIL | named tests PASS |

Named test PASS is not an instrumented branch-hit claim. Initial RED and separate mutation runs are distinguished in analysis/verification-receipt.md. Unchanged error paths are not newly claimed as exhaustively covered.

Cross-path invariants are separate from these branch IDs: TestPollBindsEachSuccessorToTheNextNewerObservedBar and TestPollConvertsTheBrokerCloseLabelIntoAnOpenInstant passed in officialbars fixtures. Singleton evidence is TestAdaptCandlesUnit, not the three-candle test.
