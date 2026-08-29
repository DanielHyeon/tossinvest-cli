# Branch Test Map: `isZeroDecimal`

| Branch | Scenario | Test | RED | GREEN |
|---|---|---|---|---|
| B1 | invalid is not zero | A110 unreadable raw table | preserve | yes |
| Return | zero exact; `0.0000000005` nonzero | A110 tolerance-zero near-zero case | yes | yes |
