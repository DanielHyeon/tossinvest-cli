# Function Logic Map: `ExitObserver.record`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a074-critical-events-reach-the-operator/base-commit.txt`
- 위험 등급: **High-risk** — 판정을 원장에 영속하고 발의를 브로커로 보내는 경로다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `snapshot` | 평가기가 만든 권위 snapshot | `exitpolicy` | — |
| `snapshot.Orderable && !proposal.Zero()` | orderable 판단 | 평가기 | false면 발의 없음 |
| `quote.FetchedAt` | zero면 사이클 시각 사용 | 관측 | — |
| `RecordExitJudgementResult` | 판정 트랜잭션 | 원장 | error 종류에 따라 분기 |
| `recorded.ArmOutcome` | `ExitArmArmed`만 제출로 이어짐 | 원장 | 그 외는 반환 |

**불변식 1 (유지)**: 순서는 crash 계약이다 — "the judgement transaction arms the
proposal, then the intent is minted, attached and submitted. Reversed … a crash in
between leaves an order the ledger cannot explain."

**불변식 2 (유지)**: `ErrProposalPending`은 조용한 보류다. 다른 발의가 미해소이므로
아무것도 하지 않는 것이 보수적 답이다.

**현재의 결함 (B9/B10)**: `RecordExitJudgementResult`가 `ErrExitSnapshotQuarantined`를
반환하면 — 즉 원장이 **방금 이 포지션을 격리했으면** — 그 error는 `ErrProposalPending`이
아니므로 B9의 일반 경로로 wrap되어 반환된다. 반환된 error는
`ObserveOnce`(`exitloop.go:429`)에서 `cycle.Err`가 되고, `Run`이 사이클을 버린다. 즉
**격리가 만들어진 사이클에는 어떤 기록도 남지 않는다.** 운영 원장의 세 건이 모두 이
경로이며, 알림은 다음 사이클(5초 뒤) `workingSet`의 판정 거부로만 나타난다.

**a074가 바꾸는 것**: B9 안에서 `ErrExitSnapshotQuarantined`를 알아보고, 활성 격리를
되읽어 생성 이벤트를 발행한 뒤 **error를 그대로 반환한다.**

**a074가 바꾸지 않는 것**: 반환 여부와 반환값. 격리는 여전히 사이클 실패이고, 판정은
여전히 기록되지 않으며, 발의는 여전히 제출되지 않는다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (1003) | `quote.FetchedAt.IsZero()` | `judgement.ObservationSource="cycle"` | — | 기존 |
| B2 (1005) | else | `="quote_fetched_at"` | — | 기존 |
| B3 (1025) | orderable && (취소 필요 \|\| 전량청산) | 심볼 정리 | — | 기존 |
| B4 (1027) | 정리 실패 | 없음 | `err` | 기존 |
| B5 (1030) | 정리 미완료 | `noteDelay`, `orderable=false` | — | 기존 |
| B6 (1034) | else | `clearDelay` | — | 기존 |
| B7 (1040) | orderable | intent 생성, `judgement.Proposal` | — | 기존 |
| B8 (1042) | intentID 비었음 | `o.opts.NewID()` | — | 기존 |
| **B9 (1054)** | 판정 기록 실패 | **(신규) 격리면 생성 이벤트 발행** | wrap된 `err` | **3.5** |
| B10 (1055) | `ErrProposalPending` | 없음 | `nil` — 조용한 보류 | 기존 |
| B11 (1063) | 무장되지 않음 | 없음 | `nil` | 기존 |

새 분기는 B9 안의 `errors.Is(err, journal.ErrExitSnapshotQuarantined)` 하나다. B10보다
**뒤에** 온다 — `ErrProposalPending`의 조용한 보류가 우선순위를 잃으면 안 되고, 두
sentinel은 서로 배타적이다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `snapshot.ExecutableProposal` | 발의 추출 | — | AST |
| `o.clearTheSymbol` | 심볼 정리 | 실패는 반환 | AST |
| `o.noteDelay`/`o.clearDelay` | 지연 시계 | latch | AST |
| `Journal.RecordExitJudgementResult` | 판정 트랜잭션 | 격리 시 커밋 **후** `ErrExitSnapshotQuarantined` (`exit_state.go:499-506`) | AST + 소스 |
| `Journal.ActiveExitSnapshotQuarantine` (신규) | 방금 만들어진 격리 행 되읽기 | error면 알림 없이 진행 | 신규 |
| `o.submit` | 브로커 제출 | — | AST |

**되읽기가 정확한 이유** (design D3): `exit_state.go`는 격리를 쓰고 `tx.Commit()`을 한
뒤 error를 반환한다(`503-506`). 격리의 세대는 `recomputed.Line.PositionGeneration`이고,
그 값은 `snapshotContext`가 `m.position.InstanceSeq`로 채운 숫자다(`exitloop.go:898`).
따라서 `ActiveExitSnapshotQuarantine(ctx, m.position.ID, m.position.InstanceSeq)`가
읽는 행은 방금 커밋된 그 행이다.

**되읽기가 실패하면**: 알림 없이 원래 error를 반환한다. 되읽기 실패가 판정 경로의
결과를 바꿔서는 안 된다.

## State mutations and fallbacks

- 원장 write는 `RecordExitJudgementResult` 한 번뿐이고 a074는 그것을 편집하지 않는다.
- 신규 읽기는 **읽기 전용 트랜잭션**이다(`exit_snapshot.go:772`).
- 신규 읽기는 **실패 경로에만** 있다. 정상 판정에는 비용이 0이다.
- `o.quarantineAlerted` latch는 `workingSet`과 공유한다 — 같은 격리를 두 경로가 각각
  한 번씩 알리는 일이 없어야 한다.

## Safety conclusion

- Safe edit boundary: B9의 error 처리 안, `ErrProposalPending` 검사 뒤에 sentinel 하나
  추가 + 알림 헬퍼 호출. 반환문·wrap 문구·B10·B11은 손대지 않는다.
- High-risk impact: **yes** — 다만 편집은 이미 실패한 경로에 관측을 더하는 것이고,
  판정 결과·발의·주문에 아무 값도 공급하지 않는다.
- §0.3: 격리된 포지션은 이번 사이클에 발의를 하지 않는다(판정 자체가 거부됐다).
  새 읽기와 알림은 그 사실이 확정된 **뒤**에 실행되므로 어떤 청산도 지연시키지 않는다.
- §0.6: 손절·익절·사이징 수식을 건드리지 않는다.
