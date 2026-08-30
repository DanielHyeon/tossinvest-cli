# Function Logic Map: `RedoSet`

- Source: `internal/verifylive/redo.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- High-risk: **yes** — 이 집합에 들어간 단계는 라이브 요청을 보낼 수 있다(승인을 거친 뒤).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `entries []Entry` | 증거 기록. nil 허용 | `LoadEntries` | nil이면 모든 `LastEntry`가 miss → 빈 집합 |
| `Steps()` | catalogue 순서 고정 | `verifylive.go` | — |
| `LastEntry(entries, id)` | 그 단계의 **최신** 항목 | `record.go` | 없으면 continue(시도된 적 없음 = redo 아님) |
| 시장 분리 | 기록 파일 1개 = 시장 1개 | `record.go:47` 주석 | US pass가 KR 단계를 settle하지 못한다 |

**불변식 (변경됨)**: 이전에는 "집합은 판정만의 함수"였다. 이제는 "판정 + 그 단계가 남긴
artifact의 현재 상태 + 의존 단계들의 판정"의 함수다. `RedoableVerdict`는 여전히 판정 절반만
답하므로, 화면이 집합과 같은 답을 하려면 집합 자체를 물어야 한다(→ `internal/console`
`pageFuncs.inRedo`).

**불변식 (유지)**: 집합은 **제안이지 인가가 아니다**. `runner.go approveBatch`는 이 파일을
알지 못하며, 모든 요청은 새 nonce의 새 계획·새 배치 승인을 다시 지난다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range Steps()` (line 76) | 없음 | catalogue 순서 보장 | `TestRedoSetTakesFailedAndSkippedStepsOnly` |
| B2 | `!ok` — 그 단계 항목이 없다 (78) | 없음 | continue. 시도된 적 없는 단계는 redo가 아니라 일반 이어하기 | `TestRedoSetOnAnEmptyRecordIsEmpty`, `TestRedoSetIgnoresTheApprovalLine` |
| B3 | `RedoableVerdict(e.Verdict) \|\| subjectLost(entries, step, e)` (82) | `out` append | 집합 포함 | 아래 분기 |
| B3-a | `RedoableVerdict` true (`fail`·`skipped`) | append | 기존 동작 | `TestRedoSetTakesFailedAndSkippedStepsOnly`, `TestRedoSetReadsTheNewestVerdict` |
| B3-b | `subjectLost` true — **이 change가 추가한 분기** | append | 통과했지만 대상이 사라진 전제 단계 | `TestRedoSetReopensARegisterWhoseConditionalIsGone` |
| B3-c | 둘 다 false | 없음 | 집합 밖 | `TestRedoSetNeverOffersAPassedStep`, `TestRedoSetLeavesACompletedChainClosed`, `TestRedoSetDoesNotReopenWhileTheConditionalIsAlive`, `TestRedoSetDoesNotReopenForADeferredDependentAlone`, `TestRedoSetDoesNotReopenAPassThatLeftNothingBehind` |

early return 없음. `return out` 하나(line 86).

`subjectLost`(신규 leaf)의 세 조건은 모두 AND이며 각각 하나의 테스트가 고정한다:

| 조건 | 거짓일 때 | 고정 테스트 |
|---|---|---|
| 최신 판정 `pass` + `Deliberate` 조건주문을 남겼다 | false | `TestRedoSetDoesNotReopenAPassThatLeftNothingBehind`, `TestRedoSetNeverOffersAPassedStep` |
| 살아 있는 조건주문이 하나도 없다 | false | `TestRedoSetDoesNotReopenWhileTheConditionalIsAlive` |
| 비-`Deferred` 의존 단계 중 미통과가 있다 | false | `TestRedoSetLeavesACompletedChainClosed`, `TestRedoSetDoesNotReopenForADeferredDependentAlone` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Steps()` (76) | catalogue 순서 | 순수, error 없음 | `verifylive.go` |
| `LastEntry` (77) | 그 단계의 최신 판정 | 순수, error 없음 | `record.go` |
| `RedoableVerdict` (82) | 판정 절반 | 순수, error 없음 | `redo.go:76` |
| `subjectLost` (82) | 대상이 사라진 전제 단계인지 | 순수, error 없음. 내부에서 `Outstanding`·`Passed`·`Steps` 호출 | 이 change 신규 leaf |
| `append` (83) | 결과 누적 | — | — |

라이브 바인딩 없음. 소비자(CodeGraph impact): `internal/console/data.go` `readVerify`(310)와
`redoSet`(337), `internal/console/pages.go` `handleStart`(202),
`cmd/tossctl/verify.go`의 `--redo`는 이 함수를 거치지 않고 사용자 플래그를 직접 쓴다.

## State mutations and fallbacks

- 지역 slice `out`만 변경한다. `entries`를 변형하지 않는다.
- fallback 없음 — 근거가 없으면 집합에 넣지 않는다(아무것도 보내지 않는 쪽).
- 새 분기는 집합을 **넓힐 수만** 있다. 넓어지는 조건이 세 개 AND이고 그중 두 개가
  "아직 아무도 이 대상을 쓰지 못했다"를 요구하므로, 정상 종료한 체인은 넓어지지 않는다.

## Safety conclusion

- Safe edit boundary: B3의 `||` 우변 추가. B1·B2와 판정 목록(`RedoableVerdict`)은 무변경.
- High-risk impact: **yes**. 이 분기는 `conditional-register` 재실행 → **조건주문 등록**이라는
  라이브 side effect로 이어질 수 있다. 완화:
  - 등록되는 것은 1주짜리 SINGLE MARKET SELL 손절이고 발동가가 시장보다 한참 아래다 —
    보호를 더하는 방향(§0.9 준수), 체결 의도 없음.
  - 집합은 인가가 아니다. `Plan.Authorises`가 목록 밖 요청을 계속 거절하고, 사람이 새
    배치 승인을 준다(§0.1·§0.7 충족).
  - 실측 US 기록으로 좁음을 확인했다: 모든 의존 단계 pass + trigger만 deferred →
    재개통되지 않는다.
- 회귀 위험: 규칙이 넓어져 통과한 단계가 재실행되는 것. `TestRedoSetNeverOffersAPassedStep`
  (기존)과 위 5건이 그것을 막는다.
