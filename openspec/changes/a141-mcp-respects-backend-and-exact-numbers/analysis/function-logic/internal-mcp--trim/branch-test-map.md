# Branch Test Map: `trim`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | No existing marker | Existing cap regression | no | existing green |
| B2 | Existing marker uses `json.Number`; malformed, fractional, negative, or overflowing values stay non-panicking | New repeated-trim and failure-safe regressions | pending | pending |
| B3 | Rounding would keep every real item | Existing cap regression proves at least one removal | no | existing green |
| B4 | Excess larger than size | New nested oversized-item regression | pending | pending |
