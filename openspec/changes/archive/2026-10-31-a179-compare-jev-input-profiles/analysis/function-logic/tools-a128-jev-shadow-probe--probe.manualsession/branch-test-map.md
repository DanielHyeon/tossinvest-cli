# Branch Test Map: `probe.manualSession`

Relevant package test surfaces: TestManualSessionExperimentCanRestart, TestManualSessionBypassesCalendarKR. These tests cover named scenarios, not every AST branch; rows below identify unproven individual branch coverage rather than asserting it. Existing socket-backed tests could not execute in this sandbox.

| Branch | Source branch | Individual branch coverage |
|---|---|---|
| B1 | if at 234:2: `if err != nil {` | Not individually established |
| B2 | if at 244:2: `if p.cfg.InputProfile != "" && p.cfg.InputProfile != "current" {` | Not individually established |
| B3 | if at 247:2: `if err := p.ledger.gap(row); err != nil {` | Not individually established |
