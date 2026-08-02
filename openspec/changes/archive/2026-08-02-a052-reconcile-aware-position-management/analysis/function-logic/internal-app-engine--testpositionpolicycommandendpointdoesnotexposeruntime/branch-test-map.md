# Branch Test Map: `TestPositionPolicyCommandEndpointDoesNotExposeRuntime`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 163 | `if err != nil {` true/entered and complementary path | TestPositionPolicyCommandEndpointDoesNotExposeRuntime | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B2 | `if` line 168 | `if err != nil {` true/entered and complementary path | TestPositionPolicyCommandEndpointDoesNotExposeRuntime | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B3 | `if` line 172 | `if err := json.Unmarshal(body, &descriptor); err != nil {` true/entered and complementary path | TestPositionPolicyCommandEndpointDoesNotExposeRuntime | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B4 | `if` line 176 | `if err != nil {` true/entered and complementary path | TestPositionPolicyCommandEndpointDoesNotExposeRuntime | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B5 | `if` line 181 | `if err != nil {` true/entered and complementary path | TestPositionPolicyCommandEndpointDoesNotExposeRuntime | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B6 | `if` line 185 | `if response.StatusCode != http.StatusNotFound \|\| commands.runtimeCalls != 0 {` true/entered and complementary path | TestPositionPolicyCommandEndpointDoesNotExposeRuntime | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
