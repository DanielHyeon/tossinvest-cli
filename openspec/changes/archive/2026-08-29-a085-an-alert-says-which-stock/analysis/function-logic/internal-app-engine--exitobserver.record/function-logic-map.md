# Function Logic Map: `ExitObserver.record`

- Source: `internal/app/engine/exitloop.go` (lines 1077–1197)
- AST evidence: `ast.json` (`source_sha256: 6625c92061d5b05f566ecb0913f5c5f74a7fdde4cc4b5d8e7dfe8e75dd71de00`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **High-risk**

## What it does

판정을 원장에 기록하고, 무장된 제안을 제출한다. 개정 2·4가 두 가지를 더했다.
(1) `ReJudgingVersion: reJudgingVersion(m)` 을 `ExitJudgement`에 실어 보낸다 — 재판정 사실은
저널이 행에서 추론할 수 없고(각인이 트랜잭션보다 먼저 일어난다) 호출자만 안다. 개정 4가
그 사실을 bool에서 **버전**으로 좁혔다: 각인과 커밋 사이에 운영자 해제와 병행 관측이 행을
교체할 수 있고, 그때 닫혀야 할 행은 재시도를 실제로 소비한 행 하나뿐이다.
(2) 재판정 통과에서 `clearTheSymbol`이 실제 취소를 브로커로 보내기 전에, 보호 제안이 아닌
제안은 보류한다. `clearTheSymbol`은 `RecordExitJudgementResult`의 복구 선택보다 먼저 돌기
때문에 취소-후-거절이면 포지션은 미결 주문도 없고 격리도 안 풀린 채 남는다.
비대칭은 의도적이다: 재판정은 익절을 보류할 뿐 보호는 절대 보류하지 않는다 (§0.3).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `snapshot.Orderable`, `proposal` | 평가 결과 | `exitpolicy` 평가기 | `Zero()`면 무장 없음 |
| `m.reJudge` | bool | `workingSet` → `judge` | false면 upstream 동작과 동일 — 보류 없음 |
| `isProtective(proposal)` | `ActionBaselineBreach` 또는 `ActionLadderStop` | 제안의 action | true면 재판정에서도 보류하지 않는다 — §0.3 |
| `quote.FetchedAt` | zero 허용 | 시세 경로 | zero면 사이클 시각으로 대체 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Required test |
|---|---|---|---|
| B1 | (1095) `if` — if quote.FetchedAt.IsZero() { | 본문 참조 | 아래 Branch Test Map |
| B2 | (1097) `else` — } else { | 본문 참조 | 아래 Branch Test Map |
| B3 | (1117) `if` — if orderable && (snapshot.CancelPendingFirst \|\| isFullExit(proposal)) { | 본문 참조 | 아래 Branch Test Map |
| B4 | (1118) `if` — if m.reJudge && !isProtective(proposal) { | 본문 참조 | 아래 Branch Test Map |
| B5 | (1140) `else` — } else { | 본문 참조 | 아래 Branch Test Map |
| B6 | (1142) `if` — if err != nil { | 본문 참조 | 아래 Branch Test Map |
| B7 | (1145) `if` — if !cleared { | 본문 참조 | 아래 Branch Test Map |
| B8 | (1149) `else` — } else { | 본문 참조 | 아래 Branch Test Map |
| B9 | (1156) `if` — if orderable { | 본문 참조 | 아래 Branch Test Map |
| B10 | (1158) `if` — if intentID == "" { | 본문 참조 | 아래 Branch Test Map |
| B11 | (1170) `if` — if err != nil { | 본문 참조 | 아래 Branch Test Map |
| B12 | (1171) `if` — if errors.Is(err, journal.ErrProposalPending) { | 본문 참조 | 아래 Branch Test Map |
| B13 | (1177) `if` — if errors.Is(err, journal.ErrExitSnapshotQuarantined) { | 본문 참조 | 아래 Branch Test Map |
| B14 | (1190) `if` — if recorded.ArmedProposal == nil \|\| recorded.ArmOutcome != journal.ExitArmArmed { | 본문 참조 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `o.clearTheSymbol` | 미결 매수·매도를 장부에서 뺀다 | 오류를 전파. 실패(=미청산)는 제안만 보류하고 판정은 계속 | AST `calls` L1141 |
| `o.noteDelay` | 지연된 청산을 알림 한도 뒤에 알린다 | 재판정 보류 통로에는 없다 — 그 통로는 손절을 쥐지 않고 `clearDelay`도 없어 타이머가 영원히 남는다 | AST `calls` L1146 |
| `Journal.RecordExitJudgementResult` | 판정 기록 + 복구 선택 + 무장 | `ErrProposalPending`은 보류, `ErrExitSnapshotQuarantined`는 알린 뒤 오류 전파 | AST `calls` L1169 |
| `o.submit` | 무장된 제안 제출 | 오류 전파 | AST `calls` L1195 |

## State mutations and fallbacks

- `judgement.ArmSuppressedReason` — `ArmSuppressedReJudge` 또는 `ArmSuppressedWorkingOrder`. 두 값 모두 `knownArmSuppression` 허용 목록에 있고, 쓰기(`validateJudgementSnapshot`)와 읽기(`validateExitEventArmSuppression`)가 같은 목록을 쓴다.
- `orderable = false` — 제안 보류. 워터마크와 baseline은 그대로 전진한다.
- `cycle.Proposed++`, 그리고 `submit`을 통한 실제 주문 side effect.

## Safety conclusion

- Safe edit boundary: 보류 조건과 `ReJudgingVersion` 전달. 무장·제출 계약은 upstream 그대로다.
- High-risk impact: yes — §0.3. `isProtective`를 빼면 손절이 보류되고, 보류된 제안은 라인이 다시 움직일 때까지 재제안되지 않으므로 무기한 지연된다. `TestAReJudgementNeverWithholdsAStop`이 이 성질을 고정한다.
