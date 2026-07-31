## 1. Contract and Evidence

- [x] 1.1 Register STORY-TOS-036, capture the implementation base, and record CodeGraph, CodeGraphContext, and current-HEAD evidence
- [x] 1.2 Complete the `remoteRuntime.security` and `Console.render` Function Logic Maps, Branch Test Maps, risk reports, Pre-Edit Gate, and proposal-freeze review

## 2. TDD Implementation

- [x] 2.1 Add RED regression tests requiring exact `same-origin` remote/rendered headers and shared/restart meta policies; add a full mutation-gate regression proving explicit `Origin: null` with canonical TLS/Host and valid CSRF is origin-refused before handler invocation
- [x] 2.2 Change only the two response header and two HTML meta policy values and make the focused console suite GREEN
- [x] 2.3 Run a real Chrome invalid-CSRF form probe proving canonical origin acceptance without invoking the restart handler

## 3. Verification

- [x] 3.1 Run focused, race, full test, vet, strict OpenSpec, SDD sync/check, and change gate verification
- [x] 3.2 Record an independent implementation/security review with no unresolved P0/P1 findings
