# Branch Test Map: `parseSignedProtectionCapability`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | caller cannot supply raw path/owner/digest/time | compile-time API surface + zero verifier test | raw exported path exists | pass after remediation |
| B2 | canonical signed input parses but is not authority | signed happy path | pass | pass regression |
| B3 | hard revoke after parse | parse-before-revoke test | copied key remains active | pass after remediation |
| B4 | payload base64 is noncanonical or wrong size | envelope encoding table | parser accepted alternate encodings | rejected |
| B5 | payload is not strict matrix JSON | strict JSON table | malformed payload risk | rejected |
| B6 | decoded matrix re-marshals differently | canonical JSON table | alternate signed forms accepted | rejected |
| B7 | matrix semantic validation fails | canonical UTC/sort/account tables | aliases/order variants accepted | rejected |
| B8 | current trust root cannot be safely loaded | root filesystem/digest table | caller root input | rejected |
| B9 | envelope key ID absent from current root | unknown-key test | copied/arbitrary key possible | rejected |
| B10 | signature fails under pinned key | signature tamper table | unsigned/raw matrix path existed | rejected |
