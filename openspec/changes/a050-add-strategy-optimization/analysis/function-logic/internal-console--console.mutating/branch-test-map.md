# Branch Test Map: `Console.mutating`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | non-POST is refused before handler | state-changing route method suite | baseline contract | PASS |
| B2 | remote cross-origin POST is refused | remote mutation-origin suite | baseline contract | PASS |
| B3 | multipart/malformed content type is refused before polluted values are read | `TestOptimizationRejectsMultipartBeforeReadingPollutedValues` and shared mutation tests | multipart parsed before exact-set validation | PASS |
| B4 | route body cap wraps request before parse | `TestOptimizationRequestBodyIsBounded` | optimization route lacked cap | PASS |
| B5 | malformed/oversized form parse fails without handler call | shared mutation parse tests | baseline plus a050 bound | PASS |
| B6 | `MaxBytesError` maps to 413 | `TestOptimizationRequestBodyIsBounded` | optimization route lacked cap | PASS |
| B7 | missing/mismatched CSRF is refused | `TestOptimizationSaveRequiresSessionAndCSRF` | baseline contract | PASS |
