# Branch Test Map: `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | base-revision `range` at line 136: `for _, test := range tests {`; hunk adjacency requires base evidence | `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity` (this regression test) | existing regression branch | package-targeted regression PASS before integration; rerun by gate |
| B2 | base-revision `if` at line 140: `if err := trade.validate(); err == nil {`; hunk adjacency requires base evidence | `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity` (this regression test) | existing regression branch | package-targeted regression PASS before integration; rerun by gate |
