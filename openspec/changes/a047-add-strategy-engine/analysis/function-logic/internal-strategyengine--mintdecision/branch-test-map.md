# Branch Test Map: `mintDecision`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | provenance/candidate/source/config binding mismatch | direct forged-record provenance row | indirect only | pass |
| B2 | candidate/session/bar/state/position clock or close-minus-45m cutoff mismatch | direct forged-record clock row plus boundary tests | indirect only | pass |
| B3 | iterate all required positive evidence | first/last positive-field forged rows | missing | pass |
| B4 | required evidence malformed or nonpositive | first/last positive-field forged rows | missing | pass |
| B5 | tangled score malformed or below 0.35 | direct tangled forged row | indirect lane gate only | pass |
| B6 | iterate expansion and HVN optional evidence | first/second optional forged rows | missing | pass |
| B7 | optional evidence present/absent | forged optional rows plus absent-optional success | indirect only | pass |
| B8 | present optional evidence malformed | first/second optional forged rows | missing | pass |
| B9 | negative live-entry delta is normalized | direct negative-drift valid remint | indirect lane boundary only | pass |
| B10 | close/stop/target/RR/drift recomputation mismatch | direct forged stop-price row | synthetic success only | pass |
| B11 | unobserved live price differs from close fallback | direct forged fallback row plus valid fallback remint | indirect only | pass |
| B12 | HVN evidence present/absent | HVN-forged row plus absent-optional success | indirect only | pass |
| B13 | present HVN distance below LVN space | direct forged HVN-below-LVN row | lane gate only | pass |
| B14 | accept reason order differs | direct forged reason-order row | success assertion only | pass |
| B15 | canonical identity calculation fails or differs | direct forged identity row | success assertion only | pass |
| Invariant | valid record returns one opaque Decision | base remint and negative/optional success tests | existing | pass |
