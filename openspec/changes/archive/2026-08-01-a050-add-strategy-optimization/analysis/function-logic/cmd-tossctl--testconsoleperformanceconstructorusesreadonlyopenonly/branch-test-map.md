# Branch Test Map: `TestConsolePerformanceConstructorUsesReadOnlyOpenOnly`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | production source is readable | this test | guard absent at `948e721` | PASS |
| B2 | production constructor explicitly uses `performance.OpenReadOnly` | this test | guard absent at `948e721` | PASS |
| B3 | every forbidden writer/collector API marker is inspected | this test | guard absent at `948e721` | PASS |
| B4 | no forbidden writer/collector marker appears | this test | guard absent at `948e721` | PASS |
