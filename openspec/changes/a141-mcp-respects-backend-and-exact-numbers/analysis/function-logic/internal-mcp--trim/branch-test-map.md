# Branch Test Map: `trim`

| Branch | Scenario | Test / evidence | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | force progress when proportional keep >= n | TestToolResultCapsOversizedPayload exercises cap behavior; exact hit not measured | not claimed | package focused PASS |
| B2 | clamp negative keep to zero | unchanged expression; static AST comparison | not claimed | no focused branch-hit claim |
| B3 | prior marker present/absent | TestTrimAcceptsExactAndMalformedOmissionMarkers first/repeated trims | pre-existing numeric marker regression | Terra focused PASS |
| B4 | valid prior count versus malformed/overflow count | TestTrimAcceptsExactAndMalformedOmissionMarkers | exact marker accumulation regression | Terra focused PASS |
