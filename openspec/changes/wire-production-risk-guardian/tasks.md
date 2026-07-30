## 1. Contract and high-risk analysis

- [x] 1.1 Validate the change strictly, capture the implementation base commit, and record adversarial Eng review decisions in `review.md`
- [x] 1.2 Scaffold and complete pre-edit Function Logic Map, Branch Test Map, and risk-pattern evidence for `engine.NewContext` and every existing high-risk function changed
- [x] 1.3 Add RED policy-transcription tests covering all five bounds, USD currency, risk-budget derivation, invalid/non-finite input, and snapshot equivalence

## 2. Production Guardian construction

- [x] 2.1 Implement the pure automation-gate-to-risk-policy transcription with stable policy version and configured cost model
- [x] 2.2 Change `engine.NewContext` to construct one real Guardian after account/journal/gateway assembly, pass it to the interlock, publish it on success, and close/refuse on construction error
- [x] 2.3 Add engine-package tests for gate-off non-construction, construction failure cleanup, injected-Guardian preservation, and interlock/exit-observer identity

## 3. Real CLI assembly regression

- [x] 3.1 Extract a private production assembly helper with `official.Option` and an unexported options decorator; keep `runEngineRun` on the zero-option/no-decorator wrapper
- [x] 3.2 Add a `tossos_testseams`-tagged RED→GREEN command test using isolated config, real SQLite journal, fixed ext4 FS probe, shipped `UNWIRED` protection state, valid USD gate/attestation, and `httptest` official account endpoints with no Guardian injection
- [x] 3.3 Prove exactly one real Guardian is constructed, issue an isolated reduce-only decision through it, read that decision through `Context.Journal`, construct the exit observer, then close the context and prove the shared journal handle is closed

## 4. Verification and evidence

- [x] 4.1 Refresh post-edit AST/Function Logic Maps and pass `check_analysis.py`
- [x] 4.2 Run targeted tests including `go test -tags=tossos_testseams ./cmd/tossctl`, `go test -race` for affected packages, the existing journal crash suite, full `make test`, `make vet`, and strict OpenSpec validation; record why no new subprocess crash test applies
- [x] 4.3 Run `make sdd-sync`, `make sdd-check`, and `make gate CHANGE=wire-production-risk-guardian`; record independent implementation review and reusable investigation learning
