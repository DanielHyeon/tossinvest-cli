# Function Logic Map: `validateLane`

- Source: `internal/strategyprojection/lanes.go`
- Source SHA-256: `9b6b15b0ebe8e74fcd582a8a6d6abed626706542ba9706e5ba551a3c302315a4`
- Signature: `validateLane(params=1, results=1)`
- Source range: `298:1`–`366:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 거부 표 시험 `TestValidateRefusesShadowOutsideTheOneCrossRule`(변이 V01 · V02).

## Branches and early returns

- Exact AST return nodes: `300:3, 306:3, 309:3, 314:3, 323:4, 325:3, 328:3, 331:3, 336:3, 340:3, 345:4, 347:3, 350:3, 353:3, 356:3, 360:3, 363:3, 365:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 299:2 | desired/effective/runtime 어휘 |
| B2 | if | 305:2 | **(새)** SHADOW ⇒ OFF/OFF ∧ health ∧ cycleGeneration>0 |
| B3 | if | 308:2 | **(새)** shadowOutcome ⇔ SHADOW |
| B4 | if | 313:2 | 거절 코드 짝 |
| B5 | if | 316:2 | 미관측 |
| B6 | if | 318:3 | 미관측 사실 추론 |
| B7 | if | 327:2 | 건강 어휘 |
| B8 | if | 330:2 | 정책 |
| B9 | if | 333:2 | 정규 값 |
| B10 | if | 339:2 | 물결 ⇔ 트리거 |
| B11 | if | 342:2 | 트리거 없음 |
| B12 | if | 343:3 | 트리거 없는 사실 |
| B13 | if | 349:2 | 트리거 어휘 |
| B14 | if | 352:2 | 시작 ⇔ ENQUEUED |
| B15 | if | 355:2 | 시작 어휘 |
| B16 | if | 359:2 | 결과 ⇔ ADMITTED |
| B17 | if | 362:2 | 결과 어휘 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `validState` | 299:6 |
| `validState` | 299:35 |
| `member` | 299:66 |
| `string` | 299:73 |
| `LaneRuntimes` | 299:95 |
| `errors.New` | 300:10 |
| `errors.New` | 306:10 |
| `member` | 308:76 |
| `string` | 308:83 |
| `LaneShadowOutcomes` | 308:112 |
| `errors.New` | 309:10 |
| `member` | 313:65 |
| `ArbitrationRefusals` | 313:87 |
| `errors.New` | 314:10 |
| `errors.New` | 323:11 |
| `member` | 327:6 |
| `string` | 327:13 |
| `LaneHealths` | 327:35 |
| `errors.New` | 328:10 |
| `validIdentity` | 330:35 |
| `errors.New` | 331:10 |
| `validText` | 333:34 |
| `lane.NextDueAt.IsZero` | 334:28 |
| `lane.RestartNotBefore.IsZero` | 334:87 |
| `validIdentity` | 335:34 |
| `validIdentity` | 335:104 |
| `errors.New` | 336:10 |
| `errors.New` | 340:10 |
| `errors.New` | 345:11 |
| `member` | 349:6 |
| `string` | 349:13 |
| `LaneTriggers` | 349:36 |
| `errors.New` | 350:10 |
| `errors.New` | 353:10 |
| `member` | 355:27 |
| `string` | 355:34 |
| `LaneStarts` | 355:55 |
| `errors.New` | 356:10 |
| `errors.New` | 360:10 |
| `member` | 362:29 |
| `string` | 362:36 |
| `LaneOutcomes` | 362:59 |
| `errors.New` | 363:10 |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- 외부 경계 검증기 — 기존 거절 규칙 불변.
