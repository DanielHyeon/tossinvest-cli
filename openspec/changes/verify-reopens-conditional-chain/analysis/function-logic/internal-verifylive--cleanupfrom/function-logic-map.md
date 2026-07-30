# Function Logic Map: `cleanupFrom`

- Source: `internal/verifylive/cleanup.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- High-risk: **yes** — 이 함수의 결과가 라이브 취소 요청의 대상 목록이 된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `entries []Entry` | 증거 기록의 append-only 순서. nil 허용 | `LoadEntries`(파일) 또는 `r.prior` | nil이면 `Outstanding`이 빈 목록 → 대상 없음 |
| `settled func(StepID) bool` | `r.settled`(redo 반영) 또는 `Settled`(순수 판정) | `runner.go settled` / `record.go Settled` | redo 요청 중이면 false → 가드 닫힘(보수) |
| `Outstanding(entries)` | 취소되지 않은 artifact만 | `record.go Outstanding` — 식별자별 **마지막 언급**이 이긴다 | 취소가 monotone이라 되살아나지 않는다 |
| artifact 생성 위치 | `entries` 안에 반드시 존재 | 같은 기록 | 못 찾으면 `decidedAfter`가 false → 정리 안 함 |

불변식: 이 함수는 **기록이 "이 도구가 만들었다"고 말하는 객체만** 돌려준다. 계좌의 다른
주문은 `Outstanding`에 없으므로 구조적으로 대상이 될 수 없다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range Outstanding(entries)` (line 123) | 없음(순수) | 루프 종료 시 `out` | `TestNoLeftoversMeansNoCleanupLines` |
| B2 | `switch a.Kind` (124) | 없음 | — | — |
| B3 | `case KindOrder` (125) | `out` append | 주문은 **항상** 대상 | `TestALeftoverOrderIsCancelledOnTheNextRun`, `TestALeftoverOrderIsNotSubjectToTheOrderingRule` |
| B4 | `case KindConditional` (127) | 없음 | B5로 | — |
| B5 | `settled(StepConditionalCancel) && decidedAfter(entries, StepConditionalCancel, a)` (128) | true면 `out` append | 조건주문은 **두 조건 모두** 참일 때만 대상 | 아래 3분기 |
| B5-a | `settled` false (판정 없음 / redo 중) | 없음 | 대상 아님 | `TestTheConditionalLeftForPersistenceIsNotCleanedUp`, `TestAConditionalWithNoCancelVerdictIsNotCleanedUp` |
| B5-b | `settled` true, `decidedAfter` **false** — 판정이 객체보다 오래됐다 | 없음 | 대상 아님 — **이 change가 추가한 분기** | `TestAVerdictOlderThanTheConditionalDoesNotCleanItUp` |
| B5-c | 둘 다 true — 등록 이후에 취소가 실패했다 | `out` append | 대상 | `TestAVerdictNewerThanTheConditionalOpensCleanup` |

early return 없음. `return out` 하나(line 133). `switch`에 default 없음 — `record.go`가
artifact kind를 `order`/`conditional-order` 둘로 고정하며, 그 밖의 kind는 조용히 대상에서
빠진다(정리하지 않는 쪽 = 보수).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Outstanding` (123) | 취소되지 않은 자기 산출물 목록 | 순수 함수, error 없음. 취소는 monotone | `record.go`, CodeGraph |
| `settled` (128) | 취소 단계가 terminal인지 + redo 반영 | 순수, error 없음 | `runner.go settled`, `record.go Settled` |
| `decidedAfter` (128) | 판정이 그 객체보다 **나중**인지 | 순수, error 없음, fail-closed | 이 change 신규 leaf |
| `append` (126, 129) | 결과 누적 | — | — |

라이브 바인딩 없음 — 이 함수는 브로커를 호출하지 않는다. 결과가 `planCleanup`을 거쳐 승인
목록이 되고, `runCleanup`이 `r.gate`를 지나 실제 취소를 보낸다. 호출자는 `cleanupTargets`
(cleanup.go:89)와 `PendingCleanup`(cleanup.go:105) 둘뿐이다(CodeGraph callers).

## State mutations and fallbacks

- 지역 slice `out`만 변경한다. `entries`와 artifact는 변형하지 않는다.
- fallback: 판단 근거가 없으면(생성 항목 미발견 / 취소 단계 항목 없음) **정리하지 않는다**.
  보수적인 방향이 "측정 대상을 남긴다"이기 때문이다.
- 이 change 이전 동작과의 차이는 **한 방향뿐**이다: 조건주문이 대상이 되는 경우가 줄고,
  늘어나지 않는다. 주문 경로(B3)는 무변경.

## Safety conclusion

- Safe edit boundary: B5의 조건 추가. B1~B4와 주문 경로는 손대지 않았다.
- High-risk impact: **yes**(라이브 취소 대상 선정). 방향은 §0.9의 보수 쪽 — 취소를 **덜**
  보낸다. §0.3(손절 즉시성)과 무관: 여기의 조건주문은 검증 도구가 등록한 1주짜리 측정용
  손절이며 엔진의 보호 주문이 아니다.
- 회귀 위험: 조건주문 잔여물이 영원히 남는 것. B5-c 테스트가 그것을 막는다.
