# Branch Test Map: `canonicalDecimal`

| Branch | Scenario | Test | RED | GREEN |
|---|---|---|---|---|
| B1 | blank canonicalizes to zero | canonical zero A110 cases | preserve | yes |
| B2 | malformed/non-finite remains visible | `TestA110ComparerRejectsIdenticalInvalidQuantityStrings` | yes | yes |
| Return | 2^53-adjacent finite decimals remain distinct | `TestA110ComparerPreservesExactLargeQuantityPromotionIdentity` | yes | yes |
