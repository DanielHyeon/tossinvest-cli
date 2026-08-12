# Function Logic Map: `ExitObserver.record`

- Source: `internal/app/engine/exitloop.go` (L1077-1197)
- AST evidence: `ast.json` — 분기 14, return 5
- Risk scan: `risk-pattern-report.md`

## 이 함수가 대상인 이유

design D8-2는 **포지션을 완전히 닫는 청산은 발행 전에 상주 보호주문 취소를 시도해야 한다**고
정했다. 그 판정이 내려지는 곳이 이 함수의 B3(L1117)이다. a087 인프로세스 보호성 매도 경로의
편집 지점이며, **High-risk이므로 면제할 수 없다**(tasks 0.4).

## AST가 정정한 것 — 취소 발동 조건은 이미 넓다

design 2차 개정에서 나는 "`isProtective`에는 익절이 없으니 취소 기준을 `isFullExit`로
넓힌다"고 적었다. AST를 보니 **L1117은 이미 `isFullExit`를 쓰고 있다.**

```go
if orderable && (snapshot.CancelPendingFirst || isFullExit(proposal)) {
```

`isFullExit`는 `ActionBaselineBreach || ActionLadderStop || ActionLadderTakeProfit`이므로
(exitloop.go:1208-1211) **익절도 이미 이 분기에 들어온다.** 넓혀야 할 것은 조건이 아니라
**취소 대상**이다 — 지금 취소하는 것은 브로커의 *working order*이고, 상주 조건주문은
`clearTheSymbol`의 시야 밖에 있다.

## 가장 중요한 발견 — 상주 주문 취소를 `clearTheSymbol` 결과에 섞으면 안 된다

B7(L1145)이 결정적이다.

```go
cleared, err := o.clearTheSymbol(ctx, m, snapshot.CancelPendingFirst)
if err != nil { return err }              // B6 — 매도 자체가 발행되지 않는다
if !cleared {
    o.noteDelay(...)
    orderable = false                      // ← 매도를 withhold 한다
    judgement.ArmSuppressedReason = journal.ArmSuppressedWorkingOrder
}
```

⇒ **`cleared == false`면 청산이 보류된다.** 상주 조건주문 취소를 `clearTheSymbol` 안에 넣거나
그 `cleared` 값에 반영하면, **조건주문 취소 실패가 손절 매도를 막는다.**

그것은 spec이 명시적으로 금지한 것이다 — "취소 확인 실패가 인프로세스 매도를 막거나
지연시켜서는 안 되며 (MUST NOT), typed reason을 기록하고 즉시 매도를 진행해야 한다".
그리고 안전 불변식 §0-4(손절·비상 청산의 즉시성을 약화·지연하지 않는다)에 직접 걸린다.

⇒ **상주 주문 취소는 `cleared`와 무관한 별도 시도여야 하고, 실패는 기록만 하고 진행한다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `snapshot.ExecutableProposal()` | 실행 가능한 제안 | exit policy 평가 | 없으면 이 함수는 제안하지 않는다 |
| `orderable` | bool | 상위 판정 + B4·B7이 끌 수 있다 | false면 arming 없이 판정만 기록 |
| `m.reJudge` | 격리 승계 여부 | 상위 | B4가 익절만 withhold |
| `snapshot.CancelPendingFirst` | bool | 정책 | working order 선취소 요구 |
| `judgement` | 기록할 판정 | 이 함수가 구성 | `RecordExitJudgementResult`가 arming 확정 |

**불변식 1 — 심볼 정리는 arming보다 앞이다**(주석 L1101-1102). 정리하는 대상 중 하나가
arming 자체이기 때문이다. 상주 주문 취소를 arming 뒤로 옮기면 이 순서 근거가 깨진다.

**불변식 2 — 정리 실패는 판정을 멈추지 않는다**(주석 L1113-1116). watermark와 baseline은
계속 전진하고 **제안만 보류**된다. 보류가 침묵하지 않도록 `noteDelay`가 있다.

**불변식 3 — 재판정은 익절만 보류하고 보호는 보류하지 않는다**(주석 L1125-1130).
`isProtective`가 stop을 이 분기에서 빼낸다.

**불변식 4 — arming이 확정되어야 제출한다.** B14(L1190): `ArmedProposal == nil` 또는
`ArmOutcome != ExitArmArmed`면 제출 없이 반환한다. 제출 경로는 오직 L1195 하나다.

## Branches and early returns — 측정 포함

`go test -covermode=set ./internal/app/engine` 기준.

| Branch | 조건 (L) | 결과 | 측정 |
|---|---|---|---|
| B1 | `quote.FetchedAt.IsZero()` (1095) | 관측 시각 대체 | 실행됨 |
| B2 | else (1097) | 관측 시각 사용 | 실행됨 |
| B3 | `orderable && (CancelPendingFirst \|\| isFullExit)` (1117) | **심볼 정리 진입 — a100 편집 지점** | 실행됨 |
| B4 | `m.reJudge && !isProtective` (1118) | 익절 보류, `ArmSuppressedReJudge` | 실행됨 |
| B5 | else (1140) | `clearTheSymbol` 호출 | 실행됨 |
| B6 | `clearTheSymbol` 에러 (1142) | **에러 반환 — 매도 없음** | **미실행** |
| B7 | `!cleared` (1145) | `noteDelay` + 매도 보류 | 실행됨 |
| B8 | else (1149) | `clearDelay` | 실행됨 |
| B9 | `orderable` (1156) | intent 구성 | 실행됨 |
| B10 | `intentID == ""` (1158) | 새 ID 발급 | **미실행** |
| B11 | 판정 기록 에러 (1170) | 에러 분기 | 실행됨 |
| B12 | `ErrProposalPending` (1171) | 조용히 보류 | 실행됨 |
| B13 | `ErrExitSnapshotQuarantined` (1177) | 격리 공지 후 에러 | 실행됨 |
| B14 | arming 미확정 (1190) | 제출 없이 반환 | 실행됨 |

**미실행 2건이 a100에 갖는 뜻.** B6은 `clearTheSymbol`이 **에러를 반환**하는 경로이고 한 번도
실행된 적이 없다. a100이 상주 주문 취소를 이 경로에 얹으면 **한 번도 실행되지 않은 분기에
새 실패 원인을 추가**하는 것이 된다. 그래서 위의 결론(별도 시도, 비차단)이 설계 선택이 아니라
측정된 요구다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `isFullExit` | 완전 청산 판정 | 순수 함수 | AST L1117, exitloop.go:1208 |
| `isProtective` | 보호성 판정 — 재판정 보류에서 stop을 제외 | 순수 함수 | AST L1118, exitloop.go:1217 |
| `o.clearTheSymbol` | **브로커에 실제 취소를 보낸다** | 에러 = 매도 없이 반환(B6); `!cleared` = 매도 보류(B7) | AST L1141, 주석 L1120-1123 |
| `o.noteDelay` / `o.clearDelay` | 지연 관측 시작·해제 | 없음 | AST L1146, L1150 |
| `o.opts.Journal.RecordExitJudgementResult` | 판정 기록 + arming 확정 | `ErrProposalPending`·`ErrExitSnapshotQuarantined` 구별(B12·B13) | AST L1169 |
| `o.announceQuarantineFromLedger` | 격리 공지 | 에러를 바꾸지 않는다 | AST L1186, 주석 L1183-1185 |
| `o.submit` | 매도 발행 | 별도 산출물 | AST L1195 |

**a100이 추가할 호출**: 상주 보호주문 취소. 위 표에서 `clearTheSymbol`과 **같은 칸에 두면 안
된다** — 그 칸의 계약이 「실패가 매도를 막는다」이기 때문이다.

## State mutations and fallbacks

- `orderable`은 **세 곳에서 꺼진다** — 상위 판정, B4(재판정 익절 보류), B7(정리 미완료).
  꺼지면 arming 없이 판정만 기록되고 매도는 다음 주기로 넘어간다.
- `judgement.ArmSuppressedReason`은 보류 이유를 남긴다(`ArmSuppressedReJudge` /
  `ArmSuppressedWorkingOrder`). **상주 주문 취소 실패는 여기에 새 사유를 추가하지 않는다** —
  추가하면 그 자체가 「매도를 보류했다」는 뜻이 되기 때문이다. typed reconcile reason으로
  따로 기록한다.
- watermark와 baseline은 정리 실패와 무관하게 전진한다(주석 L1113-1114). 즉 **보류는 상태를
  얼리지 않는다.**

## Safety conclusion

- Safe edit boundary: B3 블록 **안**에 상주 주문 취소 시도를 추가하되 `cleared`·`orderable`·
  `err` 어느 것도 그 결과로 바꾸지 않는다.
- High-risk impact: **yes.** 손절 발행 경로다.
- **RED 테스트 의무:** 상주 주문 취소가 실패해도 매도가 같은 사이클에 발행되는지 —
  편집 전에 실패하는 테스트로 고정한다. B6에 대한 테스트도 이때 함께 만든다(미실행 분기에
  새 코드를 얹지 않는다).
