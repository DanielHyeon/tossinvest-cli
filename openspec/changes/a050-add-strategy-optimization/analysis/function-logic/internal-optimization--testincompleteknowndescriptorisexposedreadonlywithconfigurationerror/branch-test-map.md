# Branch Test Map: `TestIncompleteKnownDescriptorIsExposedReadOnlyWithConfigurationError`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | stable known malformed field stays visible | this test | registry rejected all malformed metadata at `948e721` | PASS |
| B2 | visible field is read-only, has explicit error, and exposes zero preset options | this test | registry rejected all malformed metadata at `948e721` | PASS |
| B3 | blank stable key is still rejected as invalid registry | this test | explicit identity case absent at `948e721` | PASS |
