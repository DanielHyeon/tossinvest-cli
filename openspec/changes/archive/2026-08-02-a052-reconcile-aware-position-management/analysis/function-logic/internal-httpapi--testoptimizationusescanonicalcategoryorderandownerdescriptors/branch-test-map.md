# Branch Test Map: `TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 114 | `if err != nil {` entered and complementary path | TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors | pre-existing regression at frozen base | verified by current package suite |
| B2 | `range` line 119 | `for _, category := range resource.Categories {` entered and complementary path | TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors | pre-existing regression at frozen base | verified by current package suite |
| B3 | `if` line 123 | `if !reflect.DeepEqual(gotIDs, wantIDs) {` entered and complementary path | TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors | pre-existing regression at frozen base | verified by current package suite |
| B4 | `if` line 126 | `if resource.PositionManagement.AutoEnabledDefault \|\| resource.PositionManagement.StopDefault != "5%" \|\|` entered and complementary path | TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors | pre-existing regression at frozen base | verified by current package suite |
| B5 | `if` line 130 | `if len(resource.CandidateFilters) != 2 \|\| len(resource.CandidateFilters[0].Filters) == 0 {` entered and complementary path | TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors | pre-existing regression at frozen base | verified by current package suite |
| B6 | `if` line 134 | `if first.DefaultState != "unapproved" \|\| first.DesiredValue != "" \|\| first.EffectiveValue != "" {` entered and complementary path | TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors | pre-existing regression at frozen base | verified by current package suite |
| B7 | `if` line 137 | `if len(resource.Fields) != 1 \|\| resource.Fields[0].Key != "exit.common-policy" \|\| resource.Fields[0].Owner != "a041-complete-exit-line-contract" {` entered and complementary path | TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors | pre-existing regression at frozen base | verified by current package suite |
