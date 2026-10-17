# Pre-edit branch table: `internal-riskbucket--aggregateproductionriskusage` (before a066 5.6.1)

Moved out of `function-logic-map.md` when the map was re-based on the post-edit AST at `b8211926`.

| Branch | Position | Condition (AST source line at `f2decd0a`) | Coverage (pre-edit) |
|---|---|---|---|
| B1 | range at 448:2 | `for _, row := range rows {` | covered |
| B2 | if at 451:3 | `if !filledOK \|\| !heldOK \|\| rowFilled.Sign() < 0 \|\| rowHeld.Sign() < 0 \|\| rowFilled.BitLen() > 256 \|\| rowHeld.BitLen() > 256 \|\|` | covered |
| B3 | if at 459:3 | `if filled.BitLen() > 256 \|\| held.BitLen() > 256 {` | NOT covered |
