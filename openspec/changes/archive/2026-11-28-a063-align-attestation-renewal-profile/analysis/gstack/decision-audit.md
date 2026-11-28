# a063 autoplan decision audit

| ID | Phase | Decision | Class | Principle | Rationale / rejected alternative |
|---|---|---|---|---|---|
| D1 | CEO | Treat August expiry as historical | Mechanical | Explicit | Current sanitized expiry is October; no false incident claim |
| D2 | CEO | Unit + bounded status + existing console | Mechanical | Completeness/DRY | Unit-only misses visibility; journal integration adds coupling |
| D3 | CEO | Keep deployment approval and actual days separate | Mechanical | Explicit | Synthetic tests cannot prove live acceptance |
| D4 | Design | Separate current usability, renewal health, denial | Mechanical | Explicit | A renewal failure must not imply running-engine stop |
| D5 | Eng | Full attestation filename + suffix | Mechanical | Completeness | Fixed dirname sibling collides for different basenames |
| D6 | Eng | Strict 4096-byte/16-code reader, future invalid | Mechanical | Completeness | Unknown data cannot become success |
| D7 | Eng | Current attestation expiry authority | Mechanical | Explicit | Stale/mismatching issued status cannot conceal expiry |
| D8 | Eng | Preserve last good file on diagnostic failure | Mechanical | Pragmatic | Diagnostic fault must not erase valid evidence |
| D9 | Eng | Retain immutable original base | Mechanical | Explicit | Re-freeze suggestion withdrawn; no checker bypass |
| D10 | DX | Document binary-before-template and profile paths | Mechanical | Explicit | Old binary flag rejection and custom path confusion |

No taste decisions or user scope challenges remain. No outside-model agreement is claimed. Manager owns core OpenSpec decisions; this audit records accepted review dispositions without altering core files.
