# Branch Test Map: `stopProvenance`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | candidate stop selected, candidate provenance retained | continuation accepted exact-terms tests | covered | covered |
| B2 | unsupported quote scale refuses provenance | continuation unsupported-currency tests | covered | covered |
| fallthrough | saved stop selected with forged/zero seal | `TestCallerForgedSavedStopProvenanceFailsClosed` | yes | yes |
