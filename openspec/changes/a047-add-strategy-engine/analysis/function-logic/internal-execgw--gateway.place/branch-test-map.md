# Branch Test Map: `Gateway.Place`

| Branch | Scenario | Test | RED observed | GREEN observed |
| --- | --- | --- | --- | --- |
| B1 | preflight rejects unsupported/unprotected entry | `internal/execgw` preflight and protection tests | baseline | baseline |
| B2 | missing/tampered/expired/spent decision, tightened mode, or bad reservation refuses before official call | decision, modegate, reservation-gate, protection tests | baseline | baseline |
| B3 | exact journal-derived key/body submits once; ambiguous result becomes in-doubt | gateway, roundtrip, replay and indoubt tests | baseline | baseline |
| A047 | strategy has no alternate submitter/import path | a047 dependency/static test (to add in RED) | pending | no |
