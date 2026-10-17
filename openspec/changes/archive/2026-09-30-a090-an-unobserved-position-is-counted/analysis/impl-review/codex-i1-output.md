| ID | Severity | File:line | Evidence | Suggested fix |
|---|---|---|---|---|
| A090-I1 | P2 | `internal/app/engine/a090_unobserved_position_test.go:1052` | R17 inspects persisted outbox rows. Injected alert/notice INSERT failures leave **no corresponding row**, so the failed event’s key/title/body/payload escape inspection. The recorder spy only counts calls. This misses tasks 2.17’s explicit requirement to inspect arguments reaching the recording entrance on failure. | Capture events before delegating to `RecordCritical`; assert account absence on failed initial calls and retries. For mode-commit failure, assert no notice call and an empty notice queue. |
| A090-I2 | P2 | `internal/app/engine/a090_unobserved_position_test.go:680` | The comment claims position disappearance and B3 coverage, but neither position is removed. Recovery at line 693 runs with positions still held. Other B3 tests have no pending notice, leaving tasks 2.3g⑧ unverified. | Queue a failed notice, remove both positions, recover storage, then execute B3. Assert exactly one notice with the original transition key, queue drainage, and no repeated transition. Add a mutation that skips notice retries when the held set is empty. |

**Verdict: PASS-WITH-FIXES.**

No new P0/P1 implementation defect identified. Recording occurs after position judgement; the inspected production path uses the recording entrance without remote delivery or delivery leases. Counting, monotonic bases, escalation latches, and notice retries match the frozen design.

Static review only; tests and mutations were not rerun. The supplied **31/31** mutation ledger does not cover the two gaps above. Manager-accepted loop-cost measurement remains outstanding; this verdict does not establish a latency bound.