# Branch Test Map: `quantitiesAgree`

| Branch | Scenario | Test | RED | GREEN |
|---|---|---|---|---|
| B1 | identical NaN/Inf/malformed is not agreement | `TestA110ComparerRejectsIdenticalInvalidQuantityStrings` | yes | yes |
| B2 | exact valid equality remains clean | `TestAgreementIsClean` | preserve | yes |
| B3 | unreadable numeric conversion disagrees | invalid comparer table | preserve | yes |
| B4 | 2^53-adjacent exact integers do not collide | `TestA110ComparerDoesNotTreatDistinctLargeIntegersAsEqual` | yes | yes |
| Return | proven 0.3 binary expansion is accepted, generic one-ULP exact integers and MaxFloat disagreement are rejected | `TestFractionalQuantitiesSurviveTheFloatRoundTrip`, `TestA110ComparerDoesNotTreatDistinctExactIntegersOneULPApartAsArtifact`, `TestA110ComparerMaxFloatULPDoesNotOverflowToEquality` | yes | yes |
