# Function Logic Map: `newCandidateCmd`

- Source: `cmd/tossctl/candidate.go`
- AST evidence: `ast.json` (revision=current, L91–98, 분기 0개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `583772c4` — gate 요구 목록 밖의 자발적 evidence (revision=current)
- 비고: gate가 요구한 목록에 없는 **자발적** evidence다 — `cmd/tossctl/candidate.go`는 이 branch range에서 새로 생긴 파일이라 "수정된 기존 함수"에 해당하지 않는다.

발굴 CLI의 부모 명령. `scan`과 `watch` 두 개를 붙이고 그것이 전부다.

이 target은 gate가 요구한 목록에 없다 — `cmd/tossctl/candidate.go`는 이 branch range에서 새로 생긴 파일이므로 "수정된 기존 함수"가 아니다. 발굴 CLI의 안전 논거가 `newRootCmd`의 등록 한 줄에만 걸려 있지 않도록 **자발적으로** 남긴 evidence다.

**`mutating: false`**: 두 서브커맨드 모두. 발굴은 랭킹을 읽고 자기 저장소에 쓴다. 건네받는 인터페이스는 시장을 받아 행을 돌려주는 메서드 하나이고, 주문을 표현할 방법이 없다.

**runlock 게이트**(`watch`): 라이브 계좌 검증이 실행 마커를 신선하게 들고 있으면 `watch`는 **시작하지 않는다**. spec Requirement 7의 순위 engine > verification > discovery가 그 근거다 — 검증 스텝 하나를 429로 잃으면 실주문 하나와 사람의 주의를 잃지만 발굴 사이클은 15초 뒤에 또 있다. 마커 경로 자체를 해석하지 못하면 **fail-open이되 침묵하지 않는다**: stderr에 "우선순위를 강제하지 못했다"고 적고 시작한다(§5 리뷰 P2 ⑥).

**interval 의미**: 운영자의 `--interval`은 **하한**이다. `candidate.WatchInterval`이 3초 floor와 15초 기본값으로 정규화하고, 그다음 `watchWait`가 `Schedule.UntilNextDue`와 비교해 **더 늦은 쪽**을 택한다. 원천별 주기는 schedule이 소유하고 엔진이 켜지면 `engineYieldFactor`가 그것을 두 배로 늘린다. 이 둘을 독립된 두 숫자로 취급하면 tick마다 아무것도 due하지 않은 turn이 생기고, 그 turn이 `ErrNoSourceAnswered`로 루프를 끝냈다 — 루프가 끝나면 아무도 promote하지 않아 암묵 냉각(last_seen_at+10분)과 그 30분 뒤 만료로 **약 40분 안에 저장소의 모든 `first_seen_at`이 사라진다**(tasks 5.9, P0).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root *rootOptions` | non-nil | `newRootCmd` | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (무분기) | — | 서브커맨드 2개 등록 | `*cobra.Command` | `TestTheDiscoveryCommandsDeclareThemselvesReadOnly` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newCandidateScanCmd` | 1회 스캔 — `mutating: false` | 저장소 열기/닫기는 `candidateWiring`이 소유 | `TestTheDiscoveryCommandsCloseTheStoreTheyOpened` |
| `newCandidateWatchCmd` | 반복 루프 — `mutating: false`, `--interval`/`--cycles` | runlock 게이트 + interval 하한 + schedule 우선 | `TestWatchRefusesToStartWhileALiveVerificationHoldsTheRunLock`, `TestWatchStartsWhenTheRunLockIsStale`, `TestWatchSaysSoWhenItCannotTellWhetherAVerificationIsRunning` |
| `cmd.AddCommand` | 트리 구성 | — | ast.json calls |

## State mutations and fallbacks

- 명령 객체만 만든다. 부작용 없음.
- 루프의 실제 안전 규칙은 `runCandidateWatch`와 `internal/candidate`의 `watchWait`/`Schedule.UntilNextDue`가 소유하며, `TestATurnWithNothingDueIsNotAMarketFailure`·`TestTheEngineYieldDoesNotEndTheDiscoveryLoop`·`TestATickBelowTheSourceIntervalDoesNotEndTheDiscoveryLoop`·`TestABackwardClockStepDoesNotEndTheDiscoveryLoop`가 네 도달 경로를 각각 고정한다.

## Safety conclusion

- Safe edit boundary: 서브커맨드 등록 두 줄. annotation을 `mutating: true`로 바꾸는 것은 발굴이 무언가를 낼 수 있다는 선언이 된다.
- High-risk impact: no (읽기 전용 발굴 표면 — 의존 폐포가 주문 경로를 볼 수 없다). 다만 `watch`의 runlock 게이트를 없애는 편집은 라이브 검증의 rate 예산을 발굴이 빼앗게 만들고, 그 손실은 실주문 스텝으로 계산된다.
