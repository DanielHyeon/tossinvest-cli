# Function Logic Map (편집 전): `validateLane`

- Source: `internal/strategyprojection/lanes.go`
- Source SHA-256: `fb4b35f5447efe634788c0e4ba572808ddc5a703777739b3a9e143c9e7c22bd7`
- Signature: `validateLane(params=1, results=1)`
- Source range: `274:1`–`333:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD 9e5f3ccf).

## Inputs and invariants

- 편집 계획: 브리프 §6 · §7: runtime 정확 일치 {UNOBSERVED} → {UNOBSERVED, SHADOW}; SHADOW ⇒ desired=OFF ∧ effective=OFF ∧ health≠nil ∧ cycleGeneration>0; shadowOutcome≠null ⇔ runtime=SHADOW.

## Branches and early returns

- Exact AST return nodes: `276:3`, `281:3`, `290:4`, `292:3`, `295:3`, `298:3`, `303:3`, `307:3`, `312:4`, `314:3`, `317:3`, `320:3`, `323:3`, `327:3`, `330:3`, `332:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 275:2 | `if !validState(lane.Desired) // !validState(lane.Effective) // lane.Runtime != LaneRuntimeUnobserved {` |
| B2 | if | 280:2 | `if (lane.Refusal != nil) != refused // lane.Refusal != nil && !member(*lane.Refusal, ArbitrationRefusals()) {` |
| B3 | if | 283:2 | `if lane.Health == nil {` |
| B4 | if | 285:3 | `if lane.Desired != StateOff // lane.Effective != StateOff // lane.ConsecutiveFailures != 0 // lane.LatchRevision != 0 //` |
| B5 | if | 294:2 | `if !member(string(*lane.Health), LaneHealths()) {` |
| B6 | if | 297:2 | `if lane.PolicyVersion == nil // !validIdentity(*lane.PolicyVersion) // lane.CycleDeadlineMS == nil // *lane.CycleDeadlineMS <= 0 {` |
| B7 | if | 300:2 | `if lane.FirstFailure != nil && !validText(*lane.FirstFailure) // lane.Pending < 0 //` |
| B8 | if | 306:2 | `if (lane.CycleGeneration == 0) != (lane.Trigger == nil) {` |
| B9 | if | 309:2 | `if lane.Trigger == nil {` |
| B10 | if | 310:3 | `if lane.Start != nil // lane.Outcome != nil // lane.Abnormal // lane.SnapshotDigest != nil // lane.EvidenceDigest != nil //` |
| B11 | if | 316:2 | `if !member(string(*lane.Trigger), LaneTriggers()) {` |
| B12 | if | 319:2 | `if (*lane.Trigger == LaneTriggerEnqueued) != (lane.Start != nil) {` |
| B13 | if | 322:2 | `if lane.Start != nil && !member(string(*lane.Start), LaneStarts()) {` |
| B14 | if | 326:2 | `if !admitted && (lane.Outcome != nil // lane.Abnormal) {` |
| B15 | if | 329:2 | `if lane.Outcome != nil && !member(string(*lane.Outcome), LaneOutcomes()) {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `validState` | 275:6 |
| `validState` | 275:35 |
| `errors.New` | 276:10 |
| `member` | 280:65 |
| `ArbitrationRefusals` | 280:87 |
| `errors.New` | 281:10 |
| `errors.New` | 290:11 |
| `member` | 294:6 |
| `string` | 294:13 |
| `LaneHealths` | 294:35 |
| `errors.New` | 295:10 |
| `validIdentity` | 297:35 |
| `errors.New` | 298:10 |
| `validText` | 300:34 |
| `lane.NextDueAt.IsZero` | 301:28 |
| `lane.RestartNotBefore.IsZero` | 301:87 |
| `validIdentity` | 302:34 |
| `validIdentity` | 302:104 |
| `errors.New` | 303:10 |
| `errors.New` | 307:10 |
| `errors.New` | 312:11 |
| `member` | 316:6 |
| `string` | 316:13 |
| `LaneTriggers` | 316:36 |
| `errors.New` | 317:10 |
| `errors.New` | 320:10 |
| `member` | 322:27 |
| `string` | 322:34 |
| `LaneStarts` | 322:55 |
| `errors.New` | 323:10 |
| `errors.New` | 327:10 |
| `member` | 329:29 |
| `string` | 329:36 |
| `LaneOutcomes` | 329:59 |
| `errors.New` | 330:10 |

## Safety conclusion

- 외부 경계 검증기 — 기존 거절 규칙 무변경, 새 교차 규칙만 추가(거부되는 정상 입력 표 시험).
