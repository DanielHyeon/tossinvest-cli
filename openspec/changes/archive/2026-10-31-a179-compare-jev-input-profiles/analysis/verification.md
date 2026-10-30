# a179 implementation verification

Base `e32606d4bd44de3142f3c732ab86ecee913e3ab0`; working branch `feat/p1-a179-input-profiles`. User changed the default during implementation: automatic compare uses a dedicated sibling ledger; explicit current uses the legacy path. No scheduler, configuration, actual API, account, or source working-tree changes were made.

## Executed checks

- `GOCACHE=/tmp/tossos-go-cache GOPROXY=off go test -v -race ./tools/a128-jev-shadow-probe -run '^(TestDefaultAndExplicitCurrentProduceSameRows|TestCompareCycleSharesCollectionAndLabels|TestBothProfileFailuresKeepSourceAndAreCounted|TestOneProfileFailureKeepsOtherResult|TestCompareReportRejectsMismatchesAndDuplicates|TestCompareReportDistinguishesLabelsAndInvalidRows|TestComparisonRejectsTamperedStoredState|TestInitialStateProjectionIsIndependent|TestProfileOutputGuard|TestProfileOutputGuardRejectsSymlinkAlias|TestProfileOutputGuardRejectsDanglingSymlink|TestProfileOutputGuardRejectsLegacyLedger|TestProfileFlagsRejectedBeforeCredentials|TestSelectionLotSeparatesProfiles|TestPreS4LedgerStillReportsTheSameAndAddsTimeOfDay|TestAutomaticProfileOutputRouting|TestManualSessionExperimentCanRestart|TestRunRefusesToStartWithoutTheAPIKey|TestReportSeparatesSelectionLots|TestReportSplitsLotsByStateRev)$' -count=1` — PASS, 20 named Go tests, race enabled on the preceding candidate. Raw output `/tmp/a179-go-final.log`. Subsequent invalid-row accounting change: the three affected Go report tests passed again with `-race`; unrelated behavior was unchanged.
- `python3 -m unittest discover -s tools/a128-jev-shadow-probe -p test_profile_eval.py` — PASS, 3 tests. Output `/tmp/a179-python-final.log`. Covers invalid/tampered state, later failed-judgment price snapshot, fallback boundaries, latest bar replacement, KR later warning and US differing first flagged observation.
- `GOCACHE=/tmp/tossos-go-cache GOPROXY=off go vet ./tools/a128-jev-shadow-probe` — PASS; `/tmp/a179-vet-final.log` empty.
- `node /tmp/tossos-openspec-1.4.1/node_modules/@fission-ai/openspec/bin/openspec.js validate a179-compare-jev-input-profiles --strict --no-interactive` — valid.
- `python3 tools/pm/generate_master_tracker.py --check` — hierarchy and generated trackers current.
- `GOCACHE=/tmp/tossos-go-cache GOPROXY=off python3 tools/logic-map/check_analysis.py --change a179-compare-jev-input-profiles` — structurally complete for six modified functions. `/tmp/a179-logic-final.log`. AST enumeration and branch tables do not claim every individual branch was exercised.
- `git diff --cached --check` — PASS.
- Synthetic default `report` CLI read `/tmp/a179-cli-demo/.../shadow-ledger-compare/ledger.jsonl`: one paired J2, delta +0.020000 and one threshold crossing; `run --help` reports compare default without contacting an API.

## Review and limits

Gstack review skill preamble and checklist applied. `GSTACK_STATE_ROOT=/tmp/a179-gstack` enabled the documented scratch state after the default root failed EROFS. Final local gate inventory `/tmp/a179-gstack-gate-final.log`: 3 test paths listed, 0 eligible removals/tagged hunks, 0 unread; other paths unmatched by gate patterns. No existing test or gate was relaxed. The user-selected independent Terra reviewer identified cohort ID collisions, digest validation, invalid-pair/label accounting and US timing ambiguity; fixes and Manager's US policy clarification were applied. The initial-only log summary bug and manual-session restart bug were fixed; the latter had a failing regression before the fix. The cycle integration itself was written before its new regression test, so RED was not established for that part. A fresh independent Terra native adversarial review completed under the user-selected model (token 48dee55b-eddb-40c9-a25f-3e3e600641ac), with no reproducible material findings. Scoped functional QA completed and was repeated after the invalid-row fix: `qa-revalidation/report.md` with captured synthetic report and output-guard evidence; `qa/` retains the superseded earlier run. A full gstack specialist workflow is not claimed; the earlier bound record was explicitly `completed:false`. Invalid rows missing cohort/source now count as invalid in both reporting and evaluation; Manager delta review approved the result.

A normal existing socket-backed package test (`TestLedgerRowSchema`) failed before test execution because `httptest.NewServer` could not bind loopback: `socket: operation not permitted`. Existing socket tests were not modified or claimed passing. New end-to-end HTTP fixture tests use in-memory RoundTrippers, not real services. Whole-repository tests, full gates, real model/broker API calls, live measurements, The change was archived as `2026-10-31-a179-compare-jev-input-profiles`, and the new capability passed strict validation; PM and agent-config checks passed. Commit and push are pending at this checkpoint; no deployment occurred. Source CodeGraph context was available but precise checkout graph freshness and CodeGraphContext/GBrain freshness were not established.

## Source hashes (SHA-256)

```text
bc78735ffc509faaa6ba64d48f5b1723e2228df3655b18e37f7ed1e9e02d500a  tools/a128-jev-shadow-probe/collect.go
c1c4a396b5b6a629aff7dbc3dfe685ad6a10df48fad0f8bc080b6b95c735be25  tools/a128-jev-shadow-probe/ledger.go
07575a44c2d08ce22daeef8e6b2f891efe645dc992b74e0cbcaf5a5c97334d37  tools/a128-jev-shadow-probe/main.go
2dfef4e992820034273cc6ec88e12dff3d7b07248dd074c52fa9a4eee31aa8de  tools/a128-jev-shadow-probe/selection.go
a28b2057862d1f60c8b5eb84c9e9e507a12449cb956b02cf4140acd3ef5de8a2  tools/a128-jev-shadow-probe/profile.go
6ffe3b0a679aaec17a83a213cbf444dc612b2eac606ad89c4dfc5ba307931083  tools/a128-jev-shadow-probe/compare_report.go
03021d80cd75bf050cdb4f6cce61769d59ea22819910b687b2a9068b41636c44  tools/a128-jev-shadow-probe/profile_cycle_test.go
034c7a1e6526f0a40aad72efc9223963aae67dc7425378c3ce3746083a83317d  tools/a128-jev-shadow-probe/profile_test.go
0d351dd267ab66a0f35d28aa61028cc08dae88138cd3ceb2c95d9842c8eeef02  tools/a128-jev-shadow-probe/profile_eval.py
e4e2690a6e1983dc11fea5543812e7eb18b5b974718ed4c3c64fd2841616c099  tools/a128-jev-shadow-probe/test_profile_eval.py
```
