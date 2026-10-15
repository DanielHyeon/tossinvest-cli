# Branch Test Map: `toolResult`

| Branch | Scenario | Test / evidence | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | initial marshal failure returns RPC error | unchanged; static AST; no forced-error test claimed | not claimed | no focused branch-hit claim |
| B2 | over-cap versus small payload | TestToolResultKeepsSmallPayload; TestToolResultCapsOversizedPayload | not claimed | Terra focused PASS |
| B3 | exact decoding succeeds for internally marshalled JSON | TestToolResultKeepsExactNumbersAfterTrimming; decoder-failure fallback remains defensive and is not directly tested | UseNumber removal mutant detected | Terra focused PASS |
| B4 | trimmed value re-encodes and replaces text | TestToolResultCapsOversizedPayload; TestToolResultKeepsExactNumbersAfterTrimming; re-encode failure not injected | UseNumber removal mutant detected | Terra focused PASS |
