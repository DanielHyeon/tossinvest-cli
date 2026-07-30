## 1. Evidence and RED

- [x] 1.1 Record CodeGraph definition/caller/impact evidence for official account discovery and engine restart recovery
- [x] 1.2 Resolve proposal-freeze review findings and complete pre-edit Function Logic Maps for every modified existing function
- [x] 1.3 Add RED official-client regressions for one-request priming, positive explicit sequence preservation, negative explicit refusal, empty/zero/negative/error/cancel paths, lazy completion, concurrent scoped first use, public-first/scoped contention, and mutex release under `-race`
- [x] 1.4 Add RED actual `NewContext → Context.Recovery → Recovery.Run` regressions proving one account discovery, identical headers, incomplete-first-record and explicit-sequence-mismatch refusal, and genuine snapshot-endpoint 429 fail-closed behavior

## 2. Implementation

- [x] 2.1 Refactor official account discovery so `Accounts` and lazy resolution share the existing account mutex and cache
- [x] 2.2 Preserve explicit positive `WithAccountSeq` for general clients, reject nonpositive headers and engine sequence mismatches, preserve lazy-only callers and error classification, and enforce strict same-record accountRef/header selection
- [x] 2.3 Confirm no mutation retry, partial snapshot acceptance, Guardian change, or gate relaxation was introduced

## 3. Verification and Delivery

- [x] 3.1 Run focused tests and race tests for `internal/official`, engine wiring, and actual CLI assembly
- [x] 3.2 Refresh post-edit Function Logic Maps, sync CodeGraph hard evidence, attempt and record advisory CodeGraphContext status, and pass strict OpenSpec validation
- [x] 3.3 Obtain independent implementation review and run the completion gate `make gate CHANGE=prime-engine-account-seq`
- [x] 3.4 Retain the investigation memory and fix the post-gate delivery plan to stage a new sibling candidate without installing or restarting it
