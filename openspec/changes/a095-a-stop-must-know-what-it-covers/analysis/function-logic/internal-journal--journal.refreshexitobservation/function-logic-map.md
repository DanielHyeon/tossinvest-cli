# Function Logic Map: `Journal.RefreshExitObservation`

- Source: `internal/journal/exit_observation_refresh.go` (`39`–`155`)
- Qualified: `Journal.RefreshExitObservation`
- AST evidence: `ast.json` (`source_sha256` 08d1b43e2aa49d98…)
- Risk scan: `risk-pattern-report.md`
- 분기 26 · return 24 · 호출 51

**역할.** 판정 없이 관측 증거만 새로 고친다. `baseline_price`를 UPDATE 문에 포함한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `request.Snapshot` | 새 관측 스냅샷 | 호출자 | B23 — 운영 선(보호가 포함)이 다르면 거절 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:41` `if id == "" {` | `fmt.Errorf` | :42 | 예 |
| B2 | if | `:44` `if err := request.Provenance.validate(); err != nil \|\| request.Provenance.zero() {` | `request.Provenance.validate`, `request.Provenance.zero` | — | 예 |
| B3 | if | `:45` `if err != nil {` | `fmt.Errorf` | :46, :48 | 예 |
| B4 | if | `:50` `if !sameExitDecisionProvenance(request.Provenance, ExitDecisionProvenance{` | `fmt.Errorf`, `sameExitDecisionProvenance` | :54 | 예 |
| B5 | if | `:56` `if request.ObservedAt.IsZero() {` | `fmt.Errorf`, `request.ObservedAt.IsZero`, `strings.TrimSpace` | :57 | 예 |
| B6 | if | `:60` `if _, _, ok := observationSourceOrder(source); !ok {` | `fmt.Errorf`, `observationSourceOrder` | :61 | 예 |
| B7 | if | `:63` `if request.Snapshot.Orderable \|\| !request.Snapshot.ExecutableProposal().Zero() {` | `Format`, `Zero`, `fmt.Errorf`, `request.ObservedAt.UTC`, `request.Snapshot.ExecutableProposal`, `string` | :64 | 예 |
| B8 | if | `:76` `if err := validateJudgementSnapshot(id, judgement, candidate); err != nil {` | `fmt.Errorf`, `j.db.BeginTx`, `validateJudgementSnapshot` | :77 | 예 |
| B9 | if | `:80` `if err != nil {` | `fmt.Errorf`, `scanExitProgress`, `tx.Rollback` | :81 | 아니오 |
| B10 | if | `:85` `if err != nil {` | — | :86 | 아니오 |
| B11 | if | `:88` `if current.Completed {` | `fmt.Errorf` | :89 | 예 |
| B12 | if | `:91` `if current.Effective == nil {` | `exitObservationRefreshGuardTx`, `fmt.Errorf` | :92 | 예 |
| B13 | if | `:95` `if err != nil {` | — | :96 | 아니오 |
| B14 | if | `:98` `if status != SnapshotStatusEvaluated \|\| proposalPending {` | `fmt.Errorf` | :99 | 예 |
| B15 | if | `:102` `if expectedLifecycle == 0 {` | — | — | 아니오 |
| B16 | if | `:105` `if expectedLifecycle != current.LifecycleGeneration {` | `Scan`, `fmt.Errorf`, `tx.QueryRowContext` | :106 | 예 |
| B17 | if | `:112` `if errors.Is(err, sql.ErrNoRows) {` | `errors.Is` | — | 예 |
| B18 | else | `:114` `} else if err != nil {` | — | — | 예 |
| B19 | if | `:114` `} else if err != nil {` | — | :115 | 예 |
| B20 | if | `:117` `if lifecycleStatus != positionpolicy.StatusManaged \|\| lifecycleGeneration != expectedLifecycle {` | `compareObservationEvidence`, `fmt.Errorf` | :118 | 예 |
| B21 | if | `:121` `if err != nil {` | — | :122 | 예 |
| B22 | if | `:124` `if decision == observationNoop {` | — | :125 | 예 |
| B23 | if | `:127` `if current.PositionGeneration != request.Snapshot.PositionGeneration \|\|` | `encodeStoredSnapshot`, `fmt.Errorf`, `sameExitOperationalLine` | :129 | 예 |
| B24 | if | `:132` `if err != nil {` | `boolInt`, `j.nowString`, `nullableRung`, `nullableString`, `string`, `tx.ExecContext` | :133 | 아니오 |
| B25 | if | `:148` `if err != nil {` | `fmt.Errorf` | :149 | 아니오 |
| B26 | if | `:151` `if err := j.runExitWriteHook("after_refresh_state"); err != nil {` | `j.runExitWriteHook`, `tx.Commit` | :152, :154 | 예 |

## Calls and live bindings

`compareObservationEvidence` · `sameExitOperationalLine`(B23 조건) · `tx.ExecContext`(UPDATE).

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

`exit_states` 관측 열 — `baseline_price`는 B23을 통과한 스냅샷의 `CurrentProtection`.

## Safety conclusion

- **Safe edit boundary**: **a095는 이 함수를 바꾸지 않는다.** B23이 `sameExitOperationalLine`(보호가 · 워터마크 · 레벨 포함)이 거짓이면 거절하므로, 이 함수가 `baseline_price`에 쓰는 값은 **저장된 effective 스냅샷의 보호가와 같다**(B12가 그 스냅샷의 존재를 요구한다). 비교 대상은 스칼라 `baseline_price`가 아니라 effective JSON이다 — 따라서 「값이 움직이지 않는다」는 **스칼라와 effective 스냅샷이 일치할 때만** 참이고, 둘이 갈라져 있으면 스칼라를 스냅샷의 보호가로 (낮출 수도 있게) 되돌린다(4판, r3 N5). 그 갈라짐에 생산이 도달하는지는 측정하지 않았다.
- **High-risk impact**: yes — 손절선 열을 쓰는 자리다.
