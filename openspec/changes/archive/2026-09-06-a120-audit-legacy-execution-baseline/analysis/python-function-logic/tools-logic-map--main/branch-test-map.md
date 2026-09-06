# Branch Test Map: `main`

| Branch | Scenario | Test | Expected |
|---|---|---|---|
| B1 | checker diagnostic | new CLI distinction fixture | exit 1 and diagnostic text |
| B2 | multiple checker diagnostics | new CLI distinction fixture | each diagnostic is printed |
| success | ordinary / validated adoption | new CLI distinction fixture | exit 0; ordinary generic or adoption-exception label |
