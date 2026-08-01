# Branch Test Map: `positionPolicyRequestHandler`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | non-POST method is rejected | transport handler contract | yes | yes |
| B2 | parsed media type must be exactly `application/json` | `TestPositionPolicyControlStrictlyBoundsAndFramesJSON/non_JSON_media_type` | yes | yes |
| B3 | body read failure is rejected | bounded reader contract | yes | yes |
| B4 | body larger than 16 KiB is rejected before command dispatch | `TestPositionPolicyControlStrictlyBoundsAndFramesJSON/oversize_body` | yes | yes |
| B5 | malformed or unknown-field JSON is rejected | `TestPositionPolicyControlStrictlyBoundsAndFramesJSON/client_supplied_mutation_scope` | yes | yes |
| B6 | a second JSON value is rejected | `TestPositionPolicyControlStrictlyBoundsAndFramesJSON/trailing_JSON_value` | yes | yes |
| B7 | command error is mapped; success returns one JSON response | `TestCompetingControlClientsPreserveOneCASWinner` | yes | yes |
