# Function Logic Map: `validateExitEventArmSuppression`

- Source: `internal/journal/exit_snapshot.go` (lines 448–524)
- AST evidence: `ast.json` (`source_sha256: 0e376d1ca4f6e29b27308c540088d35ad9725304b6fc5cff20c8c7eed9780524`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **High-risk**

## What it does

읽기 시점에 이벤트의 무장 억제 근거를 검증한다. 개정 2가 두 곳의 리터럴 비교를
`knownArmSuppression(reason)`으로 바꿨다 (L495·L515). 쓰기 경로가 새 사유를 허용하는데
읽기 경로가 모르면, 그 행을 담은 `ExitEvents` 호출 전체가 `ErrExitSnapshotCorrupt`로 죽는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `event.Evaluation` 3-튜플 | saved(nil 가능)/recomputed/effective | 저장된 이벤트 행 | legacy(전부 nil)면 근거 없이만 통과 |
| `event.ArmSuppressedReason` | 빈 문자열 또는 알려진 사유 | 행 | 미지면 손상 |
| `event.Action`, `event.ProposedIntentID` | 동시에 있거나 동시에 없어야 한다 | 행 | 부분적이면 손상 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Required test |
|---|---|---|---|
| B1 | (454) `if` — if saved == nil && recomputed == nil && effective == nil && source == "" { | 본문 참조 | 아래 Branch Test Map |
| B2 | (455) `if` — if reason != "" { | 본문 참조 | 아래 Branch Test Map |
| B3 | (461) `if` — if recomputed == nil \|\| effective == nil { | 본문 참조 | 아래 Branch Test Map |
| B4 | (464) `if` — if recomputed.Line.PositionID != event.PositionID \|\| effective.Line.PositionID != event.PositionID \|\| | 본문 참조 | 아래 Branch Test Map |
| B5 | (469) `switch` — switch source { | 본문 참조 | 아래 Branch Test Map |
| B6 | (470) `case` — case EffectiveSourceRecomputed: | 본문 참조 | 아래 Branch Test Map |
| B7 | (471) `if` — if !reflect.DeepEqual(*effective, *recomputed) { | 본문 참조 | 아래 Branch Test Map |
| B8 | (476) `case` — case EffectiveSourceSaved: | 본문 참조 | 아래 Branch Test Map |
| B9 | (477) `if` — if saved == nil \|\| !reflect.DeepEqual(*effective, *saved) { | 본문 참조 | 아래 Branch Test Map |
| B10 | (482) `case` — default: | 본문 참조 | 아래 Branch Test Map |
| B11 | (486) `if` — if saved != nil { | 본문 참조 | 아래 Branch Test Map |
| B12 | (491) `if` — if err != nil \|\| selectedSource != expectedSource \|\| selected != effective.Line { | 본문 참조 | 아래 Branch Test Map |
| B13 | (495) `if` — if reason != "" && !knownArmSuppression(reason) { | 본문 참조 | 아래 Branch Test Map |
| B14 | (499) `if` — if (event.Action == "") != (event.ProposedIntentID == "") { | 본문 참조 | 아래 Branch Test Map |
| B15 | (502) `if` — if source == EffectiveSourceSaved { | 본문 참조 | 아래 Branch Test Map |
| B16 | (503) `if` — if armed \|\| reason != "" { | 본문 참조 | 아래 Branch Test Map |
| B17 | (508) `if` — if recomputed.Line.Orderable && armed { | 본문 참조 | 아래 Branch Test Map |
| B18 | (509) `if` — if reason != "" \|\| event.Action != string(recomputed.Line.Action) { | 본문 참조 | 아래 Branch Test Map |
| B19 | (515) `if` — if recomputed.Line.Orderable && !armed && !knownArmSuppression(reason) { | 본문 참조 | 아래 Branch Test Map |
| B20 | (519) `if` — if !recomputed.Line.Orderable && (armed \|\| reason != "") { | 본문 참조 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `reflect.DeepEqual` | effective가 선언한 출처와 일치하는지 | 순수 | AST `calls` L471·L477 |
| `exitpolicy.SelectRecoverySnapshot` | 복구 선택을 재현해 대조 | 오류·불일치면 손상 | AST `calls` L490 |
| `knownArmSuppression` | 쓰기 경로와 같은 허용 목록 | — | AST `calls` L495·L515 |

## State mutations and fallbacks

- 없음. 검증만.

## Safety conclusion

- Safe edit boundary: 허용 목록 호출 두 곳. 읽기와 쓰기는 반드시 같은 목록이어야 한다.
- High-risk impact: yes — 여기서 손상 판정이 나면 이벤트 조회 전체가 실패한다.
