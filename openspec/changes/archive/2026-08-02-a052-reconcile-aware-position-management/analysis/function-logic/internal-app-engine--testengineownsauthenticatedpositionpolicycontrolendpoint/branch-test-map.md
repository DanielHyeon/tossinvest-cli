# Branch Test Map: `TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 122 | `if err != nil {` entered and complementary path | TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint | pre-existing regression at frozen base | verified by current package suite |
| B2 | `if` line 127 | `if err != nil \|\| info.Mode().Perm()&0o077 != 0 \|\| !info.Mode().IsRegular() {` entered and complementary path | TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint | pre-existing regression at frozen base | verified by current package suite |
| B3 | `if` line 131 | `if err != nil {` entered and complementary path | TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint | pre-existing regression at frozen base | verified by current package suite |
| B4 | `if` line 135 | `if err != nil \|\| len(states) != 1 \|\| states[0].Version != 4 {` entered and complementary path | TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint | pre-existing regression at frozen base | verified by current package suite |
| B5 | `if` line 138 | `if err := server.Close(); err != nil {` entered and complementary path | TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint | pre-existing regression at frozen base | verified by current package suite |
| B6 | `if` line 141 | `if _, err := os.Stat(descriptorPath); !errors.Is(err, os.ErrNotExist) {` entered and complementary path | TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint | pre-existing regression at frozen base | verified by current package suite |
| B7 | `if` line 144 | `if entries, err := os.ReadDir(dir); err != nil \|\| len(entries) != 0 {` entered and complementary path | TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint | pre-existing regression at frozen base | verified by current package suite |
