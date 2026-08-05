# Function Logic Map: `ExitObserver.judge`

- Source: `internal/app/engine/exitloop.go` (lines 807–840)
- AST evidence: `ast.json` (`source_sha256: 6625c92061d5b05f566ecb0913f5c5f74a7fdde4cc4b5d8e7dfe8e75dd71de00`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **High-risk**

## What it does

한 포지션의 판정 진입점. 개정 2에서 재판정 각인이 `workingSet`에서 이 함수의 머리로
옮겨왔다. 가격을 손에 쥔 뒤에 한 번의 재시도를 쓰기 위해서다 — `workingSet`은 시세 조회 전에
돌기 때문에, 거기서 각인하면 호가를 못 받은 사이클이 그 포지션의 유일한 재판정을 삼키고
행이 각인된 채로 돌아온다. 이후 모든 사이클에서 손절이 평가되지 않은 채 거절된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `m.identityErr` | nil 또는 정책 동일성 오류 | `workingSet`이 저널 상태와 활성 정책을 대조해 채운다 | 비-nil이면 거절 알림 후 nil 반환 — 판정 없음 |
| `m.reJudge` | bool | `workingSet`이 `NeedsReJudgement()`로 결정 | false면 각인 없음. 상위 동작 불변 |
| `m.reJudgeVersion` | 활성 격리 행의 `quarantine_version` | `activeExitSnapshotQuarantineTx` | CAS 실패 시 `ErrExitSnapshotReleaseStale` 전파 → 사이클 오류 |
| `m.state.PolicyKind` | `ladder` 또는 그 외 | `exit_states` 행 | 그 외는 ratchet로 간다 — upstream 기본값 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Required test |
|---|---|---|---|
| B1 | (809) `if` — if m.identityErr != nil { | 본문 참조 | 아래 Branch Test Map |
| B2 | (813) `if` — if m.reJudge { | 본문 참조 | 아래 Branch Test Map |
| B3 | (823) `if` — if err := o.opts.Journal.StampExitSnapshotQuarantineSelector(ctx, | 본문 참조 | 아래 Branch Test Map |
| B4 | (829) `if` — if err != nil { | 본문 참조 | 아래 Branch Test Map |
| B5 | (834) `switch` — switch m.state.PolicyKind { | 본문 참조 | 아래 Branch Test Map |
| B6 | (835) `case` — case journal.ExitPolicyLadder: | 본문 참조 | 아래 Branch Test Map |
| B7 | (837) `case` — default: | 본문 참조 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Journal.StampExitSnapshotQuarantineSelector` | 재판정 한 번을 소비 기록 | CAS UPDATE. `RowsAffected != 1`이면 stale 오류를 그대로 올린다 — fail-closed | AST `calls` L823 |
| `o.alertRefused` | 동일성·break-even 거절을 운영자에게 | 발송 실패는 판정을 막지 않는다 | AST `calls` L810·L830 |
| `o.breakEven` | 비용 포함 손익분기 매도가 | 오류면 거절 후 nil — 주문 없음 | AST `calls` L828 |
| `o.judgeLadder` / `o.judgeRatchet` | 정책별 평가 | 오류를 그대로 전파 | AST `calls` L836·L838 |

## State mutations and fallbacks

- `exit_snapshot_quarantines.selector_revision` — 재판정 1회 소비의 유일한 기록. 각인 실패는 격리를 그대로 두고 `NeedsReJudgement`가 계속 true라 다음 사이클이 재시도한다 (fail-closed·self-healing).
- 그 외 상태 변경 없음. 주문·손절선·사이징은 이 함수가 만지지 않는다.

## Safety conclusion

- Safe edit boundary: 각인 시점과 정책 분기. 평가·제안·주문은 하위 함수의 몫이다.
- High-risk impact: yes — §0.3. 각인을 가격 획득 *전*으로 되돌리면 호가 없는 사이클이 재시도를 태우고 손절이 무기한 평가되지 않는다. 그것이 개정 2가 고친 결함이다.
