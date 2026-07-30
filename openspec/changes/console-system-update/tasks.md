## 1. Contract and analysis

- [x] 1.1 Validate the change strictly, capture the implementation base commit, and record proposal-freeze review decisions in `review.md`
- [x] 1.2 Scaffold pre-edit Function Logic Map and Branch Test Map evidence for every existing console/cmd function changed

## 2. Candidate inspection and replacement

- [x] 2.1 Add RED tests for absent candidate, no-follow symlink/non-regular/non-executable refusal, wrong module/platform refusal, metadata/hash reporting, changed-after-review refusal, and no candidate execution
- [x] 2.2 Implement `internal/localupdate` descriptor-bound candidate preparation and inspection at `<self>.candidate`
- [x] 2.3 Add RED tests for current-target drift, prepared/rollback/current sync order, no absent-current window, successful replacement, candidate-rename failure, directory-sync failure restoration, and unchanged bytes on refusal
- [x] 2.4 Implement Unix atomic replacement with durable rollback while current remains intact, plus an explicit unsupported non-Unix implementation
- [x] 2.5 Reject same-module executables whose Go main-package path is not exactly `cmd/tossctl`, with a real `tools/boundarymap` regression

## 3. Console and CLI wiring

- [x] 3.1 Add RED console tests for settings rendering, reviewed-hash binding, session/CSRF protection, ignored path fields, real engine-lock/verification evidence/rerelaunch/target-drift refusals, start-route serialization, success interstitial, logging, and failed-install no-relaunch behavior
- [x] 3.2 Add the system-update view and authenticated install handler, sharing an in-process exclusion with engine/verification starts and reusing same-port restart only after successful replacement
- [x] 3.3 Inject the fixed installer, real engine lock, and strict verification activity reader from `cmd/tossctl`; update static route/capability guards and command integration tests
- [x] 3.4 Add `make stage-local-update` and documentation that stages but never overwrites or restarts
- [x] 3.5 Close the external verification TOCTOU by making standalone and console verification hold the same real flock as engine/update before account or broker work

## 4. Verification and evidence

- [x] 4.1 Refresh post-edit AST/Function Logic Maps and pass `check_analysis.py`
- [x] 4.2 Run targeted package tests, full `make test`, `make vet`, and strict OpenSpec validation
- [x] 4.3 Run `make sdd-sync`, `make sdd-check`, and `make gate CHANGE=console-system-update`; record independent implementation review
