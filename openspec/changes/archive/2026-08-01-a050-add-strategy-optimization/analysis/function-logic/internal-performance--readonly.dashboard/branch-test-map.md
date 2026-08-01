# Branch Test Map: `ReadOnly.Dashboard`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil receiver rejects read | lifecycle unit coverage | safe reader absent at `948e721` | PASS |
| B2 | closed receiver rejects subsequent read | cmd capability lifecycle test | safe reader absent at `948e721` | PASS |
| B3 | missing/active-WAL snapshot at call time fails closed | WAL/missing read-only tests | safe reader absent at `948e721` | PASS |
| B4 | dashboard query error returns no partial view | dashboard query validation/cancellation coverage | safe reader absent at `948e721` | PASS |
| B5 | ephemeral DB close error wins only after successful query | AST/error-order review; SQLite close fault is not injectible through this narrow type | safe reader absent at `948e721` | defensive branch reviewed |
| B6 | file identity or sidecars changing during query fail closed | immutable change/WAL tests | safe reader absent at `948e721` | PASS |
