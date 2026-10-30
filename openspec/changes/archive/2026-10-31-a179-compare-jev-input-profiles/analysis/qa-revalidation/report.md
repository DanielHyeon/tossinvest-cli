## Exploratory QA and Verification Results

| Field | Value |
|---|---|
| Date / branch / revision | 2026-10-31 KST / feat/p1-a179-input-profiles / e32606d4b plus staged source hashes in /tmp/a179-source-hashes.txt |
| Caller / depth | review; 5-minute / 12-probe smoke, owned /tmp writes only |
| Surfaces | CLI default report and legacy-output rejection; targeted tests cover adjacent cycle and recovery |
| Runtime | Locally built Go CLI, gstack QA evidence and deadline helpers |
| Isolation | Synthetic /tmp/a179-cli-demo ledger; no real credentials, sockets or external requests |
| Budget / stop reason | 3 guarded captures, 45 ms aggregate; scoped success and riskiest edge completed |

### Contract outcomes

| Contract | Evidence | Expected → observed | Outcome |
|---|---|---|---|
| Default report selects comparison ledger | capture 001 | exit 0, two separate profile lots, paired J2 delta +0.02 and one crossing | pass |
| Explicit compare output cannot target legacy path | [checkpoint 002](exploration-002.json), capture 002 | exit 1; output isolation error before credential check | pass |
| Rejected output creates no ledger | [checkpoint 003](exploration-003.json), capture 003 | exit 0 from test ! -e; directory absent | pass |

### Findings

No product defect observed in the scoped smoke. Existing legacy socket-backed regression tests remain blocked by loopback EPERM; this is a setup limitation, not a passing test.

### Discoveries and permanent tests

- [checkpoint 002](exploration-002.json) records the report observation leading to output-guard probe.
- [checkpoint 003](exploration-003.json) records the rejection leading to no-effect probe.
- Existing staged profile tests cover corresponding behavior; this review added no tests during QA.

### Coverage limits and cleanup

No real broker/model requests, timer changes or operating ledger writes. Existing selected in-memory HTTP tests and targeted Go race/Go vet/Python results are recorded in change analysis/verification.md. Full package suite, production endpoints and full repository gates were not run. Owned /tmp artifacts remain for review. Source and test hashes match the final candidate; subsequent code changes would invalidate affected evidence.
