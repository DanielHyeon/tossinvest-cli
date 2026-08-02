# Branch Test Map: `TestPositionManagementDoesNotOfferReadoptForIneligibleHolding`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 294 | `if strings.Contains(page, "새 generation 재편입") {` true/entered and complementary path | TestPositionManagementDoesNotOfferReadoptForIneligibleHolding | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B2 | `if` line 297 | `if !strings.Contains(page, "관리 외(운영자 해제)") \|\| !strings.Contains(page, "OPERATOR_RELEASED") {` true/entered and complementary path | TestPositionManagementDoesNotOfferReadoptForIneligibleHolding | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
