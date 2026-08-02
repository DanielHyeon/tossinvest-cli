# Branch Test Map: `TestEveryOpenAPILocalReferenceResolves`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 66 | `if err != nil {` entered and complementary path | TestEveryOpenAPILocalReferenceResolves | pre-existing regression at frozen base | verified by current package suite |
| B2 | `if` line 70 | `if err := json.Unmarshal(raw, &document); err != nil {` entered and complementary path | TestEveryOpenAPILocalReferenceResolves | pre-existing regression at frozen base | verified by current package suite |
| B3 | `type-switch` line 75 | `switch typed := value.(type) {` entered and complementary path | TestEveryOpenAPILocalReferenceResolves | pre-existing regression at frozen base | verified by current package suite |
| B4 | `case` line 76 | `case map[string]any:` entered and complementary path | TestEveryOpenAPILocalReferenceResolves | pre-existing regression at frozen base | verified by current package suite |
| B5 | `range` line 77 | `for key, child := range typed {` entered and complementary path | TestEveryOpenAPILocalReferenceResolves | pre-existing regression at frozen base | verified by current package suite |
| B6 | `if` line 78 | `if key == "$ref" {` entered and complementary path | TestEveryOpenAPILocalReferenceResolves | pre-existing regression at frozen base | verified by current package suite |
| B7 | `if` line 80 | `if !ok \|\| !localReferenceExists(document, ref) {` entered and complementary path | TestEveryOpenAPILocalReferenceResolves | pre-existing regression at frozen base | verified by current package suite |
| B8 | `case` line 86 | `case []any:` entered and complementary path | TestEveryOpenAPILocalReferenceResolves | pre-existing regression at frozen base | verified by current package suite |
| B9 | `range` line 87 | `for _, child := range typed {` entered and complementary path | TestEveryOpenAPILocalReferenceResolves | pre-existing regression at frozen base | verified by current package suite |
